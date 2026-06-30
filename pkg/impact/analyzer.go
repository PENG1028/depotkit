// Package impact provides impact analysis for resource operations.
//
// It answers: "What will be affected if I delete this resource?
// Which services depend on it? What access endpoints are active?"
package impact

import (
	"fmt"

	"github.com/depotly/depotly/pkg/access"
	"github.com/depotly/depotly/pkg/binding"
	"github.com/depotly/depotly/pkg/resource"
	"github.com/depotly/depotly/pkg/store"
)

// Risk levels.
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)

// AnalysisResult holds the complete impact assessment.
type AnalysisResult struct {
	Resource     *resource.Resource
	Bindings     []*binding.BindingSummary
	AccessPoints []*access.AccessEndpoint
	RiskLevel    string
}

// Analyzer performs impact analysis using store queries.
type Analyzer struct {
	db *store.DB
}

// New creates a new Analyzer.
func New(db *store.DB) *Analyzer {
	return &Analyzer{db: db}
}

// AnalyzeResource checks what depends on the given resource.
func (a *Analyzer) AnalyzeResource(resourceID string) (*AnalysisResult, error) {
	// Get resource
	r, err := a.resourceByID(resourceID)
	if err != nil {
		return nil, err
	}

	// Get bindings with full info
	bindings, err := a.bindingsForResource(resourceID)
	if err != nil {
		return nil, fmt.Errorf("query bindings: %w", err)
	}

	// Get active access endpoints
	accessPoints, err := a.activeAccessForResource(resourceID)
	if err != nil {
		return nil, fmt.Errorf("query access: %w", err)
	}

	// Calculate risk
	risk := calculateRisk(r, bindings, accessPoints)

	return &AnalysisResult{
		Resource:     r,
		Bindings:     bindings,
		AccessPoints: accessPoints,
		RiskLevel:    risk,
	}, nil
}

// AnalyzeService checks all resources bound to a service.
func (a *Analyzer) AnalyzeService(service, environment string) ([]*AnalysisResult, error) {
	var results []*AnalysisResult

	bs := binding.NewService(a.db)
	filter := &binding.Filter{
		Service:     service,
		Environment: environment,
	}

	summaryList, err := bs.ListBindings(filter)
	if err != nil {
		return nil, fmt.Errorf("query service bindings: %w", err)
	}

	seen := make(map[string]bool)
	for _, s := range summaryList {
		if seen[s.ResourceID] {
			continue
		}
		seen[s.ResourceID] = true
		result, err := a.AnalyzeResource(s.ResourceID)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// AnalyzeEnvironment checks all resources in a given environment.
func (a *Analyzer) AnalyzeEnvironment(environment string) ([]*AnalysisResult, error) {
	var results []*AnalysisResult

	rs := resource.NewService(a.db)
	filter := &resource.Filter{
		Environment: environment,
	}

	resources, err := rs.ListResources(filter)
	if err != nil {
		return nil, fmt.Errorf("query resources: %w", err)
	}

	for _, r := range resources {
		result, err := a.AnalyzeResource(r.ID)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// --- internal helpers ---

func (a *Analyzer) resourceByID(id string) (*resource.Resource, error) {
	row := a.db.QueryRow(`
		SELECT id, kind, name, environment, owner_service,
		       is_production, is_temporary
		FROM resources WHERE id = ?`, id)

	r := &resource.Resource{}
	var isProd, isTemp int
	err := row.Scan(&r.ID, &r.Kind, &r.Name, &r.Environment, &r.OwnerService, &isProd, &isTemp)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s", id)
	}
	r.IsProduction = isProd != 0
	r.IsTemporary = isTemp != 0
	return r, nil
}

func (a *Analyzer) bindingsForResource(id string) ([]*binding.BindingSummary, error) {
	rows, err := a.db.Query(`
		SELECT b.id, b.service, b.environment, b.env_key,
		       b.access_role, b.required,
		       b.resource_id,
		       COALESCE(r.name, '(deleted)'), COALESCE(r.kind, ''), COALESCE(r.environment, '')
		FROM bindings b
		LEFT JOIN resources r ON r.id = b.resource_id
		WHERE b.resource_id = ?
		ORDER BY b.service`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*binding.BindingSummary
	for rows.Next() {
		s := &binding.BindingSummary{}
		var req int
		err := rows.Scan(
			&s.ID, &s.Service, &s.Environment, &s.EnvKey,
			&s.AccessRole, &req,
			&s.ResourceID, &s.ResourceName, &s.ResourceKind, &s.ResourceEnv,
		)
		if err != nil {
			return nil, err
		}
		s.Required = req != 0
		results = append(results, s)
	}
	if results == nil {
		results = []*binding.BindingSummary{}
	}
	return results, rows.Err()
}

func (a *Analyzer) activeAccessForResource(id string) ([]*access.AccessEndpoint, error) {
	rows, err := a.db.Query(`
		SELECT id, resource_id, type, host, port,
		       target_host, target_port, status,
		       COALESCE(expires_at, ''), created_at, created_by
		FROM access_endpoints
		WHERE resource_id = ? AND status = 'active'
		ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*access.AccessEndpoint
	for rows.Next() {
		e := &access.AccessEndpoint{}
		err := rows.Scan(
			&e.ID, &e.ResourceID, &e.Type, &e.Host, &e.Port,
			&e.TargetHost, &e.TargetPort, &e.Status,
			&e.ExpiresAt, &e.CreatedAt, &e.CreatedBy,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	if results == nil {
		results = []*access.AccessEndpoint{}
	}
	return results, rows.Err()
}

func calculateRisk(r *resource.Resource, bindings []*binding.BindingSummary, accessPoints []*access.AccessEndpoint) string {
	if r.IsProduction && len(bindings) > 0 {
		return RiskHigh
	}
	if r.IsProduction {
		return RiskHigh
	}
	if len(bindings) > 0 || len(accessPoints) > 0 {
		return RiskMedium
	}
	return RiskLow
}
