package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var mongoStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show MongoDB container status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		status, err := docker.ContainerStatus(cfg.Services.Mongo.ContainerName)
		if err != nil {
			ExitError("Failed to get MongoDB status: %v", err)
		}

		if status == "missing" || status == "" {
			PrintWarn("MongoDB container '%s' is not running", cfg.Services.Mongo.ContainerName)
			return
		}

		health, _ := docker.HealthCheck(cfg.Services.Mongo.ContainerName)
		fmt.Printf("MongoDB: %s [%s]\n", status, health)
	},
}

func init() {
	mongoCmd.AddCommand(mongoStatusCmd)
}
