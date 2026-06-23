package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/object"
	"github.com/spf13/cobra"
)

var objectCleanTestCmd = &cobra.Command{
	Use:   "clean-test",
	Short: "Remove test objects created by Depotly",
	Long:  `Delete all test objects (under .depotly/ prefix) from the bucket.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Object.Enabled {
			ExitError("Object storage is not enabled in config")
		}

		endpoint := fmt.Sprintf("localhost:%d", cfg.Services.Object.Port)
		client, err := object.Connect(endpoint, cfg.Services.Object.AccessKey, cfg.Services.Object.SecretKey, false)
		if err != nil {
			ExitError("Failed to connect to object storage: %v", err)
		}

		deleted, err := object.CleanTestObjects(client, cfg.Services.Object.Bucket)
		if err != nil {
			ExitError("Failed to clean test objects: %v", err)
		}

		if deleted == 0 {
			PrintInfo("No test objects found to clean")
		} else {
			PrintSuccess("Cleaned %d test object(s)", deleted)
		}
	},
}

func init() {
	objectCmd.AddCommand(objectCleanTestCmd)
}
