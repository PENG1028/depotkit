package cmd

import (
	"fmt"
	"time"

	"github.com/depotly/depotly/pkg/object"
	"github.com/spf13/cobra"
)

var objectSignedURLKey string
var objectSignedURLExpiry int

var objectSignedURLCmd = &cobra.Command{
	Use:   "signed-url",
	Short: "Generate a signed URL for an object",
	Long: `Generate a time-limited signed URL for accessing a private object.

Examples:
  depotly object signed-url --key "uploads/a.png"
  depotly object signed-url --key "uploads/a.png" --expiry 3600`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()

		if !cfg.Services.Object.Enabled {
			ExitError("Object storage is not enabled in config")
		}

		if objectSignedURLKey == "" {
			ExitError("--key is required")
		}

		endpoint := fmt.Sprintf("localhost:%d", cfg.Services.Object.Port)
		client, err := object.Connect(endpoint, cfg.Services.Object.AccessKey, cfg.Services.Object.SecretKey, false)
		if err != nil {
			ExitError("Failed to connect to object storage: %v", err)
		}

		expiry := time.Duration(objectSignedURLExpiry) * time.Second
		url, err := object.SignedURL(client, cfg.Services.Object.Bucket, objectSignedURLKey, expiry)
		if err != nil {
			ExitError("Failed to generate signed URL: %v", err)
		}

		fmt.Printf("Signed URL (expires in %ds):\n", objectSignedURLExpiry)
		fmt.Println(url)
	},
}

func init() {
	objectCmd.AddCommand(objectSignedURLCmd)
	objectSignedURLCmd.Flags().StringVarP(&objectSignedURLKey, "key", "k", "", "Object key")
	objectSignedURLCmd.Flags().IntVarP(&objectSignedURLExpiry, "expiry", "e", 3600, "URL expiry in seconds")
}
