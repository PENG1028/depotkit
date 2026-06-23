package endpoint

// NoopExposureProvider is the default provider for disabled exposure.
// It generates disabled manifests and all Apply/Test calls are no-ops.
type NoopExposureProvider struct{}

func (p *NoopExposureProvider) Name() string { return "none" }

// GenerateManifest returns a manifest with exposure disabled.
// Credentials are always omitted in the manifest.
func (p *NoopExposureProvider) GenerateManifest(instance *InstanceInfo) (*ExposureManifest, error) {
	return &ExposureManifest{
		Version: "1",
		Kind:    "DatabaseEndpointExposure",
		Instance: InstanceMeta{
			Name: instance.Name,
			Type: instance.Type,
		},
		RealEndpoint: RealEndpoint{
			Protocol: ProtocolForType(instance.Type),
			Host:     instance.Host,
			Port:     instance.Port,
			Database: instance.Database,
		},
		Exposure: instance.Endpoint.Exposure,
		Credentials: &CredentialInfo{
			Omitted: true,
		},
		Metadata: ManifestMetadata{
			GeneratedBy: "depotly",
		},
	}, nil
}

func (p *NoopExposureProvider) Apply(manifest ExposureManifest) error {
	return ErrNotImplemented
}

func (p *NoopExposureProvider) Test(manifest ExposureManifest) error {
	return ErrNotImplemented
}
