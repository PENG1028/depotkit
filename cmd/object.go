package cmd

import (
	"github.com/spf13/cobra"
)

var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "Object storage (MinIO/S3) management commands",
	Long:  `Manage object storage: status, put-test, list, signed-url, and clean-test.`,
}

func init() {
	rootCmd.AddCommand(objectCmd)
}
