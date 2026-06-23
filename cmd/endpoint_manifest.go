package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/endpoint"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var endpointManifestCmd = &cobra.Command{
	Use:   "manifest <instance>",
	Short: "Generate exposure manifest for an instance",
	Long: `Generate an exposure manifest YAML for the specified database instance.
Output is written to stdout only — no files are modified.

If exposure is disabled, the manifest will reflect that status.

MVP v0.1: only PostgreSQL instances are supported for manifest generation.

Examples:
  storepilot endpoint manifest postgres`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := GetConfig()
		name := args[0]

		inst, err := endpoint.InstanceFromConfig(cfg, name)
		if err != nil {
			ExitError("%v", err)
		}

		// MVP: only PostgreSQL supports manifest generation
		if inst.Type != "postgres" {
			ExitError("Unsupported instance type '%s' for manifest generation in this version. Only PostgreSQL is supported.", inst.Type)
		}

		if !inst.Endpoint.Exposure.Enabled {
			fmt.Printf("# Exposure is disabled for instance '%s'.\n", inst.Name)
			fmt.Printf("# The manifest below reflects the current configuration.\n")
			fmt.Printf("# Run 'storepilot endpoint expose %s --provider aegis' to enable.\n", inst.Name)
			fmt.Println()
		}

		provider := endpoint.GetProvider(inst.Endpoint.Exposure.Provider)
		manifest, err := provider.GenerateManifest(inst)
		if err != nil {
			ExitError("Failed to generate manifest: %v", err)
		}

		// Provider already handles credential omission (no passwords in manifest)

		data, err := yaml.Marshal(manifest)
		if err != nil {
			ExitError("Failed to marshal manifest: %v", err)
		}

		fmt.Println(string(data))
	},
}

func init() {
	endpointCmd.AddCommand(endpointManifestCmd)
}
