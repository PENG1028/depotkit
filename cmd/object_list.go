package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/object"
	"github.com/spf13/cobra"
)

var objectListPrefix string

var objectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects with a given prefix",
	Long: `List objects in the bucket matching a prefix.

Examples:
  depotly object list --prefix "uploads/"
  depotly object list --prefix "backups/"`,
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

		objects, err := object.ListObjects(client, cfg.Services.Object.Bucket, objectListPrefix)
		if err != nil {
			ExitError("Failed to list objects: %v", err)
		}

		if len(objects) == 0 {
			PrintInfo("No objects found with prefix '%s'", objectListPrefix)
			return
		}

		fmt.Printf("Objects in '%s' with prefix '%s':\n", cfg.Services.Object.Bucket, objectListPrefix)
		fmt.Println()
		for _, obj := range objects {
			fmt.Printf("  %s (%d bytes, %s)\n", obj.Key, obj.Size, obj.LastModified.Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
		PrintInfo("Total: %d objects", len(objects))
	},
}

func init() {
	objectCmd.AddCommand(objectListCmd)
	objectListCmd.Flags().StringVarP(&objectListPrefix, "prefix", "p", "", "Object key prefix to filter by")
}
