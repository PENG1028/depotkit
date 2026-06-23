package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all services (preserves volumes)",
	Long:  `Stop all running data services. Docker volumes are preserved.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		PrintInfo("Stopping services...")
		output, err := docker.ComposeExec(cfg.Runtime.ComposeFile, "down")
		if err != nil {
			ExitError("Failed to stop services: %v", err)
		}

		if output != "" {
			fmt.Println(output)
		}
		PrintSuccess("Services stopped (volumes preserved)")
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
