package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/depotly/depotly/pkg/postgres"
	"github.com/spf13/cobra"
)

var pgRestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore a PostgreSQL backup",
	Long: `Restore a PostgreSQL database from a dump file.
If no file is specified, lists available backups.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		backupFile := ""
		if len(args) > 0 {
			backupFile = args[0]
		} else {
			// List available backups
			entries, err := os.ReadDir(cfg.Postgres.Backups)
			if err != nil {
				ExitError("Failed to list backups: %v", err)
			}

			var backups []string
			for _, e := range entries {
				if !e.IsDir() && filepath.Ext(e.Name()) == ".dump" {
					backups = append(backups, e.Name())
				}
			}

			if len(backups) == 0 {
				ExitError("No backups found in %s", cfg.Postgres.Backups)
			}

			sort.Sort(sort.Reverse(sort.StringSlice(backups)))

			fmt.Println("Available backups:")
			for _, b := range backups {
				fmt.Printf("  %s\n", b)
			}
			fmt.Println()
			PrintInfo("Usage: depotly pg restore <filename>")
			fmt.Println("  Restores the latest backup by default if <filename> is 'latest'")
			return
		}

		if backupFile == "latest" {
			entries, _ := os.ReadDir(cfg.Postgres.Backups)
			var backups []string
			for _, e := range entries {
				if !e.IsDir() && filepath.Ext(e.Name()) == ".dump" {
					backups = append(backups, e.Name())
				}
			}
			if len(backups) == 0 {
				ExitError("No backups found")
			}
			sort.Sort(sort.Reverse(sort.StringSlice(backups)))
			backupFile = backups[0]
		}

		fullPath := backupFile
		if !filepath.IsAbs(backupFile) {
			fullPath = filepath.Join(cfg.Postgres.Backups, backupFile)
		}

		fmt.Printf("Restoring from: %s\n", fullPath)
		fmt.Printf("Target database: %s\n", cfg.Services.Postgres.Database)

		// Warn about data loss
		fmt.Println()
		PrintWarn("This will overwrite the current database contents")
		if !confirmRestore(cfg.Services.Postgres.Database) {
			PrintInfo("Restore cancelled")
			return
		}

		if err := postgres.Restore(
			cfg.Services.Postgres.User,
			cfg.Services.Postgres.Password,
			"localhost",
			cfg.Services.Postgres.Port,
			cfg.Services.Postgres.Database,
			fullPath,
			true, // drop existing
		); err != nil {
			ExitError("Restore failed: %v", err)
		}

		PrintSuccess("Restore completed from: %s", fullPath)
	},
}

func confirmRestore(database string) bool {
	fmt.Printf("Type the database name '%s' to confirm: ", database)
	var input string
	fmt.Scanln(&input)
	return input == database
}

func init() {
	pgCmd.AddCommand(pgRestoreCmd)
}
