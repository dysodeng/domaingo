package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TokenStore interface {
	SaveRefreshToken(ctx context.Context, jti string, uid uuid.UUID, expiry time.Time) error
	DeleteRefreshToken(ctx context.Context, jti string) error
	// RefreshTokenUID returns the uid associated with jti, and false if not found or expired.
	RefreshTokenUID(ctx context.Context, jti string) (uuid.UUID, bool)
	BlockAccessToken(ctx context.Context, jti string, expiry time.Time) error
	IsAccessTokenBlocked(ctx context.Context, jti string) bool
}
