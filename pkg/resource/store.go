package resource

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/store"
)

// resourceStore handles SQL operations on the resources table.
// It wraps *store.DB to avoid an import cycle (store → resource).
type resourceStore struct {
	db *store.DB
}

func newResourceStore(db *store.DB) *resourceStore {
	return &resourceStore{db: db}
}

// CreateResource inserts a new resource record and sets its ID.
func (rs *resourceStore) CreateResource(r *Resource) error {
	r.ID = store.GenerateID("res")

	_, err := rs.db.Exec(`
		INSERT INTO resources (
			id, kind, category, environment, project_id, tenant_id,
			name, description,
			host, port, database, username, password,
			owner_service, desired_state, actual_state,
			is_production, is_temporary, is_external,
			created_at, updated_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?,
		          ?, ?,
		          ?, ?, ?, ?, ?,
		          ?, ?, ?,
		          ?, ?, ?,
		          datetime('now'), datetime('now'), ?)`,
		r.ID, r.Kind, r.Category, r.Environment, r.ProjectID, r.TenantID,
		r.Name, r.Description,
		r.Host, r.Port, r.Database, r.Username, r.Password,
		r.OwnerService, r.DesiredState, r.ActualState,
		boolToInt(r.IsProduction), boolToInt(r.IsTemporary), boolToInt(r.IsExternal),
		r.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	return nil
}

// GetResource retrieves a resource by ID.
func (rs *resourceStore) GetResource(id string) (*Resource, error) {
	row := rs.db.QueryRow(`
		SELECT id, kind, category, environment, project_id, tenant_id,
		       name, description,
		       host, port, database, username, password,
		       owner_service, desired_state, actual_state,
		       is_production, is_temporary, is_external,
		       created_at, updated_at, created_by
		FROM resources WHERE id = ?`, id)

	r := &Resource{}
	var isProd, isTemp, isExt int
	err := row.Scan(
		&r.ID, &r.Kind, &r.Category, &r.Environment, &r.ProjectID, &r.TenantID,
		&r.Name, &r.Description,
		&r.Host, &r.Port, &r.Database, &r.Username, &r.Password,
		&r.OwnerService, &r.DesiredState, &r.ActualState,
		&isProd, &isTemp, &isExt,
		&r.CreatedAt, &r.UpdatedAt, &r.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s", id)
	}
	r.IsProduction = isProd != 0
	r.IsTemporary = isTemp != 0
	r.IsExternal = isExt != 0
	return r, nil
}

// ListResources returns resources matching filters.
func (rs *resourceStore) ListResources(filter *Filter) ([]*Resource, error) {
	query := `SELECT id, kind, category, environment, project_id, tenant_id,
	                  name, description,
	                  host, port, database, username, password,
	                  owner_service, desired_state, actual_state,
	                  is_production, is_temporary, is_external,
	                  created_at, updated_at, created_by
	           FROM resources`
	var args []interface{}
	var clauses []string

	if filter != nil {
		if filter.Kind != "" {
			clauses = append(clauses, "kind = ?")
			args = append(args, filter.Kind)
		}
		if filter.Environment != "" {
			clauses = append(clauses, "environment = ?")
			args = append(args, filter.Environment)
		}
		if filter.ProjectID != "" {
			clauses = append(clauses, "project_id = ?")
			args = append(args, filter.ProjectID)
		}
		if filter.OwnerService != "" {
			clauses = append(clauses, "owner_service = ?")
			args = append(args, filter.OwnerService)
		}
		if filter.DesiredState != "" {
			clauses = append(clauses, "desired_state = ?")
			args = append(args, filter.DesiredState)
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY kind, name"

	rows, err := rs.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()

	var results []*Resource
	for rows.Next() {
		r := &Resource{}
		var isProd, isTemp, isExt int
		err := rows.Scan(
			&r.ID, &r.Kind, &r.Category, &r.Environment, &r.ProjectID, &r.TenantID,
			&r.Name, &r.Description,
			&r.Host, &r.Port, &r.Database, &r.Username, &r.Password,
			&r.OwnerService, &r.DesiredState, &r.ActualState,
			&isProd, &isTemp, &isExt,
			&r.CreatedAt, &r.UpdatedAt, &r.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		r.IsProduction = isProd != 0
		r.IsTemporary = isTemp != 0
		r.IsExternal = isExt != 0
		results = append(results, r)
	}
	if results == nil {
		results = []*Resource{}
	}
	return results, rows.Err()
}

// DeleteResource removes a resource, optionally checking for bindings first.
func (rs *resourceStore) DeleteResource(id string, force bool) error {
	if !force {
		var count int
		err := rs.db.QueryRow("SELECT COUNT(*) FROM bindings WHERE resource_id = ?", id).Scan(&count)
		if err != nil {
			return fmt.Errorf("check bindings: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("resource %s has %d active binding(s); use --force to delete anyway, or remove bindings first", id, count)
		}

		err = rs.db.QueryRow("SELECT COUNT(*) FROM access_endpoints WHERE resource_id = ? AND status = 'active'", id).Scan(&count)
		if err != nil {
			return fmt.Errorf("check access endpoints: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("resource %s has %d active access endpoint(s); use --force to revoke and delete, or revoke endpoints first", id, count)
		}
	}

	_, err := rs.db.Exec("DELETE FROM resources WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete resource: %w", err)
	}
	return nil
}

// UpdateResource updates mutable resource fields.
func (rs *resourceStore) UpdateResource(r *Resource) error {
	_, err := rs.db.Exec(`
		UPDATE resources SET
			kind=?, category=?, environment=?, name=?, description=?,
			host=?, port=?, database=?, username=?, password=?,
			owner_service=?, desired_state=?, actual_state=?,
			is_production=?, is_temporary=?, is_external=?,
			updated_at=datetime('now')
		WHERE id=?`,
		r.Kind, r.Category, r.Environment, r.Name, r.Description,
		r.Host, r.Port, r.Database, r.Username, r.Password,
		r.OwnerService, r.DesiredState, r.ActualState,
		boolToInt(r.IsProduction), boolToInt(r.IsTemporary), boolToInt(r.IsExternal),
		r.ID,
	)
	if err != nil {
		return fmt.Errorf("update resource: %w", err)
	}
	return nil
}

// helper to get binding count (used by service and CLI for impact analysis)
func (rs *resourceStore) BindingCount(resourceID string) (int, error) {
	var count int
	err := rs.db.QueryRow("SELECT COUNT(*) FROM bindings WHERE resource_id = ?", resourceID).Scan(&count)
	return count, err
}

// helper to get active access endpoint count
func (rs *resourceStore) AccessEndpointCount(resourceID string) (int, error) {
	var count int
	err := rs.db.QueryRow("SELECT COUNT(*) FROM access_endpoints WHERE resource_id = ? AND status = 'active'", resourceID).Scan(&count)
	return count, err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
