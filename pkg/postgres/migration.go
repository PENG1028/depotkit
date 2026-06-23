package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration represents a single migration file.
type Migration struct {
	Version  string
	Name     string
	Filename string
	FilePath string
	Checksum string
}

// MigrationRecord represents a row in schema_migrations.
type MigrationRecord struct {
	Version        string    `json:"version"`
	Name           string    `json:"name"`
	Checksum       string    `json:"checksum"`
	AppliedAt      time.Time `json:"applied_at"`
	ExecutionTimeMs int64    `json:"execution_time_ms"`
	Dirty          bool      `json:"dirty"`
}

// migrationFileRegex parses filenames like "202606231200_create_users.up.sql"
var migrationFileRegex = regexp.MustCompile(`^(\d{14})_(.+)\.(up|down)\.sql$`)

// ParseMigrationFile parses a migration filename into version and name.
func ParseMigrationFile(filename string) (version, name string, ok bool) {
	matches := migrationFileRegex.FindStringSubmatch(filename)
	if len(matches) < 4 {
		return "", "", false
	}
	return matches[1], matches[2], true
}

// ListMigrations reads migration files from a directory.
func ListMigrations(migrationsDir string) ([]Migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	var upMigrations []Migration

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		version, name, ok := ParseMigrationFile(entry.Name())
		if !ok {
			continue
		}

		filePath := filepath.Join(migrationsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading migration file %s: %w", entry.Name(), err)
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(data))

		upMigrations = append(upMigrations, Migration{
			Version:  version,
			Name:     name,
			Filename: entry.Name(),
			FilePath: filePath,
			Checksum: checksum,
		})
	}

	// Sort by version (lexicographic order)
	sort.Slice(upMigrations, func(i, j int) bool {
		return upMigrations[i].Version < upMigrations[j].Version
	})

	return upMigrations, nil
}

// EnsureMigrationTable creates the schema_migrations table if it doesn't exist.
func EnsureMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		checksum TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT now(),
		execution_time_ms BIGINT NOT NULL DEFAULT 0,
		dirty BOOLEAN NOT NULL DEFAULT FALSE
	)`
	_, err := pool.Exec(ctx, query)
	return err
}

// GetAppliedMigrations returns all applied migration records.
func GetAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) ([]MigrationRecord, error) {
	rows, err := pool.Query(ctx, "SELECT version, name, checksum, applied_at, execution_time_ms, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations: %w", err)
	}
	defer rows.Close()

	var records []MigrationRecord
	for rows.Next() {
		var r MigrationRecord
		if err := rows.Scan(&r.Version, &r.Name, &r.Checksum, &r.AppliedAt, &r.ExecutionTimeMs, &r.Dirty); err != nil {
			return nil, fmt.Errorf("scanning migration record: %w", err)
		}
		records = append(records, r)
	}
	return records, nil
}

// CheckDirty checks if any migration is in dirty state.
func CheckDirty(records []MigrationRecord) *MigrationRecord {
	for _, r := range records {
		if r.Dirty {
			return &r
		}
	}
	return nil
}

// MigrationDiff compares local migrations against applied records.
type MigrationDiff struct {
	Applied    []MigrationRecord
	Pending    []Migration
	Mismatched []MismatchWarning
}

type MismatchWarning struct {
	Version       string
	ExpectedHash  string
	ActualHash    string
}

// DiffMigrations compares local migrations with database records.
func DiffMigrations(local []Migration, records []MigrationRecord) MigrationDiff {
	appliedMap := make(map[string]MigrationRecord)
	for _, r := range records {
		appliedMap[r.Version] = r
	}

	var pending []Migration
	var mismatched []MismatchWarning

	for _, m := range local {
		if record, exists := appliedMap[m.Version]; exists {
			if record.Checksum != m.Checksum {
				mismatched = append(mismatched, MismatchWarning{
					Version:      m.Version,
					ExpectedHash: record.Checksum,
					ActualHash:   m.Checksum,
				})
			}
		} else {
			pending = append(pending, m)
		}
	}

	return MigrationDiff{
		Applied:    records,
		Pending:    pending,
		Mismatched: mismatched,
	}
}

// ApplyMigration applies a single migration.
//
// Safety protocol:
//  1. Mark dirty in its own transaction FIRST (survives crash/disconnect)
//  2. Run migration SQL in a second transaction
//  3. On success → mark clean
//  4. On failure → dirty record persists in DB for investigation
func ApplyMigration(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	// Read migration SQL
	data, err := os.ReadFile(m.FilePath)
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}
	sql := string(data)

	start := time.Now()

	// Step 1: Mark dirty in its own transaction (survives subsequent failures)
	dirtyTx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning dirty-marker transaction: %w", err)
	}

	_, err = dirtyTx.Exec(ctx,
		`INSERT INTO schema_migrations (version, name, checksum, execution_time_ms, dirty)
		 VALUES ($1, $2, $3, 0, true)
		 ON CONFLICT (version) DO UPDATE SET dirty = true, checksum = $3, applied_at = now()`,
		m.Version, m.Name, m.Checksum)
	if err != nil {
		dirtyTx.Rollback(ctx)
		return fmt.Errorf("marking migration dirty: %w", err)
	}

	if err := dirtyTx.Commit(ctx); err != nil {
		return fmt.Errorf("committing dirty marker (may be partially written): %w", err)
	}

	// Step 2: Execute migration SQL in its own transaction
	migTx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning migration transaction (dirty marker preserved): %w", err)
	}
	defer migTx.Rollback(ctx)

	if _, err := migTx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("executing migration %s (dirty marker preserved in DB): %w", m.Filename, err)
	}

	elapsed := time.Since(start).Milliseconds()

	// Step 3: Mark clean
	_, err = migTx.Exec(ctx,
		`UPDATE schema_migrations SET dirty = false, execution_time_ms = $1 WHERE version = $2`,
		elapsed, m.Version)
	if err != nil {
		return fmt.Errorf("marking migration clean (dirty marker preserved): %w", err)
	}

	if err := migTx.Commit(ctx); err != nil {
		return fmt.Errorf("committing migration %s (dirty=true remains): %w", m.Filename, err)
	}

	return nil
}
