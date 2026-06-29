// Package binding defines the Resource Binding model — a connection between
// a service, an environment, and a resource, with an env key for runtime injection.
package binding

// Access role constants.
const (
	RoleReadWrite = "read_write"
	RoleReadOnly  = "read_only"
)

// Binding represents a service's use of a resource in a specific environment.
type Binding struct {
	ID          string `json:"id"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	ResourceID  string `json:"resource_id"`
	EnvKey      string `json:"env_key"`
	AccessRole  string `json:"access_role"`
	Required    bool   `json:"required"`
	ProjectID   string `json:"project_id"`
	TenantID    string `json:"tenant_id"`

	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
}

// Filter is used for listing bindings with optional constraints.
type Filter struct {
	Service     string
	Environment string
	ResourceID  string
}

// EnvEntry is the output format for the Runner injection API.
type EnvEntry struct {
	EnvKey    string `json:"env_key"`
	SecretRef string `json:"secret_ref"`
	Required  bool   `json:"required"`
}

// BindingSummary provides a compact view of a binding with resource info.
type BindingSummary struct {
	ID           string `json:"id"`
	Service      string `json:"service"`
	Environment  string `json:"environment"`
	EnvKey       string `json:"env_key"`
	AccessRole   string `json:"access_role"`
	Required     bool   `json:"required"`
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceKind string `json:"resource_kind"`
	ResourceEnv  string `json:"resource_env"`
}
