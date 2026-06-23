package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
)

var endpointShowCmd = &cobra.Command{
	Use:   "show <instance>",
	Short: "Show endpoint status for a database instance",
	Long: `Display the endpoint configuration and exposure status for a database instance.

Examples:
  depotly endpoint show postgres
  depotly endpoint show redis
  depotly endpoint show object
  depotly endpoint show mongo`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		inst, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		maskPassword := func(s string) string {
			if s == "" {
				return "(none)"
			}
			if len(s) <= 3 {
				return strings.Repeat("*", len(s))
			}
			return s[:1] + strings.Repeat("*", len(s)-1)
		}

		fmt.Printf("Instance:     %s\n", inst.Name)
		fmt.Printf("Type:         %s\n", inst.Type)
		fmt.Println()

		// Direct endpoint
		ep := inst.Endpoint
		fmt.Printf("Direct endpoint:\n")
		fmt.Printf("  enabled:      %v\n", ep.Direct.Enabled)
		fmt.Printf("  host:         %s\n", ep.Direct.Host)
		fmt.Printf("  port:         %d\n", ep.Direct.Port)
		if inst.Database != "" {
			fmt.Printf("  database:     %s\n", inst.Database)
		}
		if inst.User != "" {
			fmt.Printf("  user:         %s\n", inst.User)
			fmt.Printf("  password:     %s\n", maskPassword(inst.Password))
		}
		fmt.Println()

		// Exposure
		fmt.Printf("Exposure:\n")
		if !ep.Exposure.Enabled {
			fmt.Printf("  status:       disabled\n")
		} else {
			fmt.Printf("  enabled:      %v\n", ep.Exposure.Enabled)
			fmt.Printf("  provider:     %s\n", ep.Exposure.Provider)
			fmt.Printf("  protocol:     %s\n", ep.Exposure.Protocol)
			fmt.Printf("  route_name:   %s\n", ep.Exposure.RouteName)
			fmt.Printf("  public_host:  %s\n", ep.Exposure.PublicHost)
			fmt.Printf("  public_port:  %d\n", ep.Exposure.PublicPort)
			fmt.Printf("  internal_first: %v\n", ep.Exposure.InternalFirst)
		}
	},
}

func init() {
	endpointCmd.AddCommand(endpointShowCmd)
}
