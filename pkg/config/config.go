package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the full storepilot.yaml / depotly.yaml configuration.
type Config struct {
	Project  string         `yaml:"project"`
	Runtime  RuntimeConfig  `yaml:"runtime"`
	Services ServiceConfig  `yaml:"services"`
	Postgres PostgresConfig `yaml:"postgres"`
	Mongo    MongoConfig    `yaml:"mongo"`
	Object   ObjectConfig   `yaml:"object"`
}

type RuntimeConfig struct {
	Mode        string `yaml:"mode"`
	ComposeFile string `yaml:"compose_file"`
	WorkDir     string `yaml:"work_dir,omitempty"` // data directory, defaults to .datadock
}

type ServiceConfig struct {
	Postgres PostgresService `yaml:"postgres"`
	Redis    RedisService    `yaml:"redis"`
	Object   ObjectService   `yaml:"object"`
	Mongo    MongoService    `yaml:"mongo"`
}

// --- Endpoint types ---

type EndpointConfig struct {
	Direct   DirectEndpoint `yaml:"direct"`
	Exposure ExposureConfig `yaml:"exposure"`
}

type DirectEndpoint struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

type ExposureConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Provider      string `yaml:"provider"`
	Protocol      string `yaml:"protocol,omitempty"`
	RouteName     string `yaml:"route_name,omitempty"`
	PublicHost    string `yaml:"public_host,omitempty"`
	PublicPort    int    `yaml:"public_port,omitempty"`
	InternalFirst bool   `yaml:"internal_first,omitempty"`
}

// DefaultEndpointConfig returns the default endpoint settings for a service.
func DefaultEndpointConfig() EndpointConfig {
	return EndpointConfig{
		Direct: DirectEndpoint{
			Enabled: true,
			Host:    "localhost",
		},
		Exposure: ExposureConfig{
			Enabled:       false,
			Provider:      "none",
			Protocol:      "tcp",
			RouteName:     "",
			PublicPort:    443,
			InternalFirst: true,
		},
	}
}

// --- Service types with Endpoint field ---

type PostgresService struct {
	Enabled       bool           `yaml:"enabled"`
	Image         string         `yaml:"image"`
	ContainerName string         `yaml:"container_name"`
	Port          int            `yaml:"port"`
	Database      string         `yaml:"database"`
	User          string         `yaml:"user"`
	Password      string         `yaml:"password"`
	Volume        string         `yaml:"volume"`
	Endpoint      EndpointConfig `yaml:"endpoint,omitempty"`
}

type RedisService struct {
	Enabled       bool           `yaml:"enabled"`
	Image         string         `yaml:"image"`
	ContainerName string         `yaml:"container_name"`
	Port          int            `yaml:"port"`
	Volume        string         `yaml:"volume"`
	Endpoint      EndpointConfig `yaml:"endpoint,omitempty"`
}

type ObjectService struct {
	Enabled       bool           `yaml:"enabled"`
	Provider      string         `yaml:"provider"`
	Image         string         `yaml:"image"`
	ContainerName string         `yaml:"container_name"`
	Port          int            `yaml:"port"`
	ConsolePort   int            `yaml:"console_port"`
	AccessKey     string         `yaml:"access_key"`
	SecretKey     string         `yaml:"secret_key"`
	Bucket        string         `yaml:"bucket"`
	Volume        string         `yaml:"volume"`
	Endpoint      EndpointConfig `yaml:"endpoint,omitempty"`
}

type MongoService struct {
	Enabled       bool           `yaml:"enabled"`
	Image         string         `yaml:"image"`
	ContainerName string         `yaml:"container_name"`
	Port          int            `yaml:"port"`
	Database      string         `yaml:"database"`
	Volume        string         `yaml:"volume"`
	Endpoint      EndpointConfig `yaml:"endpoint,omitempty"`
}

type PostgresConfig struct {
	Migrations string `yaml:"migrations"`
	Schema     string `yaml:"schema"`
	Backups    string `yaml:"backups"`
}

type MongoConfig struct {
	Backups string `yaml:"backups"`
}

type ObjectConfig struct {
	BackupPrefix string `yaml:"backup_prefix"`
}

// Load reads and parses storepilot.yaml / depotly.yaml from the given path,
// then applies default values for any missing fields.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply endpoint defaults for any service that has zero-value Endpoint.
	applyEndpointDefaults(&cfg)

	return &cfg, nil
}

// Save writes the config to the given path as YAML.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// applyEndpointDefaults fills in default endpoint values for services
// that were loaded from an older config without the endpoint field.
func applyEndpointDefaults(cfg *Config) {
	def := DefaultEndpointConfig()

	if !cfg.Services.Postgres.Enabled {
		// even if disabled, set defaults so show commands don't panic
	}
	// Detect zero-value (old config) by checking Direct.Enabled default
	applyDefaults(&cfg.Services.Postgres.Endpoint, def)
	applyDefaults(&cfg.Services.Redis.Endpoint, def)
	applyDefaults(&cfg.Services.Object.Endpoint, def)
	applyDefaults(&cfg.Services.Mongo.Endpoint, def)
}

func applyDefaults(ep *EndpointConfig, def EndpointConfig) {
	// If Direct.Host is empty, the config was loaded from an old file.
	// Apply full defaults.
	if ep.Direct.Host == "" {
		ep.Direct = def.Direct
	}
	if ep.Exposure.Provider == "" {
		ep.Exposure = def.Exposure
	}
	// If direct port is zero, set to default from the service port later
	// (post-load fixup in service-specific code).
}

// ApplyServicePortDefaults fills in direct endpoint port from service port,
// and sets direct host from service host if not already configured.
func ApplyServicePortDefaults(ep *EndpointConfig, servicePort int, serviceHost string) {
	if ep.Direct.Port == 0 {
		ep.Direct.Port = servicePort
	}
	if ep.Direct.Host == "" {
		ep.Direct.Host = serviceHost
	}
}

// PostgresEndpoint returns the effective endpoint config for PostgreSQL.
func (s *PostgresService) EffectiveEndpoint() EndpointConfig {
	ep := s.Endpoint
	ApplyServicePortDefaults(&ep, s.Port, "localhost")
	return ep
}

// RedisEndpoint returns the effective endpoint config for Redis.
func (s *RedisService) EffectiveEndpoint() EndpointConfig {
	ep := s.Endpoint
	ApplyServicePortDefaults(&ep, s.Port, "localhost")
	return ep
}

// ObjectEndpoint returns the effective endpoint config for Object storage.
func (s *ObjectService) EffectiveEndpoint() EndpointConfig {
	ep := s.Endpoint
	ApplyServicePortDefaults(&ep, s.Port, "localhost")
	return ep
}

// MongoEndpoint returns the effective endpoint config for MongoDB.
func (s *MongoService) EffectiveEndpoint() EndpointConfig {
	ep := s.Endpoint
	ApplyServicePortDefaults(&ep, s.Port, "localhost")
	return ep
}
