// Package migrate applies versioned SQL migrations and checks schema compatibility.
package migrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/troop900/treelot/internal/platform/postgres"
)

var migrationName = regexp.MustCompile(`^(\d{6})_.*\.sql$`)

type migrationFile struct {
	version  int
	path     string
	contents string
}

// Up applies all pending migrations from directory and returns how many ran.
func Up(ctx context.Context, db *postgres.DB, directory string) (int, error) {
	if err := ensureBookkeeping(ctx, db); err != nil {
		return 0, err
	}
	files, err := loadMigrations(directory)
	if err != nil {
		return 0, err
	}
	current, err := CurrentVersion(ctx, db)
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, file := range files {
		if file.version <= current {
			continue
		}
		tx, err := db.Begin(ctx)
		if err != nil {
			return applied, fmt.Errorf("begin migration %d: %w", file.version, err)
		}
		if _, err := tx.Exec(ctx, file.contents); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("apply migration %d: %w", file.version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, file.version); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("record migration %d: %w", file.version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, fmt.Errorf("commit migration %d: %w", file.version, err)
		}
		applied++
	}
	return applied, nil
}

// CurrentVersion returns the highest applied migration version, or 0.
// It is read-only and never creates schema_migrations; only Up may mutate schema.
func CurrentVersion(ctx context.Context, db *postgres.DB) (int, error) {
	var version *int
	err := db.QueryRow(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&version)
	if err != nil {
		if isUndefinedTable(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read schema version: %w", err)
	}
	if version == nil {
		return 0, nil
	}
	return *version, nil
}

// EnsureCompatible fails when the database schema version does not match expected.
// It never mutates schema; an unmigrated database is reported as incompatible.
func EnsureCompatible(ctx context.Context, db *postgres.DB, expected int) error {
	version, err := CurrentVersion(ctx, db)
	if err != nil {
		return err
	}
	if version != expected {
		return fmt.Errorf("schema version %d is incompatible with expected version %d", version, expected)
	}
	return nil
}

func ensureBookkeeping(ctx context.Context, db *postgres.DB) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func loadMigrations(directory string) ([]migrationFile, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}
	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationName.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry.Name(), err)
		}
		path := filepath.Join(directory, entry.Name())
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		files = append(files, migrationFile{
			version:  version,
			path:     path,
			contents: string(contents),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	for i := 1; i < len(files); i++ {
		if files[i].version == files[i-1].version {
			return nil, fmt.Errorf("duplicate migration version %d", files[i].version)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no migration files found in %s", directory)
	}
	return files, nil
}

// LatestAvailableVersion returns the highest migration file version on disk.
func LatestAvailableVersion(directory string) (int, error) {
	files, err := loadMigrations(directory)
	if err != nil {
		return 0, err
	}
	return files[len(files)-1].version, nil
}

// Directory resolves the migrations directory relative to the working directory
// or a well-known absolute layout inside the container image.
func Directory(candidates ...string) (string, error) {
	if len(candidates) == 0 {
		candidates = []string{
			"migrations",
			"/app/migrations",
			filepath.Join("..", "migrations"),
			filepath.Join("..", "..", "migrations"),
		}
	}
	var tried []string
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
		tried = append(tried, candidate)
	}
	return "", fmt.Errorf("migrations directory not found (tried %s)", strings.Join(tried, ", "))
}
