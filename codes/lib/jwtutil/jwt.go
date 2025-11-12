package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/venue-master/platform/lib/config"
)

// Claims encodes user metadata into JWTs.
type Claims struct {
	UserID      string   `json:"sub"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWT/refresh tokens.
type Manager struct {
	cfg config.JWTConfig
}

// NewManager constructs a Manager from shared config.
func NewManager(cfg config.JWTConfig) *Manager {
	return &Manager{cfg: cfg}
}

// Generate issues signed access + refresh tokens.
func (m *Manager) Generate(userID string, roles, permissions []string) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	claims := Claims{
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			Audience:  []string{m.cfg.Audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshClaims := claims
	refreshClaims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(m.cfg.RefreshExpiry))

	accessToken, err = m.sign(claims)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.sign(refreshClaims)
	return accessToken, refreshToken, err
}

// Validate parses a token string and returns claims if valid.
func (m *Manager) Validate(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.cfg.Secret), nil
	}, jwt.WithAudience(m.cfg.Audience), jwt.WithIssuer(m.cfg.Issuer))
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// Refresh consumes a refresh token and returns new access+refresh tokens.
func (m *Manager) Refresh(refreshToken string) (string, string, error) {
	claims, err := m.Validate(refreshToken)
	if err != nil {
		return "", "", err
	}
	return m.Generate(claims.UserID, claims.Roles, claims.Permissions)
}

func (m *Manager) sign(claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.Secret))
}

// AccessTTL returns the configured access token TTL.
func (m *Manager) AccessTTL() time.Duration {
	return m.cfg.AccessExpiry
}

// RefreshTTL returns the configured refresh token TTL.
func (m *Manager) RefreshTTL() time.Duration {
	return m.cfg.RefreshExpiry
}
