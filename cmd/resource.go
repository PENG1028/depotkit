package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/impact"
	"github.com/depotly/depotly/pkg/resource"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources (databases, Redis, storage)",
	Long: `Register, view, and manage data service resources.

A resource is any data service instance — PostgreSQL, Redis, SQLite,
external database, or object storage — that DBManager tracks as a
first-class entity in your service system.

Commands:
  list    List registered resources
  show    Show resource details
  create  Register a new resource
  delete  Remove a resource`,
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered resources",
	Run: func(cmd *cobra.Command, args []string) {
		svc := resource.NewService(GetStore())
		filter := &resource.Filter{
			Kind:         resFilterKind,
			Environment:  resFilterEnv,
			OwnerService: resFilterService,
		}

		resources, err := svc.ListResources(filter)
		if err != nil {
			ExitError("list resources: %v", err)
		}

		if len(resources) == 0 {
			fmt.Println("No resources registered.")
			PrintInfo("Use 'depotly resource create' to register one.")
			return
		}

		fmt.Printf("%-28s %-12s %-12s %-14s %-16s %-8s\n",
			"ID", "Kind", "Environment", "Service", "Name", "Status")
		fmt.Println(strings.Repeat("-", 100))
		for _, r := range resources {
			service := r.OwnerService
			if service == "" {
				service = "-"
			}
			status := r.ActualState
			if r.DesiredState == resource.StateActive && r.ActualState == resource.StateActive {
				status = "active"
			} else if r.DesiredState == resource.StateInactive {
				status = "inactive"
			}
			fmt.Printf("%-28s %-12s %-12s %-14s %-16s %-8s\n",
				r.ID, r.KindLabel(), r.Environment, service, r.Name, status)
		}
	},
}

var resourceShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show resource details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := resource.NewService(GetStore())
		r, err := svc.GetResource(args[0])
		if err != nil {
			ExitError("%v", err)
		}

		fmt.Printf("ID:            %s\n", r.ID)
		fmt.Printf("Name:          %s\n", r.Name)
		fmt.Printf("Kind:          %s\n", r.KindLabel())
		fmt.Printf("Category:      %s\n", r.Category)
		fmt.Printf("Environment:   %s\n", r.Environment)
		fmt.Printf("Project:       %s\n", r.ProjectID)
		fmt.Printf("Tenant:        %s\n", r.TenantID)
		if r.Description != "" {
			fmt.Printf("Description:   %s\n", r.Description)
		}
		fmt.Println()

		fmt.Printf("Host:          %s\n", r.DisplayHost())
		fmt.Printf("Port:          %s\n", r.DisplayPort())
		if r.Database != "" {
			fmt.Printf("Database:      %s\n", r.Database)
		}
		if r.Username != "" {
			fmt.Printf("Username:      %s\n", r.Username)
		}
		password := r.Password
		if password != "" {
			if resShowSecret {
				PrintWarn("--show-secret enabled")
			} else {
				password = maskPassword(password)
			}
			fmt.Printf("Password:      %s\n", password)
		}
		fmt.Println()

		fmt.Printf("Owner Service: %s\n", ifEmpty(r.OwnerService, "-"))
		fmt.Printf("Desired State: %s\n", r.DesiredState)
		fmt.Printf("Actual State:  %s\n", r.ActualState)
		fmt.Printf("Production:    %v\n", r.IsProduction)
		fmt.Printf("Temporary:     %v\n", r.IsTemporary)
		fmt.Printf("External:      %v\n", r.IsExternal)
		fmt.Println()

		fmt.Printf("Secret Ref:    %s\n", r.SecretRef(""))
		fmt.Println()

		fmt.Printf("Created:       %s\n", r.CreatedAt)
		fmt.Printf("Updated:       %s\n", r.UpdatedAt)
		fmt.Printf("Created By:    %s\n", r.CreatedBy)
	},
}

var resourceCreateCmd = &cobra.Command{
	Use:   "create --kind postgres --name my-db",
	Short: "Register a new resource",
	Long: `Register a new data service resource in DBManager.

Required:
  --name        Human-readable name for the resource
  --kind        Resource type (postgres, redis, sqlite, mongo, object)

Optional:
  --env         Environment (default: "default")
  --host        Connection host
  --port        Connection port
  --database    Database name
  --username    Connection username
  --password    Connection password
  --service     Owner service name
  --production  Mark as production resource
  --external    Mark as externally managed
  --desc        Description`,
	Run: func(cmd *cobra.Command, args []string) {
		if resCreateName == "" {
			ExitError("--name is required")
		}
		if resCreateKind == "" {
			ExitError("--kind is required (postgres, redis, sqlite, mongo, object, external_unmanaged)")
		}

		validKinds := map[string]bool{
			resource.KindPostgres:          true,
			resource.KindRedis:             true,
			resource.KindSQLite:            true,
			resource.KindExternalUnmanaged: true,
			resource.KindMongo:             true,
			resource.KindObject:            true,
		}
		if !validKinds[resCreateKind] {
			ExitError("invalid kind: %s (valid: postgres, redis, sqlite, mongo, object, external_unmanaged)", resCreateKind)
		}

		r := &resource.Resource{
			Kind:         resCreateKind,
			Category:     resource.DefaultCategory(resCreateKind),
			Environment:  resCreateEnv,
			ProjectID:    "default",
			TenantID:     "default",
			Name:         resCreateName,
			Description:  resCreateDesc,
			Host:         resCreateHost,
			Port:         resCreatePort,
			Database:     resCreateDB,
			Username:     resCreateUser,
			Password:     resCreatePass,
			OwnerService: resCreateService,
			DesiredState: resource.StateActive,
			ActualState:  resource.StateUnknown,
			IsProduction: resCreateProd,
			IsExternal:   resCreateExternal,
		}

		svc := resource.NewService(GetStore())
		created, err := svc.CreateResource(r)
		if err != nil {
			ExitError("create resource: %v", err)
		}

		PrintSuccess("Resource created: %s (%s)", created.ID, created.Name)
		PrintInfo("Secret ref: %s", created.SecretRef(""))
	},
}

var resourceDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a resource",
	Long: `Remove a resource from DBManager.

Shows detailed impact analysis before deletion — which services are
bound, which access endpoints are active, and the risk level.

Use --force to bypass safety checks.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := resource.NewService(GetStore())

		if !resDeleteForce {
			analyzer := impact.New(GetStore())
			result, err := analyzer.AnalyzeResource(args[0])
			if err != nil {
				ExitError("%v", err)
			}

			fmt.Printf("Resource:     %s (%s)\n", result.Resource.Name, result.Resource.Kind)
			fmt.Printf("Environment:  %s\n", result.Resource.Environment)
			fmt.Printf("Risk Level:   %s\n", result.RiskLevel)
			fmt.Println()

			if len(result.Bindings) > 0 {
				fmt.Printf("  Service Bindings (%d):\n", len(result.Bindings))
				for _, b := range result.Bindings {
					fmt.Printf("    - %s (%s, %s)\n", b.Service, b.Environment, b.EnvKey)
				}
				fmt.Println()
			}

			if len(result.AccessPoints) > 0 {
				fmt.Printf("  Active Access (%d):\n", len(result.AccessPoints))
				for _, e := range result.AccessPoints {
					exp := ""
					if e.ExpiresAt != "" {
						exp = fmt.Sprintf(" (expires: %s)", e.ExpiresAt)
					}
					fmt.Printf("    - %s → %s:%d%s\n", e.ID, e.TargetHost, e.TargetPort, exp)
				}
				fmt.Println()
			}

			if result.Resource.IsProduction {
				fmt.Println("  🚫  This is a PRODUCTION resource.")
				fmt.Println()
			}

			if len(result.Bindings) > 0 || len(result.AccessPoints) > 0 {
				fmt.Println("Use --force to delete anyway.")
				return
			}
		}

		if err := svc.DeleteResource(args[0], resDeleteForce); err != nil {
			ExitError("%v", err)
		}
		PrintSuccess("Resource deleted: %s", args[0])
	},
}

var (
	resFilterKind    string
	resFilterEnv     string
	resFilterService string
	resShowSecret    bool
	resCreateName    string
	resCreateKind    string
	resCreateEnv     string
	resCreateDesc    string
	resCreateHost    string
	resCreatePort    int
	resCreateDB      string
	resCreateUser    string
	resCreatePass    string
	resCreateService string
	resCreateProd    bool
	resCreateExternal bool
	resDeleteForce   bool
)

func init() {
	rootCmd.AddCommand(resourceCmd)
	resourceCmd.AddCommand(resourceListCmd)
	resourceCmd.AddCommand(resourceShowCmd)
	resourceCmd.AddCommand(resourceCreateCmd)
	resourceCmd.AddCommand(resourceDeleteCmd)

	resourceListCmd.Flags().StringVar(&resFilterKind, "kind", "", "Filter by resource kind")
	resourceListCmd.Flags().StringVar(&resFilterEnv, "env", "", "Filter by environment")
	resourceListCmd.Flags().StringVar(&resFilterService, "service", "", "Filter by owner service")

	resourceShowCmd.Flags().BoolVar(&resShowSecret, "show-secret", false, "Show passwords in plain text")

	resourceCreateCmd.Flags().StringVar(&resCreateName, "name", "", "Resource name (required)")
	resourceCreateCmd.Flags().StringVar(&resCreateKind, "kind", "", "Resource kind (required): postgres, redis, sqlite, mongo, object, external_unmanaged")
	resourceCreateCmd.Flags().StringVar(&resCreateEnv, "env", "default", "Environment name")
	resourceCreateCmd.Flags().StringVar(&resCreateDesc, "desc", "", "Description")
	resourceCreateCmd.Flags().StringVar(&resCreateHost, "host", "", "Connection host")
	resourceCreateCmd.Flags().IntVar(&resCreatePort, "port", 0, "Connection port")
	resourceCreateCmd.Flags().StringVar(&resCreateDB, "database", "", "Database name")
	resourceCreateCmd.Flags().StringVar(&resCreateUser, "username", "", "Connection username")
	resourceCreateCmd.Flags().StringVar(&resCreatePass, "password", "", "Connection password")
	resourceCreateCmd.Flags().StringVar(&resCreateService, "service", "", "Owner service name")
	resourceCreateCmd.Flags().BoolVar(&resCreateProd, "production", false, "Mark as production resource")
	resourceCreateCmd.Flags().BoolVar(&resCreateExternal, "external", false, "Mark as externally managed")

	resourceDeleteCmd.Flags().BoolVar(&resDeleteForce, "force", false, "Force deletion even if bindings exist")
}

func maskPassword(s string) string {
	if s == "" {
		return "(none)"
	}
	if len(s) <= 3 {
		return strings.Repeat("*", len(s))
	}
	return s[:1] + strings.Repeat("*", len(s)-1)
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
