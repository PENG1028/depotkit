package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/object"
	"github.com/spf13/cobra"
)

var objectStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check object storage status",
	Long:  `Verify connectivity to the object storage (MinIO/S3) and check bucket status.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Object.Enabled {
			ExitError("Object storage is not enabled in config")
		}

		endpoint := fmt.Sprintf("localhost:%d", cfg.Services.Object.Port)
		result, err := object.CheckStatus(
			endpoint,
			cfg.Services.Object.AccessKey,
			cfg.Services.Object.SecretKey,
			cfg.Services.Object.Bucket,
		)
		if err != nil {
			ExitError("Failed to connect to object storage: %v", err)
		}

		fmt.Printf("Endpoint: %s\n", result.Endpoint)
		fmt.Printf("Bucket:   %s\n", result.Bucket)
		fmt.Printf("Bucket OK: %v\n", result.BucketOK)

		if result.BucketOK {
			PrintSuccess("Object storage is reachable and bucket is ready")
		} else {
			PrintWarn("Object storage is reachable but bucket check failed")
		}
	},
}

func init() {
	objectCmd.AddCommand(objectStatusCmd)
}
