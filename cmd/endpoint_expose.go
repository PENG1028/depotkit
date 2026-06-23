package cmd

import (
	"os"
	"path/filepath"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var exposeProvider string

var endpointExposeCmd = &cobra.Command{
	Use:   "expose <instance> --provider aegis",
	Short: "Enable exposure for a database instance",
	Long: `Enable endpoint exposure for a database instance.

This updates the configuration file and generates an exposure manifest.
In this version, exposure is manifest-only — no Aegis API is called.

Examples:
  depotly endpoint expose postgres --provider aegis

Supported providers: aegis`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		// Resolve instance and get current endpoint
		inst, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		// Validate provider
		if exposeProvider != "aegis" {
			ExitError("Unsupported provider: %s (supported: aegis)", exposeProvider)
		}

		// MVP: only PostgreSQL
		if inst.Type != "postgres" {
			ExitError("Unsupported instance type '%s' for endpoint exposure in this version", inst.Type)
		}

		// Update the config
		switch name {
		case "postgres", "pg", "pg-dev":
			cfg.Services.Postgres.Endpoint.Exposure.Enabled = true
			cfg.Services.Postgres.Endpoint.Exposure.Provider = "aegis"
			cfg.Services.Postgres.Endpoint.Exposure.Protocol = "tcp"
			cfg.Services.Postgres.Endpoint.Exposure.RouteName = name
			cfg.Services.Postgres.Endpoint.Exposure.PublicHost = ""
			cfg.Services.Postgres.Endpoint.Exposure.PublicPort = 443
			cfg.Services.Postgres.Endpoint.Exposure.InternalFirst = true
		}

		// Save config
		configPath := cfgFile
		if configPath == "" {
			configPath = "depotly.yaml"
			if _, statErr := os.Stat(configPath); statErr != nil {
				configPath = "depotly.yaml"
			}
		}
		if err := config.Save(configPath, cfg); err != nil {
			ExitError("Failed to save config: %v", err)
		}
		PrintSuccess("Updated %s", configPath)

		// Re-read instance to get updated endpoint
		inst, _ = endpoint.InstanceFromConfig(cfg, name)

		// Generate and write exposure manifest file
		exposureDir, resolveErr := resolveExposureDir(cfg)
		if resolveErr != nil {
			PrintWarn("%v", resolveErr)
		} else if mkErr := os.MkdirAll(exposureDir, 0755); mkErr != nil {
			PrintWarn("Failed to create exposure directory: %v", mkErr)
		} else {
			prov := endpoint.GetProvider(inst.Endpoint.Exposure.Provider)
			manifest, manErr := prov.GenerateManifest(inst)
			if manErr != nil {
				PrintWarn("Failed to generate manifest: %v", manErr)
			} else {
				manifestPath := filepath.Join(exposureDir, name+".yaml")
				data, _ := yaml.Marshal(manifest)
				if writeErr := os.WriteFile(manifestPath, data, 0644); writeErr != nil {
					PrintWarn("Failed to write manifest file: %v", writeErr)
				} else {
					PrintSuccess("Exposure manifest written: %s", manifestPath)
				}
			}
		}

		// Warning if public_host is empty
		if cfg.Services.Postgres.Endpoint.Exposure.PublicHost == "" {
			PrintWarn("public_host is empty; manifest is not directly usable by a router yet")
		}

		// Protected instance warning
		PrintWarn("This is a manifest-only exposure declaration.")
		PrintWarn("No route has been applied to any router.")
		PrintWarn("No proxy has been created.")
		PrintWarn("The routed endpoint is not reachable until Aegis applies this manifest.")

		PrintInfo("Exposure enabled for '%s' (provider: %s)", name, exposeProvider)
		PrintInfo("Run 'depotly endpoint manifest %s' to view the manifest", name)
		PrintInfo("Run 'depotly endpoint test %s' to verify direct connectivity", name)
	},
}

// resolveExposureDir determines the exposure directory using the configured work_dir.
// Defaults to .depotly if work_dir is not set.
// If both .depotly and .datadock exist, returns an error to prevent ambiguous writes.
func resolveExposureDir(cfg *config.Config) (string, error) {
	workDir := cfg.Runtime.WorkDir
	if workDir == "" {
		workDir = ".depotly"
	}

	// Always check for ambiguous dual-directory situation, regardless of workDir setting.
	// This catches cases where a user has both directories leftover from migration.
	if workDir == ".depotly" {
		if _, err := os.Stat(".datadock"); err == nil {
			// .datadoc exists alongside .depotly
			return "", NewSoftError("both .depotly and .datadock exist; set runtime.work_dir in config, or remove .datadock")
		}
	}

	return filepath.Join(workDir, "exposures"), nil
}

func init() {
	endpointCmd.AddCommand(endpointExposeCmd)
	endpointExposeCmd.Flags().StringVarP(&exposeProvider, "provider", "p", "", "Exposure provider (aegis)")
	endpointExposeCmd.MarkFlagRequired("provider")
}

// NewSoftError creates a non-fatal error for warnings.
func NewSoftError(msg string) error {
	return &softError{msg: msg}
}

type softError struct{ msg string }

func (e *softError) Error() string { return e.msg }
