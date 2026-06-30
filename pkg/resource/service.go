package resource

import (
	"fmt"

	"github.com/depotly/depotly/pkg/store"
)

// Service provides business logic for resource management.
type Service struct {
	rs  *resourceStore
	aud *store.DB
}

// NewService creates a resource service backed by the given store.
func NewService(db *store.DB) *Service {
	return &Service{
		rs:  newResourceStore(db),
		aud: db,
	}
}

// CreateResource validates and registers a new resource.
func (s *Service) CreateResource(r *Resource) (*Resource, error) {
	if r.Name == "" {
		return nil, fmt.Errorf("resource name is required")
	}
	if r.Kind == "" {
		return nil, fmt.Errorf("resource kind is required")
	}
	if r.Category == "" {
		r.Category = DefaultCategory(r.Kind)
	}
	if r.ProjectID == "" {
		r.ProjectID = "default"
	}
	if r.TenantID == "" {
		r.TenantID = "default"
	}
	if r.Environment == "" {
		r.Environment = "default"
	}
	if r.DesiredState == "" {
		r.DesiredState = StateActive
	}
	if r.ActualState == "" {
		r.ActualState = StateUnknown
	}
	if r.CreatedBy == "" {
		r.CreatedBy = "admin"
	}

	op := &store.Operation{
		Type:    "create_resource",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Creating resource: %s (%s)", r.Name, r.Kind),
	}
	s.aud.CreateOperation(op)

	if err := s.rs.CreateResource(r); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return nil, err
	}

	s.aud.FinishOperation(op.ID, "success",
		fmt.Sprintf("Resource created: %s (%s)", r.ID, r.Name))

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "resource.create",
		TargetType: "resource",
		TargetID:   r.ID,
		Details:    fmt.Sprintf(`{"name":"%s","kind":"%s","env":"%s"}`, r.Name, r.Kind, r.Environment),
	})

	return r, nil
}

// GetResource retrieves a resource by ID.
func (s *Service) GetResource(id string) (*Resource, error) {
	if id == "" {
		return nil, fmt.Errorf("resource ID is required")
	}
	return s.rs.GetResource(id)
}

// ListResources returns resources matching the filter.
func (s *Service) ListResources(filter *Filter) ([]*Resource, error) {
	return s.rs.ListResources(filter)
}

// DeleteResource removes a resource with safety checks.
func (s *Service) DeleteResource(id string, force bool) error {
	if id == "" {
		return fmt.Errorf("resource ID is required")
	}

	r, err := s.rs.GetResource(id)
	if err != nil {
		return err
	}

	op := &store.Operation{
		Type:    "delete_resource",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Deleting resource: %s (%s)", r.Name, r.Kind),
	}
	s.aud.CreateOperation(op)

	if err := s.rs.DeleteResource(id, force); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return err
	}

	s.aud.FinishOperation(op.ID, "success",
		fmt.Sprintf("Resource deleted: %s", id))

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "resource.delete",
		TargetType: "resource",
		TargetID:   id,
		Details:    fmt.Sprintf(`{"name":"%s","kind":"%s","force":%v}`, r.Name, r.Kind, force),
	})

	return nil
}

// UpdateResource updates an existing resource's mutable fields.
func (s *Service) UpdateResource(r *Resource) error {
	if r.ID == "" {
		return fmt.Errorf("resource ID is required for update")
	}

	original, err := s.rs.GetResource(r.ID)
	if err != nil {
		return err
	}

	op := &store.Operation{
		Type:    "update_resource",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Updating resource: %s", original.Name),
	}
	s.aud.CreateOperation(op)

	if err := s.rs.UpdateResource(r); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return err
	}

	s.aud.FinishOperation(op.ID, "success",
		fmt.Sprintf("Resource updated: %s", r.ID))

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "resource.update",
		TargetType: "resource",
		TargetID:   r.ID,
		Details:    fmt.Sprintf(`{"name":"%s","kind":"%s"}`, original.Name, original.Kind),
	})

	return nil
}
