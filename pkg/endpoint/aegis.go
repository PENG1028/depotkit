package endpoint

// AegisManifestProvider generates exposure manifests for the Aegis provider.
// It is manifest-only in this version — no API calls, no proxy, no routing.
type AegisManifestProvider struct{}

func (p *AegisManifestProvider) Name() string { return "aegis" }

// GenerateManifest produces a manifest with Aegis exposure configuration.
// The manifest does NOT contain database passwords.
// For local managed instances with plaintext passwords, credentials.omitted is set.
func (p *AegisManifestProvider) GenerateManifest(instance *InstanceInfo) (*ExposureManifest, error) {
	manifest := &ExposureManifest{
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
			GeneratedBy: "storepilot",
		},
	}

	// For future url_env support, set ref_type and ref_name instead of omitting.
	// Example:
	//    manifest.Credentials = &CredentialInfo{
	//        RefType: "env",
	//        RefName: "DATABASE_URL_PROD",
	//    }

	return manifest, nil
}

func (p *AegisManifestProvider) Apply(manifest ExposureManifest) error {
	return ErrNotImplemented
}

func (p *AegisManifestProvider) Test(manifest ExposureManifest) error {
	return ErrNotImplemented
}
