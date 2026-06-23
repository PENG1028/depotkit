package endpoint

import (
	"errors"
)

// ErrNotImplemented is returned when a provider method is not implemented.
var ErrNotImplemented = errors.New("not implemented in this version")

// ExposureProvider defines the interface for endpoint exposure providers.
type ExposureProvider interface {
	// Name returns the provider identifier (e.g. "none", "aegis").
	Name() string

	// GenerateManifest produces an exposure manifest for the given instance.
	GenerateManifest(instance *InstanceInfo) (*ExposureManifest, error)

	// Apply would apply the manifest to the provider. Not implemented in MVP.
	Apply(manifest ExposureManifest) error

	// Test would test the routed endpoint. Not implemented in MVP.
	Test(manifest ExposureManifest) error
}

// GetProvider returns the ExposureProvider for the given provider name.
func GetProvider(name string) ExposureProvider {
	switch name {
	case "aegis":
		return &AegisManifestProvider{}
	case "none", "":
		return &NoopExposureProvider{}
	default:
		return &NoopExposureProvider{}
	}
}
