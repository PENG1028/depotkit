package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var pgStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show PostgreSQL container status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		status, err := docker.ContainerStatus(cfg.Services.Postgres.ContainerName)
		if err != nil {
			ExitError("Failed to get PostgreSQL status: %v", err)
		}

		if status == "missing" || status == "" {
			PrintWarn("PostgreSQL container '%s' is not running", cfg.Services.Postgres.ContainerName)
			return
		}

		health, _ := docker.HealthCheck(cfg.Services.Postgres.ContainerName)
		fmt.Printf("PostgreSQL: %s [%s]\n", status, health)
	},
}

func init() {
	pgCmd.AddCommand(pgStatusCmd)
}
