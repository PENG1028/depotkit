package access

import (
	"fmt"
	"log"
)

// AegisManifestPublisher implements AccessPublisher for Aegis.
// In v0.1, this is manifest-only — it logs route declarations but does not
// call any Aegis API. Real Aegis integration will be implemented in v0.2+.
type AegisManifestPublisher struct{}

// Name returns "aegis".
func (p *AegisManifestPublisher) Name() string { return "aegis" }

// PublishHTTPRoute logs the route declaration.
// v0.1: manifest-only — no Aegis API call.
func (p *AegisManifestPublisher) PublishHTTPRoute(route HTTPRoute) (*AccessEndpoint, error) {
	log.Printf("[aegis-manifest] HTTP route: %s → %s:%d (resource: %s)",
		route.Host, route.TargetHost, route.TargetPort, route.ResourceID)
	return nil, fmt.Errorf("aegis API not implemented in v0.1; route declared in manifest")
}

// PublishTempTCPAccess logs the TCP access declaration.
// v0.1: manifest-only — no Aegis API call.
func (p *AegisManifestPublisher) PublishTempTCPAccess(access TempTCPAccess) (*AccessEndpoint, error) {
	log.Printf("[aegis-manifest] Temp TCP: %s → %s:%d (TTL: %s, resource: %s)",
		access.TargetHost, access.TargetHost, access.TargetPort, access.TTL, access.ResourceID)
	return nil, fmt.Errorf("aegis API not implemented in v0.1; temp access declared in manifest")
}

// RevokeAccess logs the revocation.
func (p *AegisManifestPublisher) RevokeAccess(endpointID string) error {
	log.Printf("[aegis-manifest] Revoke: %s", endpointID)
	return fmt.Errorf("aegis API not implemented in v0.1; revocation declared in manifest")
}

// GetAccessStatus returns a stub status.
func (p *AegisManifestPublisher) GetAccessStatus(endpointID string) (*AccessEndpoint, error) {
	return nil, fmt.Errorf("aegis API not implemented in v0.1")
}
