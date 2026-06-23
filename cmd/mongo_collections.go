package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/mongo"
	"github.com/spf13/cobra"
)

var mongoCollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "List MongoDB collections",
	Long:  `List all collections in the configured MongoDB database.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		client, db, err := mongo.Connect("localhost", cfg.Services.Mongo.Port, cfg.Services.Mongo.Database)
		if err != nil {
			ExitError("Failed to connect to MongoDB: %v", err)
		}
		defer client.Disconnect(nil)

		collections, err := mongo.ListCollections(db)
		if err != nil {
			ExitError("Failed to list collections: %v", err)
		}

		if len(collections) == 0 {
			PrintInfo("No collections found in database '%s'", cfg.Services.Mongo.Database)
			return
		}

		fmt.Printf("Collections in '%s':\n", cfg.Services.Mongo.Database)
		for _, name := range collections {
			fmt.Printf("  %s\n", name)
		}
		fmt.Println()
		PrintInfo("Total: %d collection(s)", len(collections))
	},
}

func init() {
	mongoCmd.AddCommand(mongoCollectionsCmd)
}
