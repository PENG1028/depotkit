package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all services",
	Long:  `Display the running state and health of all enabled data services.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		fmt.Printf("Project: %s\n", cfg.Project)
		fmt.Println()

		// Use docker compose ps if compose file exists
		composeOutput, err := docker.PsStatus(cfg.Runtime.ComposeFile)
		if err == nil && composeOutput != "" {
			fmt.Println(composeOutput)
		} else {
			// Fallback: check each container individually
			checkServiceStatus(cfg.Services.Postgres.ContainerName, "PostgreSQL", cfg.Services.Postgres.Enabled)
			checkServiceStatus(cfg.Services.Redis.ContainerName, "Redis", cfg.Services.Redis.Enabled)
			checkServiceStatus(cfg.Services.Object.ContainerName, "MinIO", cfg.Services.Object.Enabled)
			checkServiceStatus(cfg.Services.Mongo.ContainerName, "MongoDB", cfg.Services.Mongo.Enabled)
		}
	},
}

func checkServiceStatus(containerName, serviceName string, enabled bool) {
	if !enabled {
		fmt.Printf("  %s: disabled\n", serviceName)
		return
	}

	status, err := docker.ContainerStatus(containerName)
	if err != nil || status == "missing" || status == "" {
		fmt.Printf("  %s (%s): missing\n", serviceName, containerName)
		return
	}

	parts := strings.SplitN(status, "\t", 3)
	if len(parts) >= 2 {
		healthStr := ""
		health, err := docker.HealthCheck(containerName)
		if err == nil && health != "" && health != "no health check" {
			healthStr = fmt.Sprintf(" [health: %s]", health)
		}
		image := ""
		if len(parts) >= 3 {
			image = fmt.Sprintf(" (%s)", parts[2])
		}
		fmt.Printf("  %s%s: %s%s\n", serviceName, image, parts[1], healthStr)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
