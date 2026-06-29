package store

import "fmt"

// Operation represents a tracked operation in the system.
type Operation struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	ResourceID string `json:"resource_id,omitempty"`
	Status     string `json:"status"` // pending, running, success, failed
	Actor      string `json:"actor"`
	Message    string `json:"message"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// AuditLog represents a single audit trail entry.
type AuditLog struct {
	ID         string `json:"id"`
	Actor      string `json:"actor"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Details    string `json:"details"` // JSON string
	CreatedAt  string `json:"created_at"`
}

// CreateOperation inserts a new operation record.
func (db *DB) CreateOperation(op *Operation) error {
	op.ID = GenerateID("op")
	_, err := db.Exec(`
		INSERT INTO operations (id, type, resource_id, status, actor, message, started_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		op.ID, op.Type, op.ResourceID, op.Status, op.Actor, op.Message)
	if err != nil {
		return fmt.Errorf("create operation: %w", err)
	}
	return nil
}

// FinishOperation updates an operation's status and sets finished_at.
func (db *DB) FinishOperation(id, status, message string) error {
	_, err := db.Exec(`
		UPDATE operations SET status = ?, message = ?, finished_at = datetime('now')
		WHERE id = ?`, status, message, id)
	if err != nil {
		return fmt.Errorf("finish operation: %w", err)
	}
	return nil
}

// ListOperations returns operations with optional type filter.
func (db *DB) ListOperations(opType string, limit int) ([]*Operation, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows interface{ Close() error; Next() bool; Scan(...interface{}) error; Err() error }
	var err error

	if opType != "" {
		rows, err = db.Query(
			"SELECT id, type, resource_id, status, actor, message, started_at, COALESCE(finished_at,'') FROM operations WHERE type = ? ORDER BY started_at DESC LIMIT ?",
			opType, limit)
	} else {
		rows, err = db.Query(
			"SELECT id, type, resource_id, status, actor, message, started_at, COALESCE(finished_at,'') FROM operations ORDER BY started_at DESC LIMIT ?",
			limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list operations: %w", err)
	}
	defer rows.Close()

	var results []*Operation
	for rows.Next() {
		op := &Operation{}
		if err := rows.Scan(&op.ID, &op.Type, &op.ResourceID, &op.Status, &op.Actor, &op.Message, &op.StartedAt, &op.FinishedAt); err != nil {
			return nil, fmt.Errorf("scan operation: %w", err)
		}
		results = append(results, op)
	}
	if results == nil {
		results = []*Operation{}
	}
	return results, rows.Err()
}

// AppendAuditLog inserts a new audit log entry.
func (db *DB) AppendAuditLog(log *AuditLog) error {
	log.ID = GenerateID("aud")
	_, err := db.Exec(`
		INSERT INTO audit_logs (id, actor, action, target_type, target_id, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		log.ID, log.Actor, log.Action, log.TargetType, log.TargetID, log.Details)
	if err != nil {
		return fmt.Errorf("append audit log: %w", err)
	}
	return nil
}

// ListAuditLogs returns audit logs with optional filters.
func (db *DB) ListAuditLogs(action, targetType string, limit int) ([]*AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}

	query := "SELECT id, actor, action, target_type, target_id, details, created_at FROM audit_logs"
	var args []interface{}
	var clauses []string

	if action != "" {
		clauses = append(clauses, "action = ?")
		args = append(args, action)
	}
	if targetType != "" {
		clauses = append(clauses, "target_type = ?")
		args = append(args, targetType)
	}

	if len(clauses) > 0 {
		query += " WHERE " + joinClauses(clauses)
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var results []*AuditLog
	for rows.Next() {
		l := &AuditLog{}
		if err := rows.Scan(&l.ID, &l.Actor, &l.Action, &l.TargetType, &l.TargetID, &l.Details, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		results = append(results, l)
	}
	if results == nil {
		results = []*AuditLog{}
	}
	return results, rows.Err()
}

func joinClauses(clauses []string) string {
	result := clauses[0]
	for i := 1; i < len(clauses); i++ {
		result += " AND " + clauses[i]
	}
	return result
}
