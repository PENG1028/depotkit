package cmd

import (
	"github.com/spf13/cobra"
)

var pgMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "PostgreSQL migration management",
	Long:  `Manage PostgreSQL schema migrations: status, up, down, and more.`,
}

func init() {
	pgCmd.AddCommand(pgMigrateCmd)
}
