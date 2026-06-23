package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check Docker and prerequisites",
	Long:  `Run preflight checks: Docker availability, Docker Compose, file permissions, and port availability.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Depotly preflight checks...")
		fmt.Println()

		results := docker.RunDoctorChecks()
		hasFail := false

		for _, r := range results {
			statusSymbol := "✓"
			if r.Status == "fail" {
				statusSymbol = "✗"
				hasFail = true
			} else if r.Status == "warn" {
				statusSymbol = "⚠"
			}

			fmt.Printf(" %s %s\n", statusSymbol, r.Name)
			fmt.Printf("   %s\n", r.Detail)
			fmt.Println()
		}

		if hasFail {
			ExitError("Some checks failed. Please fix the issues above and run 'depotly doctor' again.")
		}

		PrintSuccess("All checks passed")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
