// Package resource defines the Resource model — the core abstraction in DBManager.
//
// A Resource represents a database, Redis, object store, or other data service
// instance that is registered in DBManager's resource control layer. It is
// separate from the Docker-level service configuration in depotly.yaml.
package resource

import "fmt"

// Kind constants for resource types.
const (
	KindPostgres          = "postgres"
	KindRedis             = "redis"
	KindSQLite            = "sqlite"
	KindExternalUnmanaged = "external_unmanaged"
	KindMongo             = "mongo"
	KindObject            = "object"
)

// Category constants.
const (
	CategoryRelational    = "relational"
	CategoryKV            = "kv"
	CategoryDocument      = "document"
	CategoryObjectStorage = "object_storage"
	CategoryQueue         = "queue"
)

// State constants.
const (
	StateActive   = "active"
	StateInactive = "inactive"
	StateUnknown  = "unknown"
	StateDeleted  = "deleted"
)

// DefaultCategory returns the default category for a given resource kind.
func DefaultCategory(kind string) string {
	switch kind {
	case KindPostgres:
		return CategoryRelational
	case KindRedis:
		return CategoryKV
	case KindSQLite:
		return CategoryRelational
	case KindMongo:
		return CategoryDocument
	case KindObject:
		return CategoryObjectStorage
	default:
		return CategoryRelational
	}
}

// Resource is the core domain model representing a data service resource.
type Resource struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`      // postgres, redis, sqlite, external_unmanaged, etc.
	Category  string `json:"category"`  // relational, kv, document, object_storage, queue
	Environment string `json:"environment"`
	ProjectID string `json:"project_id"`
	TenantID  string `json:"tenant_id"`

	// Human-readable identity
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Connection information (hidden from default views)
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Ownership and lifecycle
	OwnerService string `json:"owner_service,omitempty"`
	DesiredState string `json:"desired_state"`
	ActualState  string `json:"actual_state"`
	IsProduction bool   `json:"is_production"`
	IsTemporary  bool   `json:"is_temporary"`
	IsExternal   bool   `json:"is_external"`

	// Timestamps and attribution
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	CreatedBy string `json:"created_by"`
}

// Filter is used for listing resources with optional constraints.
type Filter struct {
	Kind         string
	Environment  string
	ProjectID    string
	OwnerService string
	DesiredState string
}

// SecretRef returns the secret reference string for this resource.
// Format: secret://<project_id>/<environment>/<kind>/<name>
func (r *Resource) SecretRef(key string) string {
	ref := "secret://" + r.ProjectID + "/" + r.Environment + "/" + r.Kind + "/" + r.Name
	if key != "" {
		ref += "/" + key
	}
	return ref
}

// DisplayHost returns the host with masking info if needed.
func (r *Resource) DisplayHost() string {
	if r.Host == "" {
		return "(not set)"
	}
	return r.Host
}

// DisplayPort returns the port as a string.
func (r *Resource) DisplayPort() string {
	if r.Port == 0 {
		return "(not set)"
	}
	return fmt.Sprintf("%d", r.Port)
}

// KindLabel returns a human-readable label for the kind.
func (r *Resource) KindLabel() string {
	switch r.Kind {
	case KindPostgres:
		return "PostgreSQL"
	case KindRedis:
		return "Redis"
	case KindSQLite:
		return "SQLite"
	case KindExternalUnmanaged:
		return "External (unmanaged)"
	case KindMongo:
		return "MongoDB"
	case KindObject:
		return "Object Storage"
	default:
		return r.Kind
	}
}
