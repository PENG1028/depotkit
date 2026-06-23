package cmd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/depotly/depotly/pkg/config"
	"github.com/depotly/depotly/pkg/postgres"
	"github.com/spf13/cobra"
)

var pgMigrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Compare local migration files against the database schema_migrations table.`,
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

		fmt.Printf("Migrations directory: %s\n", cfg.Postgres.Migrations)
		fmt.Printf("Local migration files: %d found\n", len(local))
		fmt.Println()

		// Connect to database and get applied migrations
		pool, err := connectPgPool(cfg)
		if err != nil {
			PrintWarn("Cannot connect to database: %v", err)
			PrintInfo("Database is not reachable — showing local files only")
			printLocalMigrations(local)
			return
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
			PrintWarn("Dirty migration found: %s (%s)", dirty.Version, dirty.Name)
			PrintInfo("The database may be in an inconsistent state. Investigate before proceeding.")
			fmt.Println()
		}

		// Diff
		diff := postgres.DiffMigrations(local, records)

		fmt.Printf("Applied migrations: %d\n", len(diff.Applied))
		for _, r := range diff.Applied {
			dirtyMark := ""
			if r.Dirty {
				dirtyMark = " [DIRTY]"
			}
			fmt.Printf("  ✓ %s_%s%s\n", r.Version, r.Name, dirtyMark)
		}

		fmt.Println()
		fmt.Printf("Pending migrations: %d\n", len(diff.Pending))
		for _, m := range diff.Pending {
			fmt.Printf("  · %s_%s\n", m.Version, m.Name)
		}

		if len(diff.Mismatched) > 0 {
			fmt.Println()
			PrintWarn("Checksum mismatches found:")
			for _, m := range diff.Mismatched {
				fmt.Printf("  ⚠ %s: expected %s, got %s\n", m.Version, m.ExpectedHash, m.ActualHash)
			}
			PrintWarn("Refusing to proceed until mismatches are resolved.")
		}
	},
}

func printLocalMigrations(migrations []postgres.Migration) {
	for _, m := range migrations {
		fmt.Printf("  %s_%s\n", m.Version, m.Name)
	}
	fmt.Println()
	fmt.Printf("Run 'depotly pg migrate up' to apply pending migrations.\n")
}

func connectPgPool(cfg *config.Config) (*pgxpool.Pool, error) {
	return postgres.Connect(
		cfg.Services.Postgres.User,
		cfg.Services.Postgres.Password,
		"localhost",
		cfg.Services.Postgres.Port,
		cfg.Services.Postgres.Database,
	)
}

func init() {
	pgMigrateCmd.AddCommand(pgMigrateStatusCmd)
}
