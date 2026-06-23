package cmd

import (
	"github.com/spf13/cobra"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Manage database endpoint exposure",
	Long: `Manage endpoint exposure for database instances.

StorePilot manages direct endpoints for local database instances.
Exposure is an optional declaration layer for future routed access.

Commands:
  show       Show endpoint status for an instance
  direct     Print direct connection information
  manifest   Generate exposure manifest (stdout only)
  expose     Enable exposure and generate manifest file
  unexpose   Disable exposure for an instance
  test       Test direct endpoint connectivity`,
}

func init() {
	rootCmd.AddCommand(endpointCmd)
}
