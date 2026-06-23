package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var redisStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Redis container status",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Redis.Enabled {
			ExitError("Redis is not enabled in config")
		}

		status, err := docker.ContainerStatus(cfg.Services.Redis.ContainerName)
		if err != nil {
			ExitError("Failed to get Redis status: %v", err)
		}

		if status == "missing" || status == "" {
			PrintWarn("Redis container '%s' is not running", cfg.Services.Redis.ContainerName)
			return
		}

		health, _ := docker.HealthCheck(cfg.Services.Redis.ContainerName)
		fmt.Printf("Redis: %s [%s]\n", status, health)
	},
}

func init() {
	redisCmd.AddCommand(redisStatusCmd)
}
