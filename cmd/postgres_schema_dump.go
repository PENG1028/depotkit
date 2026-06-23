package cmd

import (
	"github.com/depotly/depotly/pkg/postgres"
	"github.com/spf13/cobra"
)

var pgSchemaDumpCmd = &cobra.Command{
	Use:   "schema-dump",
	Short: "Dump PostgreSQL schema (--schema-only)",
	Long:  `Run pg_dump --schema-only and save to the configured schema file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		if err := postgres.SchemaDump(
			cfg.Services.Postgres.User,
			cfg.Services.Postgres.Password,
			"localhost",
			cfg.Services.Postgres.Port,
			cfg.Services.Postgres.Database,
			cfg.Postgres.Schema,
		); err != nil {
			ExitError("Schema dump failed: %v", err)
		}

		PrintSuccess("Schema dumped to: %s", cfg.Postgres.Schema)
	},
}

func init() {
	pgCmd.AddCommand(pgSchemaDumpCmd)
}
