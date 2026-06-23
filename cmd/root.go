package cmd

import (
	"fmt"
	"os"

	"github.com/depotly/depotly/pkg/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "depotly",
	Short: "Depotly - Docker-first local data service control tool",
	Long: `Depotly is a Docker-first local data service control tool
for small web projects and lightweight self-hosted development.

It supports PostgreSQL, Redis, MinIO/S3-compatible object storage, and MongoDB.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(func() {
		// Config file flag is defined here for commands that need it.
	})
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to depotly.yaml (default: ./depotly.yaml)")
}

// ExitError prints an error message and exits with code 1.
func ExitError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// PrintSuccess prints a success message to stdout.
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// PrintInfo prints an informational message to stdout.
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf("ℹ "+format+"\n", args...)
}

// PrintWarn prints a warning message to stdout.
func PrintWarn(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}

// GetConfig loads the configuration from the default or specified path.
func GetConfig() *config.Config {
	path := cfgFile
	if path == "" {
		path = "depotly.yaml"
	}
	cfg, err := config.Load(path)
	if err != nil {
		ExitError("Failed to load config: %v", err)
	}
	return cfg
}
