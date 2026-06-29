package binding

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/store"
)

// bindingStore handles SQL operations on the bindings table.
type bindingStore struct {
	db *store.DB
}

func newBindingStore(db *store.DB) *bindingStore {
	return &bindingStore{db: db}
}

// Create inserts a new binding.
func (bs *bindingStore) Create(b *Binding) error {
	b.ID = store.GenerateID("bnd")

	_, err := bs.db.Exec(`
		INSERT INTO bindings (id, service, environment, resource_id, env_key,
		                      access_role, required, project_id, tenant_id,
		                      created_at, created_by)
		VALUES (?, ?, ?, ?, ?,
		        ?, ?, ?, ?,
		        datetime('now'), ?)`,
		b.ID, b.Service, b.Environment, b.ResourceID, b.EnvKey,
		b.AccessRole, boolToInt(b.Required), b.ProjectID, b.TenantID,
		b.CreatedBy,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE") {
			return fmt.Errorf("binding already exists for service '%s' env '%s' key '%s'", b.Service, b.Environment, b.EnvKey)
		}
		if strings.Contains(errMsg, "FOREIGN KEY") {
			return fmt.Errorf("resource not found: %s", b.ResourceID)
		}
		return fmt.Errorf("create binding: %w", err)
	}
	return nil
}

// Get retrieves a binding by ID.
func (bs *bindingStore) Get(id string) (*Binding, error) {
	row := bs.db.QueryRow(`
		SELECT id, service, environment, resource_id, env_key,
		       access_role, required, project_id, tenant_id,
		       created_at, created_by
		FROM bindings WHERE id = ?`, id)

	b := &Binding{}
	var req int
	err := row.Scan(
		&b.ID, &b.Service, &b.Environment, &b.ResourceID, &b.EnvKey,
		&b.AccessRole, &req, &b.ProjectID, &b.TenantID,
		&b.CreatedAt, &b.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %s", id)
	}
	b.Required = req != 0
	return b, nil
}

// List returns bindings matching the filter.
func (bs *bindingStore) List(filter *Filter) ([]*Binding, error) {
	query := `SELECT id, service, environment, resource_id, env_key,
	                 access_role, required, project_id, tenant_id,
	                 created_at, created_by
	          FROM bindings`
	var args []interface{}
	var clauses []string

	if filter != nil {
		if filter.Service != "" {
			clauses = append(clauses, "service = ?")
			args = append(args, filter.Service)
		}
		if filter.Environment != "" {
			clauses = append(clauses, "environment = ?")
			args = append(args, filter.Environment)
		}
		if filter.ResourceID != "" {
			clauses = append(clauses, "resource_id = ?")
			args = append(args, filter.ResourceID)
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY service, env_key"

	rows, err := bs.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list bindings: %w", err)
	}
	defer rows.Close()

	var results []*Binding
	for rows.Next() {
		b := &Binding{}
		var req int
		err := rows.Scan(
			&b.ID, &b.Service, &b.Environment, &b.ResourceID, &b.EnvKey,
			&b.AccessRole, &req, &b.ProjectID, &b.TenantID,
			&b.CreatedAt, &b.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan binding: %w", err)
		}
		b.Required = req != 0
		results = append(results, b)
	}
	if results == nil {
		results = []*Binding{}
	}
	return results, rows.Err()
}

// ListWithResourceInfo returns bindings joined with resource name/kind.
func (bs *bindingStore) ListWithResourceInfo(filter *Filter) ([]*BindingSummary, error) {
	query := `SELECT b.id, b.service, b.environment, b.env_key,
	                 b.access_role, b.required,
	                 b.resource_id, r.name, r.kind, r.environment
	          FROM bindings b
	          JOIN resources r ON r.id = b.resource_id`
	var args []interface{}
	var clauses []string

	if filter != nil {
		if filter.Service != "" {
			clauses = append(clauses, "b.service = ?")
			args = append(args, filter.Service)
		}
		if filter.Environment != "" {
			clauses = append(clauses, "b.environment = ?")
			args = append(args, filter.Environment)
		}
		if filter.ResourceID != "" {
			clauses = append(clauses, "b.resource_id = ?")
			args = append(args, filter.ResourceID)
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY b.service, b.env_key"

	rows, err := bs.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list bindings with resource: %w", err)
	}
	defer rows.Close()

	var results []*BindingSummary
	for rows.Next() {
		s := &BindingSummary{}
		var req int
		err := rows.Scan(
			&s.ID, &s.Service, &s.Environment, &s.EnvKey,
			&s.AccessRole, &req,
			&s.ResourceID, &s.ResourceName, &s.ResourceKind, &s.ResourceEnv,
		)
		if err != nil {
			return nil, fmt.Errorf("scan binding summary: %w", err)
		}
		s.Required = req != 0
		results = append(results, s)
	}
	if results == nil {
		results = []*BindingSummary{}
	}
	return results, rows.Err()
}

// Delete removes a binding by ID.
func (bs *bindingStore) Delete(id string) error {
	res, err := bs.db.Exec("DELETE FROM bindings WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete binding: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("binding not found: %s", id)
	}
	return nil
}

// EnvBindings returns env entries for a service in an environment (Runner API format).
func (bs *bindingStore) EnvBindings(service, environment string) ([]*EnvEntry, error) {
	rows, err := bs.db.Query(`
		SELECT b.env_key, r.kind, r.name, b.required
		FROM bindings b
		JOIN resources r ON r.id = b.resource_id
		WHERE b.service = ? AND b.environment = ?`, service, environment)
	if err != nil {
		return nil, fmt.Errorf("query env bindings: %w", err)
	}
	defer rows.Close()

	var results []*EnvEntry
	for rows.Next() {
		e := &EnvEntry{Required: true}
		var kind, name string
		var req int
		if err := rows.Scan(&e.EnvKey, &kind, &name, &req); err != nil {
			return nil, fmt.Errorf("scan env entry: %w", err)
		}
		e.Required = req != 0
		e.SecretRef = fmt.Sprintf("secret://default/%s/%s/%s", environment, kind, name)
		results = append(results, e)
	}
	if results == nil {
		results = []*EnvEntry{}
	}
	return results, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
