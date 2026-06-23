package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/object"
	"github.com/spf13/cobra"
)

var objectPutTestCmd = &cobra.Command{
	Use:   "put-test",
	Short: "Upload a test object to object storage",
	Long:  `Upload a small test file to verify object storage write access.`,
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

		if err := object.PutTestObject(client, cfg.Services.Object.Bucket); err != nil {
			ExitError("Failed to upload test object: %v", err)
		}

		PrintSuccess("Test object uploaded to %s/%s", cfg.Services.Object.Bucket, ".datadock/test-object.txt")
	},
}

func init() {
	objectCmd.AddCommand(objectPutTestCmd)
}
