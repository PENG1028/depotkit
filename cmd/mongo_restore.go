package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/depotly/depotly/pkg/mongo"
	"github.com/spf13/cobra"
)

var mongoRestoreCmd = &cobra.Command{
	Use:   "restore [backup-dir]",
	Short: "Restore a MongoDB backup using mongorestore",
	Long: `Restore a MongoDB database from a mongodump backup.
If no backup directory is specified, lists available backups.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		backupDir := ""
		if len(args) > 0 {
			backupDir = args[0]
		} else {
			// List available backups
			entries, err := os.ReadDir(cfg.Mongo.Backups)
			if err != nil {
				ExitError("Failed to list backups: %v", err)
			}

			var backups []string
			for _, e := range entries {
				if e.IsDir() {
					backups = append(backups, e.Name())
				}
			}

			if len(backups) == 0 {
				ExitError("No backups found in %s", cfg.Mongo.Backups)
			}

			sort.Sort(sort.Reverse(sort.StringSlice(backups)))

			fmt.Println("Available backups:")
			for _, b := range backups {
				fmt.Printf("  %s\n", b)
			}
			fmt.Println()
			PrintInfo("Usage: depotly mongo restore <backup-directory>")
			return
		}

		fullPath := backupDir
		if !filepath.IsAbs(backupDir) {
			fullPath = filepath.Join(cfg.Mongo.Backups, backupDir)
		}

		fmt.Printf("Restoring from: %s\n", fullPath)
		fmt.Printf("Target database: %s\n", cfg.Services.Mongo.Database)

		fmt.Println()
		PrintWarn("This will overwrite the current database contents")
		if !confirmRestore(cfg.Services.Mongo.Database) {
			PrintInfo("Restore cancelled")
			return
		}

		if err := mongo.Restore("localhost", cfg.Services.Mongo.Port, cfg.Services.Mongo.Database, fullPath, true); err != nil {
			ExitError("MongoDB restore failed: %v", err)
		}

		PrintSuccess("MongoDB restore completed from: %s", fullPath)
	},
}

func init() {
	mongoCmd.AddCommand(mongoRestoreCmd)
}
