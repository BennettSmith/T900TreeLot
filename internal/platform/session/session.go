// Package session persists hashed browser session tokens and CSRF secrets.
package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

// ErrNotFound indicates the session token is unknown, expired, or revoked.
var ErrNotFound = errors.New("session not found")

// Session is a durable anonymous browser session used for CSRF protection.
type Session struct {
	ID        int64
	CSRFToken string
	ExpiresAt time.Time
}

// Store reads and writes session rows.
type Store struct {
	db    *postgres.DB
	clock clock.Clock
	ttl   time.Duration
}

// NewStore constructs a session store.
func NewStore(db *postgres.DB, clk clock.Clock, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Store{db: db, clock: clk, ttl: ttl}
}

// Create inserts a new session and returns the raw cookie token.
func (s *Store) Create(ctx context.Context) (Session, string, error) {
	rawToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	expiresAt := s.clock.Now().Add(s.ttl)
	hash := hashToken(rawToken)

	var id int64
	err = s.db.QueryRow(ctx, `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id
	`, hash[:], csrfToken, expiresAt).Scan(&id)
	if err != nil {
		return Session{}, "", fmt.Errorf("create session: %w", err)
	}
	return Session{ID: id, CSRFToken: csrfToken, ExpiresAt: expiresAt}, rawToken, nil
}

// Get loads a valid session by raw cookie token.
func (s *Store) Get(ctx context.Context, rawToken string) (Session, error) {
	if rawToken == "" {
		return Session{}, ErrNotFound
	}
	hash := hashToken(rawToken)
	var session Session
	err := s.db.QueryRow(ctx, `
		SELECT id, csrf_token, expires_at
		FROM sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`, hash[:], s.clock.Now()).Scan(&session.ID, &session.CSRFToken, &session.ExpiresAt)
	if err != nil {
		return Session{}, ErrNotFound
	}
	_, _ = s.db.Exec(ctx, `UPDATE sessions SET last_seen_at = $2 WHERE id = $1`, session.ID, s.clock.Now())
	return session, nil
}

// Revoke marks a session unusable.
func (s *Store) Revoke(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`, id, s.clock.Now())
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func hashToken(raw string) [32]byte {
	return sha256.Sum256([]byte(raw))
}

func randomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
