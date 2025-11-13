package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/venue-master/platform/lib/config"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Store manages booking persistence.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a booking store using shared config.
func New(ctx context.Context, cfg config.DatabaseConfig) (*Store, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close closes the underlying pool.
func (s *Store) Close() { s.pool.Close() }

// RunMigrations executes embedded SQL migrations.
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
		sqlBytes, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", entry.Name(), err)
		}
	}
	return nil
}

// Facility represents a bookable resource.
type Facility struct {
	ID               uuid.UUID
	VenueID          uuid.UUID
	Name             string
	Description      string
	Surface          string
	OpenAt           time.Time
	CloseAt          time.Time
	Available        bool
	WeekdayRateCents int
	WeekendRateCents int
	Currency         string
}

// Booking aggregates booking data plus facility linkage.
type Booking struct {
	ID            uuid.UUID
	FacilityID    uuid.UUID
	UserID        uuid.UUID
	StartsAt      time.Time
	EndsAt        time.Time
	Status        string
	AmountCents   int
	Currency      string
	PaymentIntent string
	Facility      *Facility
}

// PaymentRetry tracks pending payment retries.
type PaymentRetry struct {
	BookingID     uuid.UUID
	Attempt       int
	NextAttemptAt time.Time
	LastError     string
}

// SeedFacility ensures there is at least one facility to book.
func (s *Store) SeedFacility(ctx context.Context, f Facility) error {
	_, err := s.pool.Exec(ctx, `
        INSERT INTO facilities (id, venue_id, name, description, surface, open_at, close_at, available)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (id) DO NOTHING
    `, f.ID, f.VenueID, f.Name, f.Description, f.Surface, f.OpenAt, f.CloseAt, f.Available)
	return err
}

// GetFacility returns a facility by ID.
func (s *Store) GetFacility(ctx context.Context, id uuid.UUID) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
        FROM facilities WHERE id = $1
    `, id)
	var f Facility
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// ListFacilities fetches facilities with optional availability filter.
func (s *Store) ListFacilities(ctx context.Context, venueID uuid.UUID, onlyAvailable *bool, limit, offset int) ([]Facility, error) {
    if limit <= 0 {
        limit = 20
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }
    query := `SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency FROM facilities WHERE 1=1`
    args := []any{}
    idx := 1
    if venueID != uuid.Nil {
        query += fmt.Sprintf(" AND venue_id = $%d", idx)
        args = append(args, venueID)
        idx++
    }
    if onlyAvailable != nil {
        query += fmt.Sprintf(" AND available = $%d", idx)
        args = append(args, *onlyAvailable)
        idx++
    }
    query += fmt.Sprintf(" ORDER BY name ASC LIMIT %d OFFSET %d", limit, offset)
    rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facilities []Facility
	for rows.Next() {
		var f Facility
		if err := rows.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
			return nil, err
		}
		facilities = append(facilities, f)
	}
	return facilities, rows.Err()
}

// UpdateFacilityAvailability toggles facility availability.
func (s *Store) UpdateFacilityAvailability(ctx context.Context, id uuid.UUID, available bool) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE facilities SET available=$2, updated_at=NOW()
        WHERE id=$1 RETURNING id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
    `, id, available)
	var f Facility
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// CreateFacility inserts a new facility row.
func (s *Store) CreateFacility(ctx context.Context, f Facility) (*Facility, error) {
	row := s.pool.QueryRow(ctx, `
        INSERT INTO facilities (id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
        RETURNING id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency
    `, f.ID, f.VenueID, f.Name, f.Description, f.Surface, f.OpenAt, f.CloseAt, f.Available, f.WeekdayRateCents, f.WeekendRateCents, f.Currency)
	if err := row.Scan(&f.ID, &f.VenueID, &f.Name, &f.Description, &f.Surface, &f.OpenAt, &f.CloseAt, &f.Available, &f.WeekdayRateCents, &f.WeekendRateCents, &f.Currency); err != nil {
		return nil, err
	}
	return &f, nil
}

// ListBookings returns bookings for a user (optional) with facility data.
func (s *Store) ListBookings(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Booking, error) {
    if limit <= 0 {
        limit = 20
    }
    if limit > 100 {
        limit = 100
    }
    if offset < 0 {
        offset = 0
    }
    query := `
        SELECT b.id, b.facility_id, b.user_id, b.starts_at, b.ends_at, b.status, b.amount_cents, b.currency, b.payment_intent,
               f.id, f.venue_id, f.name, f.description, f.surface, f.open_at, f.close_at, f.available, f.weekday_rate_cents, f.weekend_rate_cents, f.currency
        FROM bookings b
        JOIN facilities f ON f.id = b.facility_id
    `
    args := []any{}
    if userID != uuid.Nil {
        query += " WHERE b.user_id = $1"
        args = append(args, userID)
    }
    query += fmt.Sprintf(" ORDER BY b.starts_at DESC LIMIT %d OFFSET %d", limit, offset)

    rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		var facility Facility
		if err := rows.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent,
			&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
			return nil, err
		}
		b.Facility = &facility
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

