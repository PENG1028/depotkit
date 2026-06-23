package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/mongo"
	"github.com/spf13/cobra"
)

var mongoVersionsCollection string

var mongoVersionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Show schemaVersion distribution in a collection",
	Long: `Inspect documents in a MongoDB collection and report schemaVersion distribution.

Example:
  depotly mongo versions --collection projects`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Mongo.Enabled {
			ExitError("MongoDB is not enabled in config")
		}

		if mongoVersionsCollection == "" {
			ExitError("--collection is required")
		}

		client, db, err := mongo.Connect("localhost", cfg.Services.Mongo.Port, cfg.Services.Mongo.Database)
		if err != nil {
			ExitError("Failed to connect to MongoDB: %v", err)
		}
		defer client.Disconnect(nil)

		// Get total document count
		total, err := mongo.CountDocuments(db, mongoVersionsCollection)
		if err != nil {
			ExitError("Failed to count documents: %v", err)
		}

		fmt.Printf("Collection: %s\n", mongoVersionsCollection)
		fmt.Printf("Total documents: %d\n", total)
		fmt.Println()

		// Get schemaVersion distribution
		versions, err := mongo.SchemaVersionReport(db, mongoVersionsCollection)
		if err != nil {
			// If aggregation fails (e.g., collection doesn't exist), report gracefully
			PrintWarn("Failed to get schemaVersion distribution: %v", err)
			return
		}

		trackedCount := int64(0)
		fmt.Println("schemaVersion:")
		for _, v := range versions {
			fmt.Printf("  %s: %d documents\n", v.Version, v.Count)
			trackedCount += v.Count
		}

		missing := total - trackedCount
		if missing > 0 {
			fmt.Printf("  missing: %d documents\n", missing)
		}

		fmt.Println()
		PrintInfo("Recommended: Use schemaVersion in documents for migration tracking")
	},
}

func init() {
	mongoCmd.AddCommand(mongoVersionsCmd)
	mongoVersionsCmd.Flags().StringVarP(&mongoVersionsCollection, "collection", "c", "", "Collection name to inspect")
	mongoVersionsCmd.MarkFlagRequired("collection")
}
