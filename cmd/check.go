package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify all services are reachable",
	Long:  `Connect to each enabled service and verify it is responding correctly.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		fmt.Printf("Checking services for project: %s\n", cfg.Project)
		fmt.Println()

		allOk := true

		if cfg.Services.Postgres.Enabled {
			ok := checkPostgresReachable(cfg)
			if !ok {
				allOk = false
			}
		}

		if cfg.Services.Redis.Enabled {
			ok := checkRedisReachable(cfg)
			if !ok {
				allOk = false
			}
		}

		if cfg.Services.Object.Enabled {
			ok := checkObjectReachable(cfg)
			if !ok {
				allOk = false
			}
		}

		if cfg.Services.Mongo.Enabled {
			ok := checkMongoReachable(cfg)
			if !ok {
				allOk = false
			}
		}

		fmt.Println()
		if allOk {
			PrintSuccess("All services are reachable")
		} else {
			ExitError("Some services are not reachable")
		}
	},
}

func checkPostgresReachable(cfg *config.Config) bool {
	// Use docker exec to check PostgreSQL connectivity
	output, err := docker.DockerExec("exec", cfg.Services.Postgres.ContainerName,
		"pg_isready", "-U", cfg.Services.Postgres.User, "-d", cfg.Services.Postgres.Database)
	if err != nil {
		PrintWarn("PostgreSQL: %v", err)
		return false
	}
	PrintSuccess("PostgreSQL: %s", output)
	return true
}

func checkRedisReachable(cfg *config.Config) bool {
	output, err := docker.DockerExec("exec", cfg.Services.Redis.ContainerName,
		"redis-cli", "ping")
	if err != nil {
		PrintWarn("Redis: %v", err)
		return false
	}
	PrintSuccess("Redis: %s", output)
	return true
}

func checkObjectReachable(cfg *config.Config) bool {
	// Use docker exec to check MinIO health
	output, err := docker.DockerExec("exec", cfg.Services.Object.ContainerName,
		"curl", "-f", "-s", "http://localhost:9000/minio/health/live")
	if err != nil {
		PrintWarn("MinIO: %v", err)
		return false
	}
	PrintSuccess("MinIO: HTTP health check passed")
	_ = output
	return true
}

func checkMongoReachable(cfg *config.Config) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use docker exec for MongoDB ping
	output, err := docker.DockerExec("exec", cfg.Services.Mongo.ContainerName,
		"mongosh", "--quiet", "--eval", "db.runCommand({ping:1}).ok")
	if err != nil {
		// Try mongosh with database
		output2, err2 := docker.DockerExec("exec", cfg.Services.Mongo.ContainerName,
			"mongosh", cfg.Services.Mongo.Database, "--quiet", "--eval", "db.runCommand({ping:1}).ok")
		if err2 != nil {
			PrintWarn("MongoDB: %v", err2)
			return false
		}
		PrintSuccess("MongoDB: ping ok (%s)", output2)
		return true
	}
	_ = ctx
	PrintSuccess("MongoDB: ping ok (%s)", output)
	return true
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
