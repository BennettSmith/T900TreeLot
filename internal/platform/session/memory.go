package session

import (
	"context"
	"encoding/hex"
	"sync"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
)

// MemoryStore is an in-process session store for focused HTTP tests.
type MemoryStore struct {
	mu         sync.Mutex
	clock      clock.Clock
	ttl        time.Duration
	sessionKey []byte
	nextID     int64
	byHash     map[string]memorySession
}

type memorySession struct {
	Session
	revoked bool
}

// NewMemoryStore constructs an in-memory session store. sessionKey is the HMAC
// key used to hash opaque cookie tokens, matching the PostgreSQL store.
func NewMemoryStore(clk clock.Clock, ttl time.Duration, sessionKey []byte) *MemoryStore {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &MemoryStore{
		clock:      clk,
		ttl:        ttl,
		sessionKey: append([]byte(nil), sessionKey...),
		byHash:     make(map[string]memorySession),
	}
}

// Create inserts a new in-memory session.
func (s *MemoryStore) Create(ctx context.Context) (Session, string, error) {
	_ = ctx
	return s.create("", nil)
}

// CreateForIdentity inserts a new authenticated in-memory session.
func (s *MemoryStore) CreateForIdentity(ctx context.Context, identityID string) (Session, string, error) {
	_ = ctx
	if identityID == "" {
		return Session{}, "", ErrNotFound
	}
	now := s.clock.Now()
	return s.create(identityID, &now)
}

func (s *MemoryStore) create(identityID string, authenticatedAt *time.Time) (Session, string, error) {
	rawToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	session := Session{ID: s.nextID, CSRFToken: csrfToken, ExpiresAt: s.clock.Now().Add(s.ttl), IdentityID: identityID, AuthenticatedAt: authenticatedAt}
	hash := HashToken(s.sessionKey, rawToken)
	s.byHash[hex.EncodeToString(hash[:])] = memorySession{Session: session}
	return session, rawToken, nil
}

// Get loads a valid in-memory session.
func (s *MemoryStore) Get(ctx context.Context, rawToken string) (Session, error) {
	_ = ctx
	if rawToken == "" {
		return Session{}, ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	hash := HashToken(s.sessionKey, rawToken)
	current, ok := s.byHash[hex.EncodeToString(hash[:])]
	if !ok || current.revoked || !current.ExpiresAt.After(s.clock.Now()) {
		return Session{}, ErrNotFound
	}
	return current.Session, nil
}

// MarkStepUp records a recent passkey step-up on an authenticated session.
func (s *MemoryStore) MarkStepUp(ctx context.Context, id int64, at time.Time) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, current := range s.byHash {
		if current.ID == id && !current.revoked {
			stepUpAt := at.UTC()
			current.StepUpAt = &stepUpAt
			s.byHash[token] = current
			return nil
		}
	}
	return ErrNotFound
}

// Revoke marks an in-memory session unusable.
func (s *MemoryStore) Revoke(ctx context.Context, id int64) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, current := range s.byHash {
		if current.ID == id {
			current.revoked = true
			s.byHash[token] = current
			return nil
		}
	}
	return ErrNotFound
}
