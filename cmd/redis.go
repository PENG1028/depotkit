package cmd

import (
	"github.com/spf13/cobra"
)

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Redis management commands",
	Long:  `Manage Redis: status, ping, scan, ttl-check, and flush-namespace.`,
}

func init() {
	rootCmd.AddCommand(redisCmd)
}
