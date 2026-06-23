package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var mongoShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open an interactive MongoDB shell (mongosh)",
	Long:  `Connect to MongoDB using mongosh via Docker exec. Requires a running MongoDB container.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		dockerArgs := []string{
			"exec", "-it",
			cfg.Services.Mongo.ContainerName,
			"mongosh",
			cfg.Services.Mongo.Database,
		}

		c := exec.Command("docker", dockerArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		fmt.Printf("Connecting to MongoDB: %s/%s\n", cfg.Services.Mongo.ContainerName, cfg.Services.Mongo.Database)
		if err := c.Run(); err != nil {
			ExitError("Failed to open mongosh shell: %v", err)
		}
	},
}

func init() {
	mongoCmd.AddCommand(mongoShellCmd)
}
