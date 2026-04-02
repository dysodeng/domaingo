package tokenstore

import (
	"context"
	"sync"
	"time"

	authdomain "github.com/CXeon/domaingo/internal/domain/auth"
	"github.com/google/uuid"
)

type refreshEntry struct {
	uid    uuid.UUID
	expiry time.Time
}

type memoryStore struct {
	mu              sync.RWMutex
	refreshTokens   map[string]refreshEntry
	accessBlocklist map[string]time.Time
}

func New() authdomain.TokenStore {
	return &memoryStore{
		refreshTokens:   make(map[string]refreshEntry),
		accessBlocklist: make(map[string]time.Time),
	}
}

func (s *memoryStore) SaveRefreshToken(_ context.Context, jti string, uid uuid.UUID, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[jti] = refreshEntry{uid: uid, expiry: expiry}
	return nil
}

func (s *memoryStore) DeleteRefreshToken(_ context.Context, jti string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.refreshTokens, jti)
	return nil
}

func (s *memoryStore) RefreshTokenUID(_ context.Context, jti string) (uuid.UUID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.refreshTokens[jti]
	if !ok || time.Now().After(entry.expiry) {
		return uuid.Nil, false
	}
	return entry.uid, true
}

func (s *memoryStore) BlockAccessToken(_ context.Context, jti string, expiry time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessBlocklist[jti] = expiry
	return nil
}

func (s *memoryStore) IsAccessTokenBlocked(_ context.Context, jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, ok := s.accessBlocklist[jti]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}
