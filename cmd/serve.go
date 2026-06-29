package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/depotly/depotly/api"
	"github.com/spf13/cobra"
)

var serveAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start DBManager HTTP API server",
	Long: `Start the DBManager HTTP API server for Runner env injection queries.

The API provides:
  GET /api/v1/health           — Health check
  GET /api/v1/bindings         — Query service bindings (Runner injection)

Example:
  depotly serve --addr :8080
  curl http://localhost:8080/api/v1/bindings?service=my-app&env=production`,
	Run: func(cmd *cobra.Command, args []string) {
		db := GetStore()
		server := api.NewServer(db, serveAddr)

		PrintInfo("DBManager API server starting...")
		PrintInfo("Health:  http://localhost%s/api/v1/health", serveAddr)
		PrintInfo("Runner:  http://localhost%s/api/v1/bindings?service=<name>&env=<env>", serveAddr)
		fmt.Println()

		// Handle graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			fmt.Println()
			PrintInfo("Shutting down...")
			CloseStore()
			os.Exit(0)
		}()

		if err := server.ListenAndServe(); err != nil {
			ExitError("server error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "HTTP server address (e.g. :8080, 0.0.0.0:9090)")
}
