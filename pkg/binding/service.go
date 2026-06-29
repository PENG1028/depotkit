package binding

import (
	"fmt"

	"github.com/depotly/depotly/pkg/store"
)

// Service provides business logic for binding management.
type Service struct {
	bs  *bindingStore
	aud *store.DB
}

// NewService creates a binding service backed by the given store.
func NewService(db *store.DB) *Service {
	return &Service{
		bs:  newBindingStore(db),
		aud: db,
	}
}

// CreateBinding validates and creates a new binding.
func (s *Service) CreateBinding(b *Binding) (*Binding, error) {
	// --- Validation ---
	if b.Service == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if b.ResourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}
	if b.EnvKey == "" {
		return nil, fmt.Errorf("env key is required (e.g. DATABASE_URL)")
	}
	if b.Environment == "" {
		b.Environment = "default"
	}
	if b.AccessRole == "" {
		b.AccessRole = RoleReadWrite
	}
	if b.ProjectID == "" {
		b.ProjectID = "default"
	}
	if b.TenantID == "" {
		b.TenantID = "default"
	}
	if b.CreatedBy == "" {
		b.CreatedBy = "admin"
	}

	// Verify resource exists
	err := s.aud.QueryRow("SELECT id FROM resources WHERE id = ?", b.ResourceID).Scan(new(string))
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s", b.ResourceID)
	}

	// --- Record operation ---
	op := &store.Operation{
		Type:    "create_binding",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Binding %s → %s (%s)", b.Service, b.ResourceID, b.EnvKey),
	}
	s.aud.CreateOperation(op)

	// --- Execute ---
	if err := s.bs.Create(b); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return nil, err
	}

	s.aud.FinishOperation(op.ID, "success",
		fmt.Sprintf("Binding created: %s", b.ID))

	// --- Audit log ---
	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "binding.create",
		TargetType: "binding",
		TargetID:   b.ID,
		Details:    fmt.Sprintf(`{"service":"%s","resource":"%s","env":"%s","key":"%s"}`, b.Service, b.ResourceID, b.Environment, b.EnvKey),
	})

	return b, nil
}

// GetBinding retrieves a binding by ID.
func (s *Service) GetBinding(id string) (*Binding, error) {
	if id == "" {
		return nil, fmt.Errorf("binding ID is required")
	}
	return s.bs.Get(id)
}

// ListBindings returns bindings with resource info.
func (s *Service) ListBindings(filter *Filter) ([]*BindingSummary, error) {
	return s.bs.ListWithResourceInfo(filter)
}

// DeleteBinding removes a binding and audits the action.
func (s *Service) DeleteBinding(id string) error {
	if id == "" {
		return fmt.Errorf("binding ID is required")
	}

	b, err := s.bs.Get(id)
	if err != nil {
		return err
	}

	op := &store.Operation{
		Type:    "delete_binding",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Unbinding %s from %s", b.Service, b.ResourceID),
	}
	s.aud.CreateOperation(op)

	if err := s.bs.Delete(id); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return err
	}

	s.aud.FinishOperation(op.ID, "success", fmt.Sprintf("Binding deleted: %s", id))

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "binding.delete",
		TargetType: "binding",
		TargetID:   id,
		Details:    fmt.Sprintf(`{"service":"%s","resource":"%s","key":"%s"}`, b.Service, b.ResourceID, b.EnvKey),
	})

	return nil
}

// QueryServiceEnv returns env entries for the Runner injection API.
func (s *Service) QueryServiceEnv(service, environment string) ([]*EnvEntry, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if environment == "" {
		environment = "default"
	}
	return s.bs.EnvBindings(service, environment)
}
