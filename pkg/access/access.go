// Package access provides the AccessPublisher abstraction and implementations
// for publishing database and service access endpoints through external routers.
//
// In v0.1, the Aegis publisher is manifest-only — it generates route declarations
// but does not call any Aegis API. This will be implemented in a later version.
package access

import "time"

// Type constants.
const (
	TypeHTTPRoute = "http_route"
	TypeTempTCP   = "temp_tcp"
)

// Status constants.
const (
	StatusActive  = "active"
	StatusExpired = "expired"
	StatusRevoked = "revoked"
)

// AccessEndpoint represents a published access endpoint.
type AccessEndpoint struct {
	ID         string `json:"id"`
	ResourceID string `json:"resource_id"`
	Type       string `json:"type"` // http_route, temp_tcp
	Host       string `json:"host"`
	Port       int    `json:"port"`
	TargetHost string `json:"target_host"`
	TargetPort int    `json:"target_port"`
	Status     string `json:"status"` // active, expired, revoked
	ExpiresAt  string `json:"expires_at,omitempty"`
	CreatedAt  string `json:"created_at"`
	CreatedBy  string `json:"created_by"`
}

// HTTPRoute describes an HTTP route to publish.
type HTTPRoute struct {
	ResourceID string `json:"resource_id"`
	Host       string `json:"host"`       // e.g. "pgadmin.example.com"
	Port       int    `json:"port"`       // e.g. 443
	TargetHost string `json:"target_host"` // internal service host
	TargetPort int    `json:"target_port"` // internal service port
	PathPrefix string `json:"path_prefix,omitempty"`
	TLS        bool   `json:"tls,omitempty"`
}

// TempTCPAccess describes a temporary TCP tunnel to publish.
type TempTCPAccess struct {
	ResourceID string        `json:"resource_id"`
	TargetHost string        `json:"target_host"`
	TargetPort int           `json:"target_port"`
	TTL        time.Duration `json:"ttl"` // e.g. 30 * time.Minute
	AllowedIPs []string      `json:"allowed_ips,omitempty"`
}

// AccessPublisher defines the interface for publishing access endpoints.
// In v0.1, only manifest generation is implemented — Apply and Test are stubs.
type AccessPublisher interface {
	// Name returns the provider identifier (e.g. "aegis", "noop").
	Name() string

	// PublishHTTPRoute publishes an HTTP route and returns the endpoint descriptor.
	PublishHTTPRoute(route HTTPRoute) (*AccessEndpoint, error)

	// PublishTempTCPAccess publishes a temporary TCP access and returns the endpoint descriptor.
	PublishTempTCPAccess(access TempTCPAccess) (*AccessEndpoint, error)

	// RevokeAccess revokes a previously published endpoint.
	RevokeAccess(endpointID string) error

	// GetAccessStatus returns the current status of an endpoint.
	GetAccessStatus(endpointID string) (*AccessEndpoint, error)
}

// GetPublisher returns the AccessPublisher for the given provider name.
func GetPublisher(name string) AccessPublisher {
	switch name {
	case "aegis":
		return &AegisManifestPublisher{}
	case "noop", "":
		return &NoopPublisher{}
	default:
		return &NoopPublisher{}
	}
}

// Filter is used for listing access endpoints with optional constraints.
type Filter struct {
	ResourceID string
	Type       string
	Status     string
}
