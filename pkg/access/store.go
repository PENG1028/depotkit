package access

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/store"
)

// accessStore handles SQL operations on the access_endpoints table.
type accessStore struct {
	db *store.DB
}

func newAccessStore(db *store.DB) *accessStore {
	return &accessStore{db: db}
}

// Create inserts a new access endpoint record.
func (as *accessStore) Create(e *AccessEndpoint) error {
	e.ID = store.GenerateID("acc")

	_, err := as.db.Exec(`
		INSERT INTO access_endpoints (id, resource_id, type, host, port,
		                              target_host, target_port, status, expires_at,
		                              created_at, created_by)
		VALUES (?, ?, ?, ?, ?,
		        ?, ?, ?, ?,
		        datetime('now'), ?)`,
		e.ID, e.ResourceID, e.Type, e.Host, e.Port,
		e.TargetHost, e.TargetPort, e.Status, e.ExpiresAt,
		e.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("create access endpoint: %w", err)
	}
	return nil
}

// Get retrieves an access endpoint by ID.
func (as *accessStore) Get(id string) (*AccessEndpoint, error) {
	row := as.db.QueryRow(`
		SELECT id, resource_id, type, host, port,
		       target_host, target_port, status,
		       COALESCE(expires_at, ''), created_at, created_by
		FROM access_endpoints WHERE id = ?`, id)

	e := &AccessEndpoint{}
	err := row.Scan(
		&e.ID, &e.ResourceID, &e.Type, &e.Host, &e.Port,
		&e.TargetHost, &e.TargetPort, &e.Status,
		&e.ExpiresAt, &e.CreatedAt, &e.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("access endpoint not found: %s", id)
	}
	return e, nil
}

// List returns access endpoints matching the filter.
func (as *accessStore) List(filter *Filter) ([]*AccessEndpoint, error) {
	query := `SELECT id, resource_id, type, host, port,
	                 target_host, target_port, status,
	                 COALESCE(expires_at, ''), created_at, created_by
	          FROM access_endpoints`
	var args []interface{}
	var clauses []string

	if filter != nil {
		if filter.ResourceID != "" {
			clauses = append(clauses, "resource_id = ?")
			args = append(args, filter.ResourceID)
		}
		if filter.Type != "" {
			clauses = append(clauses, "type = ?")
			args = append(args, filter.Type)
		}
		if filter.Status != "" {
			clauses = append(clauses, "status = ?")
			args = append(args, filter.Status)
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := as.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list access endpoints: %w", err)
	}
	defer rows.Close()

	var results []*AccessEndpoint
	for rows.Next() {
		e := &AccessEndpoint{}
		err := rows.Scan(
			&e.ID, &e.ResourceID, &e.Type, &e.Host, &e.Port,
			&e.TargetHost, &e.TargetPort, &e.Status,
			&e.ExpiresAt, &e.CreatedAt, &e.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan access endpoint: %w", err)
		}
		results = append(results, e)
	}
	if results == nil {
		results = []*AccessEndpoint{}
	}
	return results, rows.Err()
}

// UpdateStatus changes the status of an access endpoint.
func (as *accessStore) UpdateStatus(id, status string) error {
	res, err := as.db.Exec("UPDATE access_endpoints SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("update access endpoint status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("access endpoint not found: %s", id)
	}
	return nil
}
