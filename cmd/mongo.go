package cmd

import (
	"github.com/spf13/cobra"
)

var mongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "MongoDB management commands",
	Long:  `Manage MongoDB: status, shell, collections, backup, restore, and versions.`,
}

func init() {
	rootCmd.AddCommand(mongoCmd)
}
