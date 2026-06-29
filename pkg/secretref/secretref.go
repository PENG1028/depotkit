// Package secretref defines the SecretRef type — a reference to a secret
// stored outside of DBManager's control plane.
//
// Format: secret://<project_id>/<environment>/<kind>/<name>[/<key>]
//
// Secrets are never stored in plaintext in DBManager records. The ref
// points to where the runner or operator can resolve the actual value.
package secretref

import (
	"fmt"
	"strings"
)

// Ref represents a parsed secret reference.
type Ref struct {
	Project     string
	Environment string
	Kind        string
	Name        string
	Key         string // optional sub-key (e.g. "DATABASE_URL", "password")
}

// String returns the canonical secret ref string.
func (r *Ref) String() string {
	s := fmt.Sprintf("secret://%s/%s/%s/%s", r.Project, r.Environment, r.Kind, r.Name)
	if r.Key != "" {
		s += "/" + r.Key
	}
	return s
}

// Parse parses a secret ref string into its components.
// Returns nil if the string is not a valid secret ref.
func Parse(ref string) (*Ref, error) {
	if !strings.HasPrefix(ref, "secret://") {
		return nil, fmt.Errorf("invalid secret ref: missing secret:// prefix")
	}

	trimmed := strings.TrimPrefix(ref, "secret://")
	parts := strings.Split(trimmed, "/")

	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid secret ref: expected at least 4 segments (project/env/kind/name), got %d", len(parts))
	}

	r := &Ref{
		Project:     parts[0],
		Environment: parts[1],
		Kind:        parts[2],
		Name:        strings.Join(parts[3:], "/"),
	}

	// If there are more than 4 parts, the last part is the key
	if len(parts) > 4 {
		r.Key = parts[len(parts)-1]
		r.Name = strings.Join(parts[3:len(parts)-1], "/")
	}

	if r.Project == "" || r.Environment == "" || r.Kind == "" || r.Name == "" {
		return nil, fmt.Errorf("invalid secret ref: all segments must be non-empty")
	}

	return r, nil
}

// MustParse parses a secret ref string or panics.
func MustParse(ref string) *Ref {
	r, err := Parse(ref)
	if err != nil {
		panic(err)
	}
	return r
}

// Validate returns true if the ref string is syntactically valid.
func Validate(ref string) bool {
	_, err := Parse(ref)
	return err == nil
}
