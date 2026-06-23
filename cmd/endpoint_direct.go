package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
)

var endpointDirectShowSecret bool

var endpointDirectCmd = &cobra.Command{
	Use:   "direct <instance>",
	Short: "Print direct connection information",
	Long: `Print the real direct connection string for a database instance.
This command does not use exposure providers.

Examples:
  storepilot endpoint direct postgres
  storepilot endpoint direct redis
  storepilot endpoint direct object
  storepilot endpoint direct mongo`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		inst, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		if endpointDirectShowSecret {
			PrintWarn("--show-secret is enabled. Secrets will be printed in plain text to stdout.")
			PrintWarn("Ensure this output is not captured in logs or shared.")
			fmt.Println()
		}

		switch inst.Type {
		case "postgres":
			password := inst.Password
			if !endpointDirectShowSecret && password != "" {
				if len(password) <= 3 {
					password = "***"
				} else {
					password = password[:1] + "***"
				}
			}
			fmt.Printf("DATABASE_URL=postgres://%s:%s@%s:%d/%s\n",
				inst.User, password, inst.Host, inst.Port, inst.Database)
		case "redis":
			fmt.Printf("REDIS_URL=redis://%s:%d\n", inst.Host, inst.Port)
		case "object":
			fmt.Printf("S3_ENDPOINT=http://%s:%d\n", inst.Host, inst.Port)
			if inst.User != "" {
				secret := inst.Password
				if !endpointDirectShowSecret {
					secret = "***"
				}
				fmt.Printf("S3_ACCESS_KEY=%s\n", inst.User)
				fmt.Printf("S3_SECRET_KEY=%s\n", secret)
			}
			if inst.Database != "" {
				fmt.Printf("S3_BUCKET=%s\n", inst.Database)
			}
		case "mongo":
			fmt.Printf("MONGO_URL=mongodb://%s:%d/%s\n", inst.Host, inst.Port, inst.Database)
		default:
			fmt.Printf("HOST=%s\nPORT=%d\n", inst.Host, inst.Port)
		}
	},
}

func init() {
	endpointCmd.AddCommand(endpointDirectCmd)
	endpointDirectCmd.Flags().BoolVar(&endpointDirectShowSecret, "show-secret", false, "Show passwords in plain text")
}
