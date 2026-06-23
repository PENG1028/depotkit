package endpoint

import (
	"fmt"

	"github.com/depotly/depotly/pkg/config"
)

// RealEndpoint describes the actual database endpoint.
type RealEndpoint struct {
	Protocol string `yaml:"protocol"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database,omitempty"`
}

// InstanceMeta describes the instance in the manifest.
type InstanceMeta struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// CredentialInfo describes how credentials are referenced in the manifest.
// For local managed instances, Omitted is true (no secret in manifest).
// For url_env instances, RefType/RefName record the environment variable.
type CredentialInfo struct {
	Omitted bool   `yaml:"omitted,omitempty"`
	RefType string `yaml:"ref_type,omitempty"`
	RefName string `yaml:"ref_name,omitempty"`
}

// ExposureManifest is the output of endpoint manifest.
type ExposureManifest struct {
	Version      string              `yaml:"version"`
	Kind         string              `yaml:"kind"`
	Instance     InstanceMeta        `yaml:"instance"`
	RealEndpoint RealEndpoint        `yaml:"real_endpoint"`
	Exposure     config.ExposureConfig `yaml:"exposure"`
	Credentials  *CredentialInfo     `yaml:"credentials,omitempty"`
	Metadata     ManifestMetadata    `yaml:"metadata"`
}

type ManifestMetadata struct {
	GeneratedBy string `yaml:"generated_by"`
}

// NewManifest creates an ExposureManifest for a given instance.
func NewManifest(name, instanceType, protocol, host string, port int, database string, exposure config.ExposureConfig) *ExposureManifest {
	return &ExposureManifest{
		Version: "1",
		Kind:    "DatabaseEndpointExposure",
		Instance: InstanceMeta{
			Name: name,
			Type: instanceType,
		},
		RealEndpoint: RealEndpoint{
			Protocol: protocol,
			Host:     host,
			Port:     port,
			Database: database,
		},
		Exposure: exposure,
		Metadata: ManifestMetadata{
			GeneratedBy: "storepilot",
		},
	}
}

// InstanceInfo holds runtime information about a database instance.
type InstanceInfo struct {
	Name     string
	Type     string // postgres, redis, object, mongo
	Host     string
	Port     int
	Database string
	User     string
	Password string
	Endpoint config.EndpointConfig
}

// ProtocolForType returns the protocol string for an instance type.
func ProtocolForType(instanceType string) string {
	switch instanceType {
	case "postgres":
		return "postgres"
	case "redis":
		return "redis"
	case "object", "s3", "minio":
		return "s3"
	case "mongo", "mongodb":
		return "mongodb"
	default:
		return "tcp"
	}
}

// InstanceFromConfig extracts an InstanceInfo from config by service name.
func InstanceFromConfig(cfg *config.Config, name string) (*InstanceInfo, error) {
	switch name {
	case "postgres", "pg", "pg-dev":
		s := cfg.Services.Postgres
		if !s.Enabled {
			return nil, fmt.Errorf("postgres is not enabled in config")
		}
		return &InstanceInfo{
			Name:     "postgres",
			Type:     "postgres",
			Host:     s.EffectiveEndpoint().Direct.Host,
			Port:     s.EffectiveEndpoint().Direct.Port,
			Database: s.Database,
			User:     s.User,
			Password: s.Password,
			Endpoint: s.EffectiveEndpoint(),
		}, nil
	case "redis":
		s := cfg.Services.Redis
		if !s.Enabled {
			return nil, fmt.Errorf("redis is not enabled in config")
		}
		return &InstanceInfo{
			Name:     "redis",
			Type:     "redis",
			Host:     s.EffectiveEndpoint().Direct.Host,
			Port:     s.EffectiveEndpoint().Direct.Port,
			Database: "",
			User:     "",
			Password: "",
			Endpoint: s.EffectiveEndpoint(),
		}, nil
	case "object", "s3", "minio":
		s := cfg.Services.Object
		if !s.Enabled {
			return nil, fmt.Errorf("object storage is not enabled in config")
		}
		return &InstanceInfo{
			Name:     "object",
			Type:     "object",
			Host:     s.EffectiveEndpoint().Direct.Host,
			Port:     s.EffectiveEndpoint().Direct.Port,
			Database: s.Bucket,
			User:     s.AccessKey,
			Password: s.SecretKey,
			Endpoint: s.EffectiveEndpoint(),
		}, nil
	case "mongo", "mongodb":
		s := cfg.Services.Mongo
		if !s.Enabled {
			return nil, fmt.Errorf("mongo is not enabled in config")
		}
		return &InstanceInfo{
			Name:     "mongo",
			Type:     "mongo",
			Host:     s.EffectiveEndpoint().Direct.Host,
			Port:     s.EffectiveEndpoint().Direct.Port,
			Database: s.Database,
			User:     "",
			Password: "",
			Endpoint: s.EffectiveEndpoint(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown instance: %s", name)
	}
}

// ListInstances returns all enabled service instances from config.
func ListInstances(cfg *config.Config) []*InstanceInfo {
	var instances []*InstanceInfo
	if cfg.Services.Postgres.Enabled {
		instances = append(instances, &InstanceInfo{
			Name:     "postgres",
			Type:     "postgres",
			Host:     cfg.Services.Postgres.EffectiveEndpoint().Direct.Host,
			Port:     cfg.Services.Postgres.EffectiveEndpoint().Direct.Port,
			Database: cfg.Services.Postgres.Database,
			User:     cfg.Services.Postgres.User,
			Password: cfg.Services.Postgres.Password,
			Endpoint: cfg.Services.Postgres.EffectiveEndpoint(),
		})
	}
	if cfg.Services.Redis.Enabled {
		instances = append(instances, &InstanceInfo{
			Name:     "redis",
			Type:     "redis",
			Host:     cfg.Services.Redis.EffectiveEndpoint().Direct.Host,
			Port:     cfg.Services.Redis.EffectiveEndpoint().Direct.Port,
			Endpoint: cfg.Services.Redis.EffectiveEndpoint(),
		})
	}
	if cfg.Services.Object.Enabled {
		instances = append(instances, &InstanceInfo{
			Name:     "object",
			Type:     "object",
			Host:     cfg.Services.Object.EffectiveEndpoint().Direct.Host,
			Port:     cfg.Services.Object.EffectiveEndpoint().Direct.Port,
			Database: cfg.Services.Object.Bucket,
			User:     cfg.Services.Object.AccessKey,
			Password: cfg.Services.Object.SecretKey,
			Endpoint: cfg.Services.Object.EffectiveEndpoint(),
		})
	}
	if cfg.Services.Mongo.Enabled {
		instances = append(instances, &InstanceInfo{
			Name:     "mongo",
			Type:     "mongo",
			Host:     cfg.Services.Mongo.EffectiveEndpoint().Direct.Host,
			Port:     cfg.Services.Mongo.EffectiveEndpoint().Direct.Port,
			Database: cfg.Services.Mongo.Database,
			Endpoint: cfg.Services.Mongo.EffectiveEndpoint(),
		})
	}
	return instances
}
