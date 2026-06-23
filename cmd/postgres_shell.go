package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var pgShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open an interactive PostgreSQL shell (psql)",
	Long:  `Connect to PostgreSQL using psql via Docker exec. Requires a running PostgreSQL container.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		dockerArgs := []string{
			"exec", "-it",
			cfg.Services.Postgres.ContainerName,
			"psql",
			"-U", cfg.Services.Postgres.User,
			"-d", cfg.Services.Postgres.Database,
		}

		c := exec.Command("docker", dockerArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		fmt.Printf("Connecting to PostgreSQL: %s\n", cfg.Services.Postgres.ContainerName)
		if err := c.Run(); err != nil {
			ExitError("Failed to open psql shell: %v", err)
		}
	},
}

func init() {
	pgCmd.AddCommand(pgShellCmd)
}
