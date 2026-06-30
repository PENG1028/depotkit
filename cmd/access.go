package cmd

import (
	"fmt"
	"time"

	"github.com/depotly/depotly/pkg/access"
	"github.com/spf13/cobra"
)

// separator line for tables
var dashLine = "---------------------------------------------------------------------------------------------------------------"

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Manage published access endpoints",
	Long: `Publish, list, and revoke access endpoints for resources.

Access endpoints represent routes published through a gateway (e.g. Aegis)
for HTTP services or temporary TCP database access.

In v0.1, publishing records the intent and persists the endpoint record,
but does not call any external gateway API (manifest-only).

Commands:
  list       List published access endpoints
  publish-http   Publish an HTTP route
  publish-tcp    Publish a temporary TCP access
  revoke     Revoke an access endpoint`,
}

var accessListCmd = &cobra.Command{
	Use:   "list",
	Short: "List access endpoints",
	Run: func(cmd *cobra.Command, args []string) {
		svc := access.NewService(GetStore())
		filter := &access.Filter{
			ResourceID: accFilterResource,
			Type:       accFilterType,
			Status:     accFilterStatus,
		}

		list, err := svc.ListEndpoints(filter)
		if err != nil {
			ExitError("list endpoints: %v", err)
		}

		if len(list) == 0 {
			fmt.Println("No access endpoints.")
			PrintInfo("Use 'depotly access publish-http' or 'depotly access publish-tcp' to create one.")
			return
		}

		fmt.Printf("%-28s %-14s %-12s %-32s %-8s %-8s\n",
			"ID", "Type", "Status", "Target", "Port", "Expires")
		fmt.Println(dashLine)
		for _, e := range list {
			target := e.TargetHost
			if target == "" {
				target = e.Host
			}
			exp := e.ExpiresAt
			if exp == "" {
				exp = "-"
			} else if len(exp) > 10 {
				exp = exp[:10]
			}
			fmt.Printf("%-28s %-14s %-12s %-32s %-8d %-8s\n",
				e.ID, e.Type, e.Status, target, e.TargetPort, exp)
		}
	},
}

var accessPublishHTTPCmd = &cobra.Command{
	Use:   "publish-http --resource <id> --host pgadmin.example.com",
	Short: "Publish an HTTP route",
	Long: `Publish an HTTP route to a resource through the access gateway.

Required:
  --resource    Resource ID to publish
  --host        Public hostname (e.g. pgadmin.example.com)

Optional:
  --target-host Internal target host (defaults to resource host)
  --target-port Internal target port (defaults to resource port)
  --port        Public port (default: 443)`,
	Run: func(cmd *cobra.Command, args []string) {
		if accPubResource == "" {
			ExitError("--resource is required")
		}
		if accPubHost == "" {
			ExitError("--host is required")
		}

		route := access.HTTPRoute{
			ResourceID: accPubResource,
			Host:       accPubHost,
			Port:       accPubPort,
			TargetHost: accPubTargetHost,
			TargetPort: accPubTargetPort,
		}

		svc := access.NewService(GetStore())
		endpoint, err := svc.PublishHTTP(route)
		if err != nil {
			ExitError("publish HTTP route: %v", err)
		}

		PrintSuccess("HTTP route published: %s", endpoint.ID)
		PrintInfo("%s → %s:%d", route.Host, endpoint.TargetHost, endpoint.TargetPort)
		if accPubTargetHost == "" {
			PrintWarn("No target-host specified; route is recorded but not applied to any gateway")
		}
		PrintWarn("v0.1: manifest-only — no Aegis API call has been made")
	},
}

var accessPublishTCPCmd = &cobra.Command{
	Use:   "publish-tcp --resource <id> --target-host 10.0.0.23 --target-port 5432",
	Short: "Publish a temporary TCP access",
	Long: `Publish a temporary TCP tunnel to a database resource.

Required:
  --resource      Resource ID to expose
  --target-host   Internal database host
  --target-port   Internal database port

Optional:
  --ttl           Time-to-live (e.g. 30m, 1h) (default: 30m)`,
	Run: func(cmd *cobra.Command, args []string) {
		if accTCPResource == "" {
			ExitError("--resource is required")
		}
		if accTCPTargetHost == "" {
			ExitError("--target-host is required")
		}
		if accTCPTargetPort == 0 {
			ExitError("--target-port is required")
		}

		ttl, err := time.ParseDuration(accTCPTTL)
		if err != nil {
			ExitError("invalid TTL: %s (use e.g. 30m, 1h)", accTCPTTL)
		}

		tcp := access.TempTCPAccess{
			ResourceID: accTCPResource,
			TargetHost: accTCPTargetHost,
			TargetPort: accTCPTargetPort,
			TTL:        ttl,
		}

		svc := access.NewService(GetStore())
		endpoint, err := svc.PublishTempTCP(tcp)
		if err != nil {
			ExitError("publish TCP access: %v", err)
		}

		PrintSuccess("TCP access published: %s", endpoint.ID)
		PrintInfo("%s → %s:%d (expires: %s)", endpoint.Host, tcp.TargetHost, tcp.TargetPort, endpoint.ExpiresAt)
		PrintWarn("v0.1: manifest-only — no Aegis tunnel has been created")
	},
}

var accessRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke an access endpoint",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := access.NewService(GetStore())
		if err := svc.Revoke(args[0]); err != nil {
			ExitError("%v", err)
		}
		PrintSuccess("Access endpoint revoked: %s", args[0])
	},
}

// access command flags
var (
	accFilterResource string
	accFilterType     string
	accFilterStatus   string

	accPubResource   string
	accPubHost       string
	accPubPort       int
	accPubTargetHost string
	accPubTargetPort int

	accTCPResource   string
	accTCPTargetHost string
	accTCPTargetPort int
	accTCPTTL        string
)

func init() {
	rootCmd.AddCommand(accessCmd)
	accessCmd.AddCommand(accessListCmd)
	accessCmd.AddCommand(accessPublishHTTPCmd)
	accessCmd.AddCommand(accessPublishTCPCmd)
	accessCmd.AddCommand(accessRevokeCmd)

	// list flags
	accessListCmd.Flags().StringVar(&accFilterResource, "resource", "", "Filter by resource ID")
	accessListCmd.Flags().StringVar(&accFilterType, "type", "", "Filter by type (http_route, temp_tcp)")
	accessListCmd.Flags().StringVar(&accFilterStatus, "status", "", "Filter by status (active, expired, revoked)")

	// publish-http flags
	accessPublishHTTPCmd.Flags().StringVar(&accPubResource, "resource", "", "Resource ID (required)")
	accessPublishHTTPCmd.Flags().StringVar(&accPubHost, "host", "", "Public hostname (required)")
	accessPublishHTTPCmd.Flags().IntVar(&accPubPort, "port", 443, "Public port")
	accessPublishHTTPCmd.Flags().StringVar(&accPubTargetHost, "target-host", "", "Internal target host")
	accessPublishHTTPCmd.Flags().IntVar(&accPubTargetPort, "target-port", 0, "Internal target port")

	// publish-tcp flags
	accessPublishTCPCmd.Flags().StringVar(&accTCPResource, "resource", "", "Resource ID (required)")
	accessPublishTCPCmd.Flags().StringVar(&accTCPTargetHost, "target-host", "", "Internal target host (required)")
	accessPublishTCPCmd.Flags().IntVar(&accTCPTargetPort, "target-port", 0, "Internal target port (required)")
	accessPublishTCPCmd.Flags().StringVar(&accTCPTTL, "ttl", "30m", "Time-to-live (e.g. 30m, 1h)")
}
