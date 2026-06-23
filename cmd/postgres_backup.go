package cmd

import (
	"github.com/depotly/depotly/pkg/postgres"
	"github.com/spf13/cobra"
)

var pgBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a PostgreSQL backup dump",
	Long:  `Run pg_dump -Fc and save a compressed backup to .depotly/backups/postgres/.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Postgres.Enabled {
			ExitError("PostgreSQL is not enabled in config")
		}

		path, err := postgres.Backup(
			cfg.Services.Postgres.User,
			cfg.Services.Postgres.Password,
			"localhost",
			cfg.Services.Postgres.Port,
			cfg.Services.Postgres.Database,
			cfg.Postgres.Backups,
		)
		if err != nil {
			ExitError("Backup failed: %v", err)
		}

		PrintSuccess("Backup created: %s", path)
	},
}

func init() {
	pgCmd.AddCommand(pgBackupCmd)
}
