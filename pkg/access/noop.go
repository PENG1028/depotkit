package access

// NoopPublisher implements AccessPublisher as a no-op for testing
// and for when no access publisher is configured.
type NoopPublisher struct{}

// Name returns "noop".
func (p *NoopPublisher) Name() string { return "noop" }

// PublishHTTPRoute returns nil (no error) but also no endpoint.
func (p *NoopPublisher) PublishHTTPRoute(route HTTPRoute) (*AccessEndpoint, error) {
	return nil, nil
}

// PublishTempTCPAccess returns nil (no error) but also no endpoint.
func (p *NoopPublisher) PublishTempTCPAccess(access TempTCPAccess) (*AccessEndpoint, error) {
	return nil, nil
}

// RevokeAccess is a no-op.
func (p *NoopPublisher) RevokeAccess(endpointID string) error {
	return nil
}

// GetAccessStatus returns nil.
func (p *NoopPublisher) GetAccessStatus(endpointID string) (*AccessEndpoint, error) {
	return nil, nil
}
