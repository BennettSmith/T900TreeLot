// Package session persists hashed browser session tokens and CSRF secrets.
package session

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/troop900/treelot/internal/platform/clock"
	"github.com/troop900/treelot/internal/platform/postgres"
)

// ErrNotFound indicates the session token is unknown, expired, or revoked.
var ErrNotFound = errors.New("session not found")

// TestKey is a fixed ≥32-byte SESSION_KEY for focused unit tests.
var TestKey = []byte("0123456789abcdef0123456789abcdef")

// Session is a durable anonymous browser session used for CSRF protection.
type Session struct {
	ID              int64
	CSRFToken       string
	ExpiresAt       time.Time
	IdentityID      string
	AuthenticatedAt *time.Time
	StepUpAt        *time.Time
}

// Store reads and writes session rows.
type Store struct {
	db         *postgres.DB
	clock      clock.Clock
	ttl        time.Duration
	sessionKey []byte
}

// NewStore constructs a session store. sessionKey is the HMAC key used to hash
// opaque cookie tokens; rotating it invalidates existing sessions.
func NewStore(db *postgres.DB, clk clock.Clock, ttl time.Duration, sessionKey []byte) *Store {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Store{db: db, clock: clk, ttl: ttl, sessionKey: append([]byte(nil), sessionKey...)}
}

// Create inserts a new session and returns the raw cookie token.
func (s *Store) Create(ctx context.Context) (Session, string, error) {
	return s.create(ctx, s.db, "", nil)
}

// CreateForIdentity inserts a new authenticated session and returns the raw cookie token.
func (s *Store) CreateForIdentity(ctx context.Context, identityID string) (Session, string, error) {
	if identityID == "" {
		return Session{}, "", fmt.Errorf("identity id is required")
	}
	now := s.clock.Now()
	return s.create(ctx, s.db, identityID, &now)
}

func (s *Store) create(ctx context.Context, exec sessionExecutor, identityID string, authenticatedAt *time.Time) (Session, string, error) {
	rawToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		return Session{}, "", err
	}
	expiresAt := s.clock.Now().Add(s.ttl)
	hash := HashToken(s.sessionKey, rawToken)

	var id int64
	err = exec.QueryRow(ctx, `
		INSERT INTO sessions (token_hash, csrf_token, expires_at, last_seen_at, identity_id, authenticated_at)
		VALUES ($1, $2, $3, $3, NULLIF($4, ''), $5)
		RETURNING id
	`, hash[:], csrfToken, expiresAt, identityID, authenticatedAt).Scan(&id)
	if err != nil {
		return Session{}, "", fmt.Errorf("create session: %w", err)
	}
	return Session{ID: id, CSRFToken: csrfToken, ExpiresAt: expiresAt, IdentityID: identityID, AuthenticatedAt: authenticatedAt}, rawToken, nil
}

// Get loads a valid session by raw cookie token.
func (s *Store) Get(ctx context.Context, rawToken string) (Session, error) {
	if rawToken == "" {
		return Session{}, ErrNotFound
	}
	hash := HashToken(s.sessionKey, rawToken)
	var session Session
	var identityID *string
	var authenticatedAt *time.Time
	var stepUpAt *time.Time
	err := s.db.QueryRow(ctx, `
		SELECT id, csrf_token, expires_at, identity_id, authenticated_at, step_up_at
		FROM sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
	`, hash[:], s.clock.Now()).Scan(&session.ID, &session.CSRFToken, &session.ExpiresAt, &identityID, &authenticatedAt, &stepUpAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get session: %w", err)
	}
	if identityID != nil {
		session.IdentityID = *identityID
	}
	session.AuthenticatedAt = authenticatedAt
	session.StepUpAt = stepUpAt
	_, _ = s.db.Exec(ctx, `UPDATE sessions SET last_seen_at = $2 WHERE id = $1`, session.ID, s.clock.Now())
	return session, nil
}

// BindIdentity marks an existing session as authenticated.
func (s *Store) BindIdentity(ctx context.Context, id int64, identityID string) error {
	if identityID == "" {
		return fmt.Errorf("identity id is required")
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE sessions
		SET identity_id = $2, authenticated_at = $3, last_seen_at = $3
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > $3
	`, id, identityID, s.clock.Now())
	if err != nil {
		return fmt.Errorf("bind session identity: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RotateForIdentity revokes an existing session and creates a fresh authenticated one.
func (s *Store) RotateForIdentity(ctx context.Context, oldID int64, identityID string) (Session, string, error) {
	if identityID == "" {
		return Session{}, "", fmt.Errorf("identity id is required")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, "", fmt.Errorf("begin session rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`, oldID, s.clock.Now())
	if err != nil {
		return Session{}, "", fmt.Errorf("revoke old session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Session{}, "", ErrNotFound
	}
	now := s.clock.Now()
	created, rawToken, err := s.create(ctx, tx, identityID, &now)
	if err != nil {
		return Session{}, "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, "", fmt.Errorf("commit session rotation: %w", err)
	}
	return created, rawToken, nil
}

// RotateForIdentityInTx performs session rotation inside a caller-owned transaction.
func (s *Store) RotateForIdentityInTx(ctx context.Context, tx pgx.Tx, oldID int64, identityID string) (Session, string, error) {
	if identityID == "" {
		return Session{}, "", fmt.Errorf("identity id is required")
	}
	tag, err := tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`, oldID, s.clock.Now())
	if err != nil {
		return Session{}, "", fmt.Errorf("revoke old session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Session{}, "", ErrNotFound
	}
	now := s.clock.Now()
	return s.create(ctx, tx, identityID, &now)
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

// HashToken returns the HMAC-SHA256 digest of a raw session token under sessionKey.
// Rotating sessionKey changes the digest and invalidates stored lookups.
func HashToken(sessionKey []byte, raw string) [32]byte {
	mac := hmac.New(sha256.New, sessionKey)
	_, _ = mac.Write([]byte(raw))
	var out [32]byte
	copy(out[:], mac.Sum(nil))
	return out
}

func randomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

type sessionExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
