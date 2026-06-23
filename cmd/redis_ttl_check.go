package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/redis"
	"github.com/spf13/cobra"
)

var redisTTLPattern string

var redisTTLCheckCmd = &cobra.Command{
	Use:   "ttl-check",
	Short: "Check TTL for keys matching a pattern",
	Long: `Scan Redis keys matching a pattern and report keys without TTL.
Keys without TTL (no expiry) indicate potential issues.

Examples:
  depotly redis ttl-check --pattern "cache:*"
  depotly redis ttl-check --pattern "session:*"`,
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

		results, err := redis.CheckTTL(client, redisTTLPattern)
		if err != nil {
			ExitError("TTL check failed: %v", err)
		}

		if len(results) == 0 {
			PrintInfo("No keys match pattern: %s", redisTTLPattern)
			return
		}

		withTTL := 0
		withoutTTL := 0

		for _, r := range results {
			if r.HasTTL {
				withTTL++
			} else {
				withoutTTL++
			}
		}

		fmt.Printf("Keys matching '%s': %d\n", redisTTLPattern, len(results))
		fmt.Printf("  With TTL:    %d\n", withTTL)
		fmt.Printf("  Without TTL: %d\n", withoutTTL)
		fmt.Println()

		if withoutTTL > 0 {
			PrintWarn("Keys without TTL:")
			for _, r := range results {
				if !r.HasTTL {
					fmt.Printf("  %s\n", r.Key)
				}
			}
			fmt.Println()
			PrintInfo("Keys without TTL may accumulate indefinitely. Consider setting TTL or using versioned keys.")
		} else {
			PrintSuccess("All keys have TTL set")
		}
	},
}

func init() {
	redisCmd.AddCommand(redisTTLCheckCmd)
	redisTTLCheckCmd.Flags().StringVarP(&redisTTLPattern, "pattern", "p", "*", "Key pattern to check (glob-style)")
}
