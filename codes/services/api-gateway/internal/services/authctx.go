package services

import "context"

type authContextKey struct{}

// AuthMetadata carries the JWT-derived info to downstream services.
type AuthMetadata struct {
	UserID string
	Roles  []string
}

func WithAuth(ctx context.Context, meta AuthMetadata) context.Context {
	return context.WithValue(ctx, authContextKey{}, meta)
}

func AuthFromContext(ctx context.Context) (AuthMetadata, bool) {
	meta, ok := ctx.Value(authContextKey{}).(AuthMetadata)
	return meta, ok
}
