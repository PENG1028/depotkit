package cmd

import (
	"github.com/spf13/cobra"
)

var pgCmd = &cobra.Command{
	Use:   "pg",
	Short: "PostgreSQL management commands",
	Long:  `Manage PostgreSQL: status, shell, backup, restore, schema-dump, and migrations.`,
}

func init() {
	rootCmd.AddCommand(pgCmd)
}
