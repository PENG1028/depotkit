// Package store provides SQLite-based persistence for DBManager metadata.
//
// It stores Resource, Binding, AccessEndpoint, Operation, and AuditLog records
// in a local SQLite database. This is the single source of truth for DBManager's
// resource control layer, separate from depotly.yaml which remains the config
// for Docker-level execution.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection and provides store operations.
type DB struct {
	*sql.DB
	path string
}

// Open opens or creates the SQLite database at the given path,
// runs migrations, and returns the DB handle.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	// Enable WAL mode for better concurrent reads
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	db := &DB{DB: sqlDB, path: path}

	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate store: %w", err)
	}

	return db, nil
}

// Path returns the database file path.
func (db *DB) Path() string { return db.path }

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// migrate creates tables if they don't exist.
func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS resources (
		id           TEXT PRIMARY KEY,
		kind         TEXT NOT NULL,
		category     TEXT NOT NULL DEFAULT 'relational',
		environment  TEXT NOT NULL DEFAULT 'default',
		project_id   TEXT NOT NULL DEFAULT 'default',
		tenant_id    TEXT NOT NULL DEFAULT 'default',

		name         TEXT NOT NULL,
		description  TEXT NOT NULL DEFAULT '',

		host         TEXT NOT NULL DEFAULT '',
		port         INTEGER NOT NULL DEFAULT 0,
		database     TEXT NOT NULL DEFAULT '',
		username     TEXT NOT NULL DEFAULT '',
		password     TEXT NOT NULL DEFAULT '',

		owner_service TEXT NOT NULL DEFAULT '',
		desired_state TEXT NOT NULL DEFAULT 'active',
		actual_state  TEXT NOT NULL DEFAULT 'unknown',
		is_production INTEGER NOT NULL DEFAULT 0,
		is_temporary  INTEGER NOT NULL DEFAULT 0,
		is_external   INTEGER NOT NULL DEFAULT 0,

		created_at   TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at   TEXT NOT NULL DEFAULT (datetime('now')),
		created_by   TEXT NOT NULL DEFAULT 'admin'
	);

	CREATE TABLE IF NOT EXISTS bindings (
		id           TEXT PRIMARY KEY,
		service      TEXT NOT NULL,
		environment  TEXT NOT NULL DEFAULT 'default',
		resource_id  TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
		env_key      TEXT NOT NULL,
		access_role  TEXT NOT NULL DEFAULT 'read_write',
		required     INTEGER NOT NULL DEFAULT 1,
		project_id   TEXT NOT NULL DEFAULT 'default',
		tenant_id    TEXT NOT NULL DEFAULT 'default',

		created_at   TEXT NOT NULL DEFAULT (datetime('now')),
		created_by   TEXT NOT NULL DEFAULT 'admin',

		UNIQUE(service, environment, env_key)
	);

	CREATE TABLE IF NOT EXISTS access_endpoints (
		id           TEXT PRIMARY KEY,
		resource_id  TEXT NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
		type         TEXT NOT NULL,
		host         TEXT NOT NULL,
		port         INTEGER NOT NULL,
		target_host  TEXT NOT NULL,
		target_port  INTEGER NOT NULL,
		status       TEXT NOT NULL DEFAULT 'active',
		expires_at   TEXT,
		created_at   TEXT NOT NULL DEFAULT (datetime('now')),
		created_by   TEXT NOT NULL DEFAULT 'admin'
	);

	CREATE TABLE IF NOT EXISTS operations (
		id           TEXT PRIMARY KEY,
		type         TEXT NOT NULL,
		resource_id  TEXT,
		status       TEXT NOT NULL DEFAULT 'pending',
		actor        TEXT NOT NULL DEFAULT 'admin',
		message      TEXT NOT NULL DEFAULT '',
		started_at   TEXT NOT NULL DEFAULT (datetime('now')),
		finished_at  TEXT
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id           TEXT PRIMARY KEY,
		actor        TEXT NOT NULL,
		action       TEXT NOT NULL,
		target_type  TEXT NOT NULL DEFAULT '',
		target_id    TEXT NOT NULL DEFAULT '',
		details      TEXT NOT NULL DEFAULT '{}',
		created_at   TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_bindings_service ON bindings(service, environment);
	CREATE INDEX IF NOT EXISTS idx_bindings_resource ON bindings(resource_id);
	CREATE INDEX IF NOT EXISTS idx_access_resource ON access_endpoints(resource_id, status);
	CREATE INDEX IF NOT EXISTS idx_operations_type ON operations(type);
	CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action, created_at);
	`

	_, err := db.Exec(schema)
	return err
}

// GenerateID creates a simple unique ID with a prefix.
func GenerateID(prefix string) string {
	// Use a timestamp-based approach for simplicity without external deps.
	// Format: <prefix>_<timestamp>_<random>
	// We rely on the store layer to keep this simple.
	// A proper UUID library can be added later if needed.
	return prefix + "_" + fmt.Sprintf("%x", os.Getpid()) + "_" + fmt.Sprintf("%d", nanoTime())
}

// nanoTime returns a monotonic-like timestamp for ID generation.
func nanoTime() int64 {
	// Use a simple counter that increments per-call for uniqueness within a process.
	// This avoids importing time just for ID generation.
	return fastCounter()
}

var counter int64 = 0

func fastCounter() int64 {
	counter++
	return counter
}
