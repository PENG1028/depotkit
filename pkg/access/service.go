package access

import (
	"fmt"
	"time"

	"github.com/depotly/depotly/pkg/store"
)

// Service provides business logic for access endpoint management.
type Service struct {
	as  *accessStore
	aud *store.DB
}

// NewService creates an access service backed by the given store.
func NewService(db *store.DB) *Service {
	return &Service{
		as:  newAccessStore(db),
		aud: db,
	}
}

// PublishHTTP creates an HTTP route access endpoint.
func (s *Service) PublishHTTP(route HTTPRoute) (*AccessEndpoint, error) {
	if route.ResourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}
	if route.Host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if route.Port == 0 {
		route.Port = 443
	}

	// Verify resource exists
	err := s.aud.QueryRow("SELECT id FROM resources WHERE id = ?", route.ResourceID).Scan(new(string))
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s", route.ResourceID)
	}

	e := &AccessEndpoint{
		ResourceID: route.ResourceID,
		Type:       TypeHTTPRoute,
		Host:       route.Host,
		Port:       route.Port,
		TargetHost: route.TargetHost,
		TargetPort: route.TargetPort,
		Status:     StatusActive,
		CreatedBy:  "admin",
	}

	// Record operation
	op := &store.Operation{
		Type:    "publish_http",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Publishing HTTP route: %s → %s:%d", route.Host, route.TargetHost, route.TargetPort),
	}
	s.aud.CreateOperation(op)

	if err := s.as.Create(e); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return nil, err
	}

	s.aud.FinishOperation(op.ID, "success", fmt.Sprintf("HTTP route published: %s", e.ID))

	// Use publisher to also generate manifest (v0.1: manifest-only)
	pub := GetPublisher("aegis")
	pub.PublishHTTPRoute(route)

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "access.publish_http",
		TargetType: "access_endpoint",
		TargetID:   e.ID,
		Details:    fmt.Sprintf(`{"resource":"%s","host":"%s","port":%d}`, route.ResourceID, route.Host, route.Port),
	})

	return e, nil
}

// PublishTempTCP creates a temporary TCP access endpoint.
func (s *Service) PublishTempTCP(access TempTCPAccess) (*AccessEndpoint, error) {
	if access.ResourceID == "" {
		return nil, fmt.Errorf("resource ID is required")
	}
	if access.TargetHost == "" {
		return nil, fmt.Errorf("target host is required")
	}
	if access.TargetPort == 0 {
		return nil, fmt.Errorf("target port is required")
	}
	if access.TTL <= 0 {
		access.TTL = 30 * time.Minute
	}

	// Verify resource exists
	err := s.aud.QueryRow("SELECT id FROM resources WHERE id = ?", access.ResourceID).Scan(new(string))
	if err != nil {
		return nil, fmt.Errorf("resource not found: %s", access.ResourceID)
	}

	expiresAt := time.Now().Add(access.TTL).UTC().Format(time.RFC3339)

	e := &AccessEndpoint{
		ResourceID: access.ResourceID,
		Type:       TypeTempTCP,
		Host:       fmt.Sprintf("dbtmp-%s.local", access.ResourceID),
		Port:       15000, // default port range start
		TargetHost: access.TargetHost,
		TargetPort: access.TargetPort,
		Status:     StatusActive,
		ExpiresAt:  expiresAt,
		CreatedBy:  "admin",
	}

	op := &store.Operation{
		Type:    "publish_temp_tcp",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Publishing temp TCP: %s → %s:%d (expires %s)", access.ResourceID, access.TargetHost, access.TargetPort, expiresAt),
	}
	s.aud.CreateOperation(op)

	if err := s.as.Create(e); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return nil, err
	}

	s.aud.FinishOperation(op.ID, "success", fmt.Sprintf("Temp TCP published: %s", e.ID))

	// Use publisher for manifest generation
	pub := GetPublisher("aegis")
	pub.PublishTempTCPAccess(access)

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "access.publish_temp_tcp",
		TargetType: "access_endpoint",
		TargetID:   e.ID,
		Details:    fmt.Sprintf(`{"resource":"%s","target":"%s:%d","ttl":"%s"}`, access.ResourceID, access.TargetHost, access.TargetPort, access.TTL.String()),
	})

	return e, nil
}

// Revoke sets an access endpoint's status to revoked.
func (s *Service) Revoke(id string) error {
	if id == "" {
		return fmt.Errorf("access endpoint ID is required")
	}

	e, err := s.as.Get(id)
	if err != nil {
		return err
	}

	op := &store.Operation{
		Type:    "revoke_access",
		Status:  "running",
		Actor:   "admin",
		Message: fmt.Sprintf("Revoking access: %s (%s)", e.ID, e.Type),
	}
	s.aud.CreateOperation(op)

	if err := s.as.UpdateStatus(id, StatusRevoked); err != nil {
		s.aud.FinishOperation(op.ID, "failed", err.Error())
		return err
	}

	s.aud.FinishOperation(op.ID, "success", fmt.Sprintf("Access revoked: %s", id))

	// Also try publisher revocation (v0.1: manifest-only, no-op)
	pub := GetPublisher("aegis")
	pub.RevokeAccess(id)

	s.aud.AppendAuditLog(&store.AuditLog{
		Actor:      "admin",
		Action:     "access.revoke",
		TargetType: "access_endpoint",
		TargetID:   id,
		Details:    fmt.Sprintf(`{"type":"%s","resource":"%s"}`, e.Type, e.ResourceID),
	})

	return nil
}

// GetEndpoint retrieves an access endpoint by ID.
func (s *Service) GetEndpoint(id string) (*AccessEndpoint, error) {
	if id == "" {
		return nil, fmt.Errorf("access endpoint ID is required")
	}
	return s.as.Get(id)
}

// ListEndpoints returns access endpoints matching the filter.
func (s *Service) ListEndpoints(filter *Filter) ([]*AccessEndpoint, error) {
	return s.as.List(filter)
}
