package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/redis"
	"github.com/spf13/cobra"
)

var redisScanPattern string

var redisScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Redis keys matching a pattern",
	Long: `Scan Redis for keys matching a glob-style pattern.

Examples:
  depotly redis scan --pattern "cache:*"
  depotly redis scan --pattern "user:*:profile:*"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Redis.Enabled {
			ExitError("Redis is not enabled in config")
		}

		client, err := redis.Connect("localhost", cfg.Services.Redis.Port)
		if err != nil {
			ExitError("Failed to connect to Redis: %v", err)
		}
		defer client.Close()

		keys, err := redis.ScanKeys(client, redisScanPattern, 100)
		if err != nil {
			ExitError("Redis scan failed: %v", err)
		}

		if len(keys) == 0 {
			PrintInfo("No keys match pattern: %s", redisScanPattern)
			return
		}

		fmt.Printf("Found %d keys matching '%s':\n", len(keys), redisScanPattern)
		fmt.Println()
		for _, key := range keys {
			fmt.Printf("  %s\n", key)
		}
	},
}

func init() {
	redisCmd.AddCommand(redisScanCmd)
	redisScanCmd.Flags().StringVarP(&redisScanPattern, "pattern", "p", "*", "Key pattern to scan (glob-style)")
	// --pattern is optional, defaults to "*"
}
