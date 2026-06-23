package cmd

import (
	"github.com/depotly/depotly/pkg/mongo"
	"github.com/spf13/cobra"
)

var mongoBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a MongoDB backup using mongodump",
	Long:  `Run mongodump and save a backup to .depotly/backups/mongo/.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		path, err := mongo.Backup("localhost", cfg.Services.Mongo.Port, cfg.Services.Mongo.Database, cfg.Mongo.Backups)
		if err != nil {
			ExitError("MongoDB backup failed: %v", err)
		}

		PrintSuccess("MongoDB backup created: %s", path)
	},
}

func init() {
	mongoCmd.AddCommand(mongoBackupCmd)
}
