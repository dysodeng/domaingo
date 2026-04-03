package tokenstore

import (
	"context"
	"time"

	tilesCache "github.com/CXeon/tiles/cache"
	authdomain "github.com/CXeon/domaingo/internal/domain/auth"
	"github.com/google/uuid"
)

const (
	refreshPrefix = "ts:refresh:"
	blockPrefix   = "ts:block:"
)

type redisStore struct {
	cache tilesCache.Cache
}

func NewRedis(c tilesCache.Cache) authdomain.TokenStore {
	return &redisStore{cache: c}
}

func (s *redisStore) SaveRefreshToken(ctx context.Context, jti string, uid uuid.UUID, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		return nil
	}
	return s.cache.Set(ctx, refreshPrefix+jti, uid.String(), ttl)
}

func (s *redisStore) DeleteRefreshToken(ctx context.Context, jti string) error {
	return s.cache.Delete(ctx, refreshPrefix+jti)
}

func (s *redisStore) RefreshTokenUID(ctx context.Context, jti string) (uuid.UUID, bool) {
	val, err := s.cache.Get(ctx, refreshPrefix+jti)
	if err != nil {
		return uuid.Nil, false
	}
	uid, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, false
	}
	return uid, true
}

func (s *redisStore) BlockAccessToken(ctx context.Context, jti string, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		return nil
	}
	return s.cache.Set(ctx, blockPrefix+jti, "1", ttl)
}

func (s *redisStore) IsAccessTokenBlocked(ctx context.Context, jti string) bool {
	ok, err := s.cache.Exists(ctx, blockPrefix+jti)
	if err != nil {
		return false
	}
	return ok
}