// CreateBookingInput carries booking parameters.
type CreateBookingInput struct {
	FacilityID  uuid.UUID
	UserID      uuid.UUID
	StartsAt    time.Time
	EndsAt      time.Time
	AmountCents int
	Currency    string
}

// CreateBooking inserts a booking row if no conflict exists.
func (s *Store) CreateBooking(ctx context.Context, input CreateBookingInput) (*Booking, error) {
	conflict, err := s.hasConflict(ctx, input.FacilityID, input.StartsAt, input.EndsAt)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, errors.New("facility already booked for that time range")
	}

	bookingID := uuid.New()
	row := s.pool.QueryRow(ctx, `
        INSERT INTO bookings (id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency)
        VALUES ($1,$2,$3,$4,$5,'PENDING_PAYMENT',$6,$7)
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, bookingID, input.FacilityID, input.UserID, input.StartsAt, input.EndsAt, input.AmountCents, input.Currency)

	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// AttachFacility hydrates booking with facility details.
func (s *Store) AttachFacility(ctx context.Context, booking *Booking) error {
	if booking == nil {
		return nil
	}
	row := s.pool.QueryRow(ctx, `SELECT id, venue_id, name, description, surface, open_at, close_at, available, weekday_rate_cents, weekend_rate_cents, currency FROM facilities WHERE id=$1`, booking.FacilityID)
	var facility Facility
	if err := row.Scan(&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
		return err
	}
	booking.Facility = &facility
	return nil
}

// UpdateBookingStatus sets status/payment info.
func (s *Store) UpdateBookingStatus(ctx context.Context, id uuid.UUID, status, paymentIntent string) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE bookings SET status=$2, payment_intent=$3, updated_at=NOW()
        WHERE id=$1
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, id, status, paymentIntent)
	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// CancelBooking marks a booking cancelled.
func (s *Store) CancelBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        UPDATE bookings SET status='CANCELLED', updated_at=NOW()
        WHERE id=$1
        RETURNING id, facility_id, user_id, starts_at, ends_at, status, amount_cents, currency, payment_intent
    `, id)
	var b Booking
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBooking fetches a single booking by id.
func (s *Store) GetBooking(ctx context.Context, id uuid.UUID) (*Booking, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT b.id, b.facility_id, b.user_id, b.starts_at, b.ends_at, b.status, b.amount_cents, b.currency, b.payment_intent,
               f.id, f.venue_id, f.name, f.description, f.surface, f.open_at, f.close_at, f.available, f.weekday_rate_cents, f.weekend_rate_cents, f.currency
        FROM bookings b
        JOIN facilities f ON f.id = b.facility_id
        WHERE b.id = $1
    `, id)
	var b Booking
	var facility Facility
	if err := row.Scan(&b.ID, &b.FacilityID, &b.UserID, &b.StartsAt, &b.EndsAt, &b.Status, &b.AmountCents, &b.Currency, &b.PaymentIntent,
		&facility.ID, &facility.VenueID, &facility.Name, &facility.Description, &facility.Surface, &facility.OpenAt, &facility.CloseAt, &facility.Available, &facility.WeekdayRateCents, &facility.WeekendRateCents, &facility.Currency); err != nil {
		return nil, err
	}
	b.Facility = &facility
	return &b, nil
}

func (s *Store) hasConflict(ctx context.Context, facilityID uuid.UUID, start, end time.Time) (bool, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM bookings
            WHERE facility_id=$1 AND status IN ('PENDING_PAYMENT','CONFIRMED')
              AND starts_at < $3 AND ends_at > $2
        )
    `, facilityID, start, end)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// SchedulePaymentRetry inserts/updates a pending retry.
func (s *Store) SchedulePaymentRetry(ctx context.Context, bookingID uuid.UUID, next time.Time, attempt int, lastErr string) error {
	_, err := s.pool.Exec(ctx, `
        INSERT INTO payment_retries (booking_id, attempt, next_attempt_at, last_error)
        VALUES ($1,$2,$3,$4)
        ON CONFLICT (booking_id) DO UPDATE SET attempt=$2, next_attempt_at=$3, last_error=$4, updated_at=NOW()
    `, bookingID, attempt, next, lastErr)
	return err
}

// FetchDuePaymentRetries returns retries due for processing.
func (s *Store) FetchDuePaymentRetries(ctx context.Context, limit int) ([]PaymentRetry, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
        SELECT booking_id, attempt, next_attempt_at, last_error
        FROM payment_retries
        WHERE next_attempt_at <= NOW()
        ORDER BY next_attempt_at ASC
        LIMIT $1
    `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retries []PaymentRetry
	for rows.Next() {
		var pr PaymentRetry
		if err := rows.Scan(&pr.BookingID, &pr.Attempt, &pr.NextAttemptAt, &pr.LastError); err != nil {
			return nil, err
		}
		retries = append(retries, pr)
	}
	return retries, rows.Err()
}

// DeletePaymentRetry removes a retry row.
func (s *Store) DeletePaymentRetry(ctx context.Context, bookingID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM payment_retries WHERE booking_id = $1`, bookingID)
	return err
}
