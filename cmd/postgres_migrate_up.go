package cmd

import (
	"context"
	"fmt"

	"github.com/depotly/depotly/pkg/postgres"
	"github.com/spf13/cobra"
)

var pgMigrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply pending migrations",
	Long:  `Read all pending migration files and execute them in order against the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		// List local migration files
		local, err := postgres.ListMigrations(cfg.Postgres.Migrations)
		if err != nil {
			ExitError("Failed to list migrations: %v", err)
		}

		if len(local) == 0 {
			PrintWarn("No migration files found in %s", cfg.Postgres.Migrations)
			return
		}

		// Connect to database
		pool, err := connectPgPool(cfg)
		if err != nil {
			ExitError("Failed to connect to database: %v", err)
		}
		defer pool.Close()

		ctx := context.Background()

		// Ensure schema_migrations table exists
		if err := postgres.EnsureMigrationTable(ctx, pool); err != nil {
			ExitError("Failed to ensure schema_migrations table: %v", err)
		}

		// Get applied records
		records, err := postgres.GetAppliedMigrations(ctx, pool)
		if err != nil {
			ExitError("Failed to get applied migrations: %v", err)
		}

		// Check dirty
		dirty := postgres.CheckDirty(records)
		if dirty != nil {
			ExitError("Dirty migration found: %s (%s). Resolve the issue and manually clean the record before proceeding.",
				dirty.Version, dirty.Name)
		}

		// Diff
		diff := postgres.DiffMigrations(local, records)

		// Refuse if checksum mismatch
		if len(diff.Mismatched) > 0 {
			ExitError("Checksum mismatch detected for already-applied migrations. Resolve before proceeding.")
		}

		if len(diff.Pending) == 0 {
			PrintSuccess("All migrations are already applied")
			return
		}

		fmt.Printf("Applying %d pending migration(s):\n", len(diff.Pending))
		fmt.Println()

		for _, m := range diff.Pending {
			fmt.Printf("  → %s_%s... ", m.Version, m.Name)

			if err := postgres.ApplyMigration(ctx, pool, m); err != nil {
				fmt.Println("FAILED")
				fmt.Println()
				ExitError("Migration %s failed: %v\nDatabase may be in dirty state. Investigate before retrying.", m.Filename, err)
			}

			fmt.Println("OK")
		}

		fmt.Println()
		PrintSuccess("All pending migrations applied successfully")
	},
}

func init() {
	pgMigrateCmd.AddCommand(pgMigrateUpCmd)
}
