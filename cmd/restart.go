package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all services",
	Long:  `Stop and then start all enabled data services.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		PrintInfo("Restarting services...")
		output, err := docker.ComposeExec(cfg.Runtime.ComposeFile, "restart")
		if err != nil {
			ExitError("Failed to restart services: %v", err)
		}

		if output != "" {
			fmt.Println(output)
		}
		PrintSuccess("Services restarted")
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
