package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/binding"
	"github.com/spf13/cobra"
)

var bindingCmd = &cobra.Command{
	Use:   "binding",
	Short: "Manage service-to-resource bindings",
	Long: `Bind services to resources (databases, Redis, storage) with environment
isolation and env key mapping.

A binding connects a service to a resource in a specific environment,
defining which env key (e.g. DATABASE_URL) the Runner should inject.

Commands:
  list    List bindings
  create  Bind a service to a resource
  delete  Remove a binding`,
}

var bindingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bindings with resource info",
	Run: func(cmd *cobra.Command, args []string) {
		svc := binding.NewService(GetStore())
		filter := &binding.Filter{
			Service:     bndFilterService,
			Environment: bndFilterEnv,
			ResourceID:  bndFilterResource,
		}

		list, err := svc.ListBindings(filter)
		if err != nil {
			ExitError("list bindings: %v", err)
		}

		if len(list) == 0 {
			fmt.Println("No bindings found.")
			PrintInfo("Use 'depotly binding create' to bind a service to a resource.")
			return
		}

		fmt.Printf("%-28s %-16s %-12s %-22s %-22s %-10s %-8s\n",
			"ID", "Service", "Env", "Env Key", "Resource", "Role", "Req")
		fmt.Println(strings.Repeat("-", 130))
		for _, s := range list {
			resource := s.ResourceName
			if resource == "(deleted)" {
				resource = "⚠ (deleted)"
			}
			req := "yes"
			if !s.Required {
				req = "no"
			}
			fmt.Printf("%-28s %-16s %-12s %-22s %-22s %-10s %-8s\n",
				s.ID, s.Service, s.Environment, s.EnvKey, resource, s.AccessRole, req)
		}
	},
}

var bindingCreateCmd = &cobra.Command{
	Use:   "create --service my-app --resource <id> --env-key DATABASE_URL",
	Short: "Bind a service to a resource",
	Long: `Create a binding between a service and a resource.

Required:
  --service     Service name (e.g. my-app, blog-api)
  --resource    Resource ID to bind to
  --env-key     Environment variable key (e.g. DATABASE_URL, REDIS_URL)

Optional:
  --env         Environment (default: "default")
  --role        Access role: read_write (default) or read_only
  --no-require  Mark binding as non-required (default: required)`,
	Run: func(cmd *cobra.Command, args []string) {
		if bndCreateService == "" {
			ExitError("--service is required")
		}
		if bndCreateResource == "" {
			ExitError("--resource is required")
		}
		if bndCreateEnvKey == "" {
			ExitError("--env-key is required (e.g. DATABASE_URL)")
		}

		role := bndCreateRole
		if role == "" {
			role = binding.RoleReadWrite
		} else if role != binding.RoleReadWrite && role != binding.RoleReadOnly {
			ExitError("invalid role: %s (valid: read_write, read_only)", role)
		}

		b := &binding.Binding{
			Service:     bndCreateService,
			Environment: bndCreateEnv,
			ResourceID:  bndCreateResource,
			EnvKey:      bndCreateEnvKey,
			AccessRole:  role,
			Required:    !bndCreateNoRequire,
		}

		svc := binding.NewService(GetStore())
		created, err := svc.CreateBinding(b)
		if err != nil {
			ExitError("create binding: %v", err)
		}

		PrintSuccess("Binding created: %s", created.ID)
		PrintInfo("%s → %s (%s)", created.Service, created.ResourceID, created.EnvKey)
	},
}

var bindingDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a binding",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		svc := binding.NewService(GetStore())
		if err := svc.DeleteBinding(args[0]); err != nil {
			ExitError("%v", err)
		}
		PrintSuccess("Binding deleted: %s", args[0])
	},
}

// binding command flags
var (
	bndFilterService  string
	bndFilterEnv      string
	bndFilterResource string
	bndCreateService  string
	bndCreateResource string
	bndCreateEnvKey   string
	bndCreateEnv      string
	bndCreateRole     string
	bndCreateNoRequire bool
)

func init() {
	rootCmd.AddCommand(bindingCmd)
	bindingCmd.AddCommand(bindingListCmd)
	bindingCmd.AddCommand(bindingCreateCmd)
	bindingCmd.AddCommand(bindingDeleteCmd)

	// list flags
	bindingListCmd.Flags().StringVar(&bndFilterService, "service", "", "Filter by service name")
	bindingListCmd.Flags().StringVar(&bndFilterEnv, "env", "", "Filter by environment")
	bindingListCmd.Flags().StringVar(&bndFilterResource, "resource", "", "Filter by resource ID")

	// create flags
	bindingCreateCmd.Flags().StringVar(&bndCreateService, "service", "", "Service name (required)")
	bindingCreateCmd.Flags().StringVar(&bndCreateResource, "resource", "", "Resource ID (required)")
	bindingCreateCmd.Flags().StringVar(&bndCreateEnvKey, "env-key", "", "Environment variable key (required, e.g. DATABASE_URL)")
	bindingCreateCmd.Flags().StringVar(&bndCreateEnv, "env", "default", "Environment name")
	bindingCreateCmd.Flags().StringVar(&bndCreateRole, "role", "read_write", "Access role: read_write or read_only")
	bindingCreateCmd.Flags().BoolVar(&bndCreateNoRequire, "no-require", false, "Mark binding as non-required")
}
