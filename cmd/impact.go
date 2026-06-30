package cmd

import (
	"fmt"

	"github.com/depotly/depotly/pkg/impact"
	"github.com/spf13/cobra"
)

var impactCmd = &cobra.Command{
	Use:   "impact",
	Short: "Analyze impact of resource changes",
	Long: `Assess what would be affected by resource operations.

The impact analysis shows which services depend on a resource,
what access endpoints are active, and calculates a risk level.

Commands:
  analyze    Analyze impact for a resource, service, or environment`,
}

var impactAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze impact of resource changes",
	Long: `Analyze what would be affected if a resource were deleted or modified.

Examples:
  depotly impact analyze --resource <id>
  depotly impact analyze --service my-app --env production
  depotly impact analyze --env production`,
	Run: func(cmd *cobra.Command, args []string) {
		svc := impact.New(GetStore())

		switch {
		case impResource != "":
			result, err := svc.AnalyzeResource(impResource)
			if err != nil {
				ExitError("%v", err)
			}
			printResourceImpact(result)

		case impService != "":
			results, err := svc.AnalyzeService(impService, impEnv)
			if err != nil {
				ExitError("%v", err)
			}
			if len(results) == 0 {
				fmt.Printf("No bindings found for service '%s'.\n", impService)
				return
			}
			printResourceList(results)

		case impEnv != "":
			results, err := svc.AnalyzeEnvironment(impEnv)
			if err != nil {
				ExitError("%v", err)
			}
			if len(results) == 0 {
				fmt.Printf("No resources found in environment '%s'.\n", impEnv)
				return
			}
			printResourceList(results)

		default:
			ExitError("specify --resource, --service, or --env")
		}
	},
}

func printResourceImpact(r *impact.AnalysisResult) {
	riskColor := "⚠"
	if r.RiskLevel == impact.RiskHigh {
		riskColor = "🚫"
	} else if r.RiskLevel == impact.RiskMedium {
		riskColor = "⚠"
	}

	fmt.Printf("Impact Analysis: %s (%s)\n", r.Resource.Name, r.Resource.Kind)
	fmt.Printf("  Environment:   %s\n", r.Resource.Environment)
	fmt.Printf("  Risk Level:    %s %s\n", riskColor, r.RiskLevel)
	fmt.Println()

	// Bindings
	if len(r.Bindings) > 0 {
		fmt.Printf("  Service Bindings (%d):\n", len(r.Bindings))
		for _, b := range r.Bindings {
			fmt.Printf("    - %s (%s, %s)\n", b.Service, b.Environment, b.EnvKey)
		}
	} else {
		fmt.Println("  Service Bindings: none")
	}
	fmt.Println()

	// Access endpoints
	if len(r.AccessPoints) > 0 {
		fmt.Printf("  Active Access Endpoints (%d):\n", len(r.AccessPoints))
		for _, e := range r.AccessPoints {
			exp := ""
			if e.ExpiresAt != "" {
				exp = fmt.Sprintf(" (expires: %s)", e.ExpiresAt)
			}
			fmt.Printf("    - %s → %s:%d%s\n", e.ID, e.TargetHost, e.TargetPort, exp)
		}
	} else {
		fmt.Println("  Active Access Endpoints: none")
	}
	fmt.Println()

	// Production warning
	if r.Resource.IsProduction {
		fmt.Println("  🚫  This is a PRODUCTION resource.")
	}

	// Suggestions
	if r.RiskLevel == impact.RiskHigh {
		fmt.Println()
		fmt.Println("  Recommendation: Remove or reassign all bindings before deletion.")
		fmt.Println("  Use 'depotly binding delete <id>' to unbind services first.")
	} else if r.RiskLevel == impact.RiskMedium {
		fmt.Println()
		fmt.Println("  Recommendation: Review bindings and endpoints before proceeding.")
	}
}

func printResourceList(results []*impact.AnalysisResult) {
	for _, r := range results {
		printResourceImpact(r)
		fmt.Println("---")
	}
}

var (
	impResource string
	impService  string
	impEnv      string
)

func init() {
	rootCmd.AddCommand(impactCmd)
	impactCmd.AddCommand(impactAnalyzeCmd)

	impactAnalyzeCmd.Flags().StringVar(&impResource, "resource", "", "Resource ID to analyze")
	impactAnalyzeCmd.Flags().StringVar(&impService, "service", "", "Service name to analyze")
	impactAnalyzeCmd.Flags().StringVar(&impEnv, "env", "", "Environment to analyze")
}
