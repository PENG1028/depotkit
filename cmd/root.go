package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/store"
	"github.com/spf13/cobra"
)

var cfgFile string

// Store management
var (
	globalStore    *store.DB
	globalStoreMu  sync.Mutex
	storePath      string // set by the first command that needs it
)

var rootCmd = &cobra.Command{
	Use:   "depotly",
	Short: "Depotly - Docker-first local data service control tool",
	Long: `Depotly is a Docker-first local data service control tool
for small web projects and lightweight self-hosted development.

It supports PostgreSQL, Redis, MinIO/S3-compatible object storage, and MongoDB.`,
}

func init() {
	cobra.OnInitialize(func() {
		// Config file flag is defined here for commands that need it.
	})
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to depotly.yaml (default: ./depotly.yaml)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
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

// GetConfig loads the configuration.
// Priority: --config flag > depotly.yaml > datadock.yaml (legacy).
func GetConfig() *config.Config {
	path := cfgFile
	if path == "" {
		// Try new config name first, fall back to legacy
		if _, err := os.Stat("depotly.yaml"); err == nil {
			path = "depotly.yaml"
		} else if _, err := os.Stat("datadock.yaml"); err == nil {
			path = "datadock.yaml"
			PrintWarn("Legacy config detected: datadock.yaml")
			PrintInfo("Rename to depotly.yaml or run 'depotly init' to create a fresh config")
		} else {
			ExitError("No config found. Run 'depotly init' or create depotly.yaml")
		}
	}
	cfg, err := config.Load(path)
	if err != nil {
		ExitError("Failed to load config: %v", err)
	}
	return cfg
}

// defaultMetadataPath returns the default path for the metadata store.
func defaultMetadataPath() string {
	// Check configured work_dir first
	if cfgFile != "" {
		if cfg, err := config.Load(cfgFile); err == nil {
			workDir := cfg.Runtime.WorkDir
			if workDir == "" {
				workDir = ".depotly"
			}
			return filepath.Join(workDir, "metadata.db")
		}
	}
	// Try depotly.yaml in current dir
	if _, err := os.Stat("depotly.yaml"); err == nil {
		if cfg, err := config.Load("depotly.yaml"); err == nil {
			workDir := cfg.Runtime.WorkDir
			if workDir == "" {
				workDir = ".depotly"
			}
			return filepath.Join(workDir, "metadata.db")
		}
	}
	return filepath.Join(".depotly", "metadata.db")
}

// GetStore opens (or returns an existing) DBManager metadata store.
// The store is lazily initialized on first call.
func GetStore() *store.DB {
	globalStoreMu.Lock()
	defer globalStoreMu.Unlock()

	if globalStore != nil {
		return globalStore
	}

	path := storePath
	if path == "" {
		path = defaultMetadataPath()
	}

	db, err := store.Open(path)
	if err != nil {
		ExitError("Failed to open metadata store (%s): %v", path, err)
	}
	globalStore = db
	storePath = path
	return db
}

// CloseStore closes the global metadata store if open.
func CloseStore() {
	globalStoreMu.Lock()
	defer globalStoreMu.Unlock()
	if globalStore != nil {
		globalStore.Close()
		globalStore = nil
	}
}
