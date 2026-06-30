package cmd

import (
	"context"
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	mongoclient "github.com/depotly/depotly/pkg/mongo"
	pgclient "github.com/depotly/depotly/pkg/postgres"
	redisclient "github.com/depotly/depotly/pkg/redis"
	"github.com/depotly/depotly/pkg/resource"
	"github.com/spf13/cobra"
)

var resourceVerifyCmd = &cobra.Command{
	Use:   "verify <id>",
	Short: "Verify a resource is reachable",
	Long: `Test connectivity to a registered resource and update its actual state.

Connects to the database, pings it, detects Docker availability,
and updates the resource's actual_state to 'active' or 'unreachable'.

Examples:
  depotly resource verify <id>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db := GetStore()

		r, err := resource.NewService(db).GetResource(args[0])
		if err != nil {
			ExitError("%v", err)
		}

		fmt.Printf("Verifying:    %s (%s)\n", r.Name, r.KindLabel())
		fmt.Printf("Target:       %s:%d\n", r.DisplayHost(), r.Port)
		fmt.Println()

		reachable, errMsg := testConnectivity(r)

		if reachable {
			db.Exec("UPDATE resources SET actual_state = 'active', updated_at = datetime('now') WHERE id = ?", r.ID)
			PrintSuccess("Connectivity: reachable")
		} else {
			db.Exec("UPDATE resources SET actual_state = 'unreachable', updated_at = datetime('now') WHERE id = ?", r.ID)
			fmt.Printf("  ✗ Connectivity: %s\n", errMsg)
		}

		if ok, _ := docker.IsDockerAvailable(); ok {
			PrintInfo("Docker:       available")
		} else {
			PrintInfo("Docker:       not detected")
		}

		if reachable {
			PrintInfo("Actual state: active")
		} else {
			PrintWarn("Actual state: unreachable")
		}
	},
}

func testConnectivity(r *resource.Resource) (bool, string) {
	switch r.Kind {
	case resource.KindPostgres:
		if r.Username == "" || r.Password == "" || r.Host == "" || r.Port == 0 {
			return false, "incomplete connection info (need username, password, host, port)"
		}
		pool, err := pgclient.Connect(r.Username, r.Password, r.Host, r.Port, r.Database)
		if err != nil {
			return false, fmt.Sprintf("PostgreSQL connect failed: %v", err)
		}
		pool.Close()
		return true, ""

	case resource.KindRedis:
		if r.Host == "" || r.Port == 0 {
			return false, "incomplete connection info (need host, port)"
		}
		_, err := redisclient.Ping(r.Host, r.Port)
		if err != nil {
			return false, fmt.Sprintf("Redis ping failed: %v", err)
		}
		return true, ""

	case resource.KindMongo:
		if r.Host == "" || r.Port == 0 {
			return false, "incomplete connection info (need host, port)"
		}
		client, _, err := mongoclient.Connect(r.Host, r.Port, r.Database)
		if err != nil {
			return false, fmt.Sprintf("MongoDB connect failed: %v", err)
		}
		client.Disconnect(context.TODO())
		return true, ""

	case resource.KindSQLite:
		return true, ""

	case resource.KindExternalUnmanaged:
		return true, ""

	default:
		return false, fmt.Sprintf("unsupported resource kind: %s", r.Kind)
	}
}

func init() {
	resourceCmd.AddCommand(resourceVerifyCmd)
}
