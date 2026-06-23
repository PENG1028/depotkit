package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/redis"
	"github.com/spf13/cobra"
)

var redisPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping Redis server",
	Long:  `Send a PING command to Redis to verify connectivity.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Redis.Enabled {
			ExitError("Redis is not enabled in config")
		}

		result, err := redis.Ping("localhost", cfg.Services.Redis.Port)
		if err != nil {
			ExitError("Redis ping failed: %v", err)
		}

		fmt.Printf("Redis ping response: %s\n", result)
		PrintSuccess("Redis is reachable")
	},
}

func init() {
	redisCmd.AddCommand(redisPingCmd)
}
