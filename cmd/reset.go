package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/docker"
	"github.com/depotly/depotly/pkg/utils"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all services and delete volumes (requires confirmation)",
	Long:  `Stop all services and delete Docker volumes. WARNING: This deletes all data. Requires typing the project name to confirm.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		// Require project name confirmation
		if !utils.ConfirmProjectName(cfg.Project) {
			PrintInfo("Reset cancelled")
			return
		}

		// Stop services
		PrintInfo("Stopping services...")
		docker.ComposeExec(cfg.Runtime.ComposeFile, "down")

		// Remove volumes
		PrintInfo("Removing volumes...")
		volumes := getVolumeList(cfg)
		for _, v := range volumes {
			output, err := docker.DockerExec("volume", "rm", "-f", v)
			if err != nil {
				PrintWarn("Failed to remove volume %s: %v", v, err)
			} else if output != "" {
				fmt.Printf("  Removed volume: %s\n", output)
			}
		}

		PrintSuccess("Reset complete. All volumes have been removed.")
	},
}

func getVolumeList(cfg *config.Config) []string {
	var volumes []string
	if cfg.Services.Postgres.Enabled && cfg.Services.Postgres.Volume != "" {
		volumes = append(volumes, cfg.Services.Postgres.Volume)
	}
	if cfg.Services.Redis.Enabled && cfg.Services.Redis.Volume != "" {
		volumes = append(volumes, cfg.Services.Redis.Volume)
	}
	if cfg.Services.Object.Enabled && cfg.Services.Object.Volume != "" {
		volumes = append(volumes, cfg.Services.Object.Volume)
	}
	if cfg.Services.Mongo.Enabled && cfg.Services.Mongo.Volume != "" {
		volumes = append(volumes, cfg.Services.Mongo.Volume)
	}
	return volumes
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
