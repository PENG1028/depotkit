package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/redis"
	"github.com/spf13/cobra"
)

var redisFlushPattern string

var redisFlushNamespaceCmd = &cobra.Command{
	Use:   "flush-namespace",
	Short: "Delete all keys matching a pattern (requires confirmation)",
	Long: `Delete all Redis keys matching a glob-style pattern.
This is a destructive operation and requires confirmation.

Example:
  depotly redis flush-namespace --pattern "cache:old:*"`,
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

		// Preview keys to be deleted
		keys, err := redis.ScanKeys(client, redisFlushPattern, 100)
		if err != nil {
			ExitError("Failed to scan keys: %v", err)
		}

		if len(keys) == 0 {
			PrintInfo("No keys match pattern: %s", redisFlushPattern)
			return
		}

		fmt.Printf("The following %d keys will be DELETED:\n", len(keys))
		for _, key := range keys {
			fmt.Printf("  %s\n", key)
		}
		fmt.Println()

		if !confirmFlushNamespace(redisFlushPattern, len(keys)) {
			PrintInfo("Flush cancelled")
			return
		}

		count, err := redis.FlushNamespace(client, redisFlushPattern)
		if err != nil {
			ExitError("Failed to flush namespace: %v", err)
		}

		PrintSuccess("Deleted %d keys matching '%s'", count, redisFlushPattern)
	},
}

func confirmFlushNamespace(pattern string, keyCount int) bool {
	fmt.Printf("⚠  Are you sure you want to delete %d keys matching '%s'?\n", keyCount, pattern)
	fmt.Printf("This action cannot be undone.\n")
	fmt.Printf("Type 'yes' to confirm: ")
	var input string
	fmt.Scanln(&input)
	return input == "yes"
}

func init() {
	redisCmd.AddCommand(redisFlushNamespaceCmd)
	redisFlushNamespaceCmd.Flags().StringVarP(&redisFlushPattern, "pattern", "p", "", "Key pattern to delete (glob-style)")
	redisFlushNamespaceCmd.MarkFlagRequired("pattern")
}
