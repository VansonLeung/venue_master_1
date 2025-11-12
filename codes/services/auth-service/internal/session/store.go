package session

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/venue-master/platform/lib/config"
)

// Store persists refresh tokens to Redis so they can be revoked.
type Store struct {
	client *redis.Client
	ttl    time.Duration
}

// NewStore connects to Redis using shared config.
func NewStore(cfg config.RedisConfig, ttl time.Duration) (*Store, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Store{client: client, ttl: ttl}, nil
}

// Save stores a refresh token for a user.
func (s *Store) Save(ctx context.Context, userID, refreshToken string) error {
	return s.client.Set(ctx, s.key(userID, refreshToken), "active", s.ttl).Err()
}

// Exists checks whether the refresh token is still active.
func (s *Store) Exists(ctx context.Context, userID, refreshToken string) (bool, error) {
	count, err := s.client.Exists(ctx, s.key(userID, refreshToken)).Result()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// Delete removes a refresh token (used when rotating tokens).
func (s *Store) Delete(ctx context.Context, userID, refreshToken string) error {
	return s.client.Del(ctx, s.key(userID, refreshToken)).Err()
}

func (s *Store) key(userID, token string) string {
	return fmt.Sprintf("session:%s:%s", userID, token)
}
