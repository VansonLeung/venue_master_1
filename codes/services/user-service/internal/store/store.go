package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/venue-master/platform/lib/config"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Store wraps PostgreSQL access for the user service.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a Store using the shared database config.
func New(ctx context.Context, cfg config.DatabaseConfig) (*Store, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close releases database resources.
func (s *Store) Close() {
	s.pool.Close()
}

// RunMigrations executes embedded SQL migrations in lexical order.
func (s *Store) RunMigrations(ctx context.Context) error {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		contents, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, string(contents)); err != nil {
			return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
		}
	}
	return nil
}

// Ping verifies the database connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// User represents a stored user row.
type User struct {
	ID           uuid.UUID
	Email        string
	FirstName    string
	LastName     string
	Roles        []string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetUserByID fetches a user by UUID.
func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT id, email, first_name, last_name, password_hash, roles, created_at, updated_at
        FROM users
        WHERE id = $1
    `, id)
	return scanUser(row)
}

// GetUserByEmail fetches a user by email address.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT id, email, first_name, last_name, password_hash, roles, created_at, updated_at
        FROM users
        WHERE LOWER(email) = LOWER($1)
    `, email)
	return scanUser(row)
}

// UpsertUser inserts or updates a user row.
func (s *Store) UpsertUser(ctx context.Context, user *User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	_, err := s.pool.Exec(ctx, `
        INSERT INTO users (id, email, first_name, last_name, password_hash, roles, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            email = EXCLUDED.email,
            first_name = EXCLUDED.first_name,
            last_name = EXCLUDED.last_name,
            password_hash = EXCLUDED.password_hash,
            roles = EXCLUDED.roles,
            updated_at = NOW()
    `, user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash, user.Roles)
	return err
}

// SeedDefaultUser ensures an initial member exists for local development.
func (s *Store) SeedDefaultUser(ctx context.Context, email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password required for seeding")
	}
	user, err := s.GetUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user = &User{
		ID:           uuid.New(),
		Email:        strings.ToLower(email),
		FirstName:    "Venue",
		LastName:     "Member",
		Roles:        []string{"MEMBER"},
		PasswordHash: hash,
	}
	if err := s.UpsertUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// HashPassword hashes the provided password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword checks a plaintext password against the stored hash.
func ComparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func scanUser(row pgx.Row) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.PasswordHash, &u.Roles, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func buildDSN(cfg config.DatabaseConfig) (string, error) {
	host := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	userInfo := url.UserPassword(cfg.User, cfg.Password)
	dsn := url.URL{
		Scheme: "postgres",
		User:   userInfo,
		Host:   host,
		Path:   "/" + cfg.Name,
	}
	query := dsn.Query()
	query.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = query.Encode()
	return dsn.String(), nil
}
