package cmd

import (
	"fmt"
	"os"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
)

var endpointUnexposeCmd = &cobra.Command{
	Use:   "unexpose <instance>",
	Short: "Disable exposure for a database instance",
	Long: `Disable endpoint exposure for a database instance.

This updates the configuration file but does NOT:
  - Delete the database
  - Stop containers
  - Remove volumes
  - Delete backups
  - Delete migration state

Examples:
  depotly endpoint unexpose postgres`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		// Resolve instance (validates the name exists)
		_, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		// Update exposure to disabled
		switch name {
		case "postgres", "pg", "pg-dev":
			cfg.Services.Postgres.Endpoint.Exposure.Enabled = false
			cfg.Services.Postgres.Endpoint.Exposure.Provider = "none"
		case "redis":
			cfg.Services.Redis.Endpoint.Exposure.Enabled = false
			cfg.Services.Redis.Endpoint.Exposure.Provider = "none"
		case "object", "s3", "minio":
			cfg.Services.Object.Endpoint.Exposure.Enabled = false
			cfg.Services.Object.Endpoint.Exposure.Provider = "none"
		case "mongo", "mongodb":
			cfg.Services.Mongo.Endpoint.Exposure.Enabled = false
			cfg.Services.Mongo.Endpoint.Exposure.Provider = "none"
		default:
			ExitError("Unknown instance: %s", name)
		}

		// Save config
		configPath := cfgFile
		if configPath == "" {
			configPath = "depotly.yaml"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				configPath = "depotly.yaml"
			}
		}
		if err := config.Save(configPath, cfg); err != nil {
			ExitError("Failed to save config: %v", err)
		}

		PrintSuccess("Updated %s — exposure disabled for '%s'", configPath, name)

		// Warn about stale manifest
		exposureDir, resolveErr := resolveExposureDir(cfg)
		if resolveErr == nil {
			manifestPath := exposureDir + "/" + name + ".yaml"
			if _, err := os.Stat(manifestPath); err == nil {
				fmt.Println()
				PrintWarn("Manifest file may be stale and is no longer active: %s", manifestPath)
				PrintInfo("You can safely delete it: rm %s", manifestPath)
			}
		}

		fmt.Println()
		PrintInfo("Unexpose completed. The database instance is unchanged.")
		PrintInfo("Direct endpoint is still available via 'depotly endpoint direct %s'", name)
	},
}

func init() {
	endpointCmd.AddCommand(endpointUnexposeCmd)
}
