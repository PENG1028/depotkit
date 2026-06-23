package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/docker"
	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
)

var endpointTestCmd = &cobra.Command{
	Use:   "test <instance>",
	Short: "Test endpoint connectivity for an instance",
	Long: `Test connectivity for a database instance.

Always tests the direct endpoint.
For PostgreSQL, executes connection check.
For other types, performs a connectivity check.

If exposure is enabled, the routed endpoint is NOT tested in this version.

Examples:
  storepilot endpoint test postgres
  storepilot endpoint test redis`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		inst, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		fmt.Printf("Testing endpoint for instance: %s (%s)\n", inst.Name, inst.Type)
		fmt.Printf("  Direct endpoint: %s:%d\n", inst.Host, inst.Port)
		fmt.Println()

		// === DIRECT ENDPOINT TEST ===
		fmt.Println("--- Direct Endpoint ---")
		directOK := testDirectEndpoint(cfg, inst)
		if directOK {
			PrintSuccess("Direct endpoint: reachable")
		} else {
			PrintWarn("Direct endpoint: unreachable")
		}
		fmt.Println()

		// === ROUTED ENDPOINT TEST ===
		fmt.Println("--- Routed Endpoint ---")
		if !inst.Endpoint.Exposure.Enabled {
			PrintInfo("Routed endpoint: exposure disabled — not tested")
		} else {
			provider := inst.Endpoint.Exposure.Provider
			switch provider {
			case "aegis":
				PrintInfo("Routed endpoint: exposure provider 'aegis' is manifest-only in this version")
				PrintInfo("Routed endpoint: test is not implemented")
			default:
				PrintInfo("Routed endpoint: provider '%s' is not supported for testing", provider)
			}
		}
		fmt.Println()

		// === SUMMARY ===
		fmt.Println("=== Summary ===")
		if directOK {
			fmt.Println("  [PASS] Direct endpoint")
		} else {
			fmt.Println("  [FAIL] Direct endpoint")
		}
		if !inst.Endpoint.Exposure.Enabled {
			fmt.Println("  [SKIP] Routed endpoint (exposure disabled)")
		} else {
			fmt.Println("  [SKIP] Routed endpoint (manifest-only, not tested)")
		}
		fmt.Println()
		fmt.Println("Note: This test covers direct connectivity only.")
		fmt.Println("Routed endpoint tests are not available in this version.")
	},
}

func testDirectEndpoint(cfg *config.Config, inst *endpoint.InstanceInfo) bool {
	switch inst.Type {
	case "postgres":
		return testPostgresDirect(cfg, inst)
	case "redis":
		return testRedisDirect(cfg)
	case "object":
		return testObjectDirect(cfg)
	case "mongo":
		return testMongoDirect(cfg)
	default:
		return false
	}
}

func testPostgresDirect(cfg *config.Config, inst *endpoint.InstanceInfo) bool {
	output, err := docker.DockerExec("exec", cfg.Services.Postgres.ContainerName,
		"pg_isready", "-U", inst.User, "-d", inst.Database)
	if err != nil {
		fmt.Printf("  PostgreSQL pg_isready: %v\n", err)
		output2, err2 := docker.DockerExec("exec", cfg.Services.Postgres.ContainerName,
			"psql", "-U", inst.User, "-d", inst.Database, "-c", "SELECT 1")
		if err2 != nil {
			fmt.Printf("  PostgreSQL SELECT 1: %v\n", err2)
			return false
		}
		fmt.Printf("  PostgreSQL SELECT 1: %s\n", output2)
		return true
	}
	fmt.Printf("  PostgreSQL: %s\n", output)
	return true
}

func testRedisDirect(cfg *config.Config) bool {
	output, err := docker.DockerExec("exec", cfg.Services.Redis.ContainerName,
		"redis-cli", "ping")
	if err != nil {
		fmt.Printf("  Redis PING: %v\n", err)
		return false
	}
	fmt.Printf("  Redis PING: %s\n", output)
	return true
}

func testObjectDirect(cfg *config.Config) bool {
	output, err := docker.DockerExec("exec", cfg.Services.Object.ContainerName,
		"curl", "-f", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"http://localhost:9000/minio/health/live")
	if err != nil {
		fmt.Printf("  MinIO health: %v\n", err)
		return false
	}
	fmt.Printf("  MinIO health: HTTP %s\n", output)
	return true
}

func testMongoDirect(cfg *config.Config) bool {
	output, err := docker.DockerExec("exec", cfg.Services.Mongo.ContainerName,
		"mongosh", "--quiet", "--eval", "db.runCommand({ping:1}).ok")
	if err != nil {
		fmt.Printf("  MongoDB ping: %v\n", err)
		return false
	}
	fmt.Printf("  MongoDB ping: %s\n", output)
	return true
}

func init() {
	endpointCmd.AddCommand(endpointTestCmd)
}
