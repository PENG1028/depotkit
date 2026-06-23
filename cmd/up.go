package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/docker"
	"github.com/depotly/depotly/pkg/utils"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start all services",
	Long:  `Generate docker-compose.yml and start all enabled data services.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		// Generate docker-compose.yml
		PrintInfo("Generating docker-compose configuration...")
		if err := docker.WriteComposeFile(cfg, cfg.Runtime.ComposeFile); err != nil {
			ExitError("Failed to generate docker-compose.yml: %v", err)
		}
		PrintSuccess("Generated %s", cfg.Runtime.ComposeFile)

		// Pre-check port conflicts before starting
		ports := collectServicePorts(cfg)
		occupied := utils.CheckPorts(ports)
		if len(occupied) > 0 {
			fmt.Println()
			PrintWarn("Port conflict detected. The following ports are already in use:")
			for _, p := range occupied {
				fmt.Printf("  Port %d\n", p)
			}
			fmt.Println()
			fmt.Println("Docker may fail to bind these ports. Stop the conflicting services first.")
			fmt.Println()
		}

		// Start services
		PrintInfo("Starting services...")
		output, err := docker.ComposeExec(cfg.Runtime.ComposeFile, "up", "-d")
		if err != nil {
			ExitError("Failed to start services: %v", err)
		}

		fmt.Println()
		fmt.Println(output)
		fmt.Println()
		PrintSuccess("Services started")
		fmt.Println()
		fmt.Println("Run 'depotly status' to check service states.")
		fmt.Println("Run 'depotly check' to verify connectivity.")
		fmt.Println("Run 'depotly connect' to get connection strings.")
	},
}

func collectServicePorts(cfg *config.Config) []int {
	var ports []int
	if cfg.Services.Postgres.Enabled {
		ports = append(ports, cfg.Services.Postgres.Port)
	}
	if cfg.Services.Redis.Enabled {
		ports = append(ports, cfg.Services.Redis.Port)
	}
	if cfg.Services.Object.Enabled {
		ports = append(ports, cfg.Services.Object.Port, cfg.Services.Object.ConsolePort)
	}
	if cfg.Services.Mongo.Enabled {
		ports = append(ports, cfg.Services.Mongo.Port)
	}
	return ports
}

func init() {
	rootCmd.AddCommand(upCmd)
}
