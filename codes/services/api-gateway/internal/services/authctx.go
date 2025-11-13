package services

import "context"

// authContextKey is used to store auth metadata in context.
type authContextKey struct{}

// AuthMetadata carries JWT-derived claims for downstream calls.
type AuthMetadata struct {
	UserID string
	Roles  []string
}

// WithAuth injects auth metadata into a context.
func WithAuth(ctx context.Context, meta AuthMetadata) context.Context {
	return context.WithValue(ctx, authContextKey{}, meta)
}

// AuthFromContext retrieves auth metadata.
func AuthFromContext(ctx context.Context) (AuthMetadata, bool) {
	meta, ok := ctx.Value(authContextKey{}).(AuthMetadata)
	return meta, ok
}
