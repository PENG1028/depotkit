package cmd

import (
	"fmt"
	"strings"

	"github.com/depotly/depotly/pkg/docker"
	"github.com/depotly/depotly/pkg/resource"
	"github.com/spf13/cobra"
)

// knownImages maps Docker image prefixes to resource kinds.
var knownImages = map[string]struct {
	kind     string
	category string
	defaultPort int
}{
	"postgres":    {resource.KindPostgres, resource.CategoryRelational, 5432},
	"redis":       {resource.KindRedis, resource.CategoryKV, 6379},
	"mongo":       {resource.KindMongo, resource.CategoryDocument, 27017},
	"minio":       {resource.KindObject, resource.CategoryObjectStorage, 9000},
}

// discoveredContainer holds info about a discovered data service container.
type discoveredContainer struct {
	Name      string
	Image     string
	Port      int
	Kind      string
	ContainerName string
}

var resourceDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover data services from Docker",
	Long: `Scan running Docker containers for known data services
(PostgreSQL, Redis, MongoDB, MinIO) and register them as resources.

Examples:
  depotly resource discover
  depotly resource discover --dry-run    # Show what would be registered`,
	Run: func(cmd *cobra.Command, args []string) {
		ok, info := docker.IsDockerAvailable()
		if !ok {
			ExitError("Docker not available: %s", info)
		}

		PrintInfo("Docker: %s", info)
		fmt.Println()

		containers, err := scanDockerContainers()
		if err != nil {
			ExitError("scan containers: %v", err)
		}

		if len(containers) == 0 {
			fmt.Println("No known data services found in running containers.")
			fmt.Println()
			PrintInfo("Supported images: postgres, redis, mongo, minio/minio")
			return
		}

		fmt.Printf("Found %d data service(s):\n", len(containers))
		fmt.Println()
		for _, c := range containers {
			fmt.Printf("  [%s]\n", c.Kind)
			fmt.Printf("    Container: %s\n", c.ContainerName)
			fmt.Printf("    Image:     %s\n", c.Image)
			fmt.Printf("    Port:      %d\n", c.Port)
			fmt.Println()
		}

		if discoverDryRun {
			fmt.Println("---")
			fmt.Println("Dry-run mode. Use without --dry-run to register.")
			return
		}

		// Auto-register each discovered container
		db := GetStore()
		for _, c := range containers {
			// Check if already registered by container name heuristic
			existing, _ := resource.NewService(db).ListResources(&resource.Filter{
				Kind: c.Kind,
			})
			skip := false
			for _, e := range existing {
				if strings.Contains(e.Name, c.ContainerName) || strings.Contains(e.Name, strings.Split(c.Image, ":")[0]) {
					skip = true
					break
				}
			}
			if skip {
				PrintInfo("Already registered: %s (%s)", c.Name, c.Kind)
				continue
			}

			r := &resource.Resource{
				Kind:        c.Kind,
				Category:    resource.DefaultCategory(c.Kind),
				Environment: "default",
				ProjectID:   "default",
				TenantID:    "default",
				Name:        c.Name,
				Host:        "localhost",
				Port:        c.Port,
				DesiredState: resource.StateActive,
				ActualState:  resource.StateUnknown,
				CreatedBy:   "discover",
			}

			created, err := resource.NewService(db).CreateResource(r)
			if err != nil {
				PrintWarn("Failed to register %s: %v", c.Name, err)
				continue
			}
			PrintSuccess("Registered: %s (%s, port %d)", created.Name, created.KindLabel(), created.Port)
		}
	},
}

// scanDockerContainers runs docker ps and parses the output for known images.
func scanDockerContainers() ([]discoveredContainer, error) {
	output, err := docker.DockerExec("ps", "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil, nil
	}

	var results []discoveredContainer
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		containerName := parts[0]
		image := parts[1]
		ports := ""
		if len(parts) >= 3 {
			ports = parts[2]
		}

		// Match against known images
		imageLower := strings.ToLower(image)
		for prefix, info := range knownImages {
			if strings.Contains(imageLower, prefix) {
				port := extractPort(ports, info.defaultPort)
				displayName := fmt.Sprintf("%s-%s", containerName, prefix)
				results = append(results, discoveredContainer{
					Name:          displayName,
					Image:         image,
					Port:          port,
					Kind:          info.kind,
					ContainerName: containerName,
				})
				break
			}
		}
	}

	return results, nil
}

// extractPort parses the Docker ports string to find the host port.
// Format examples: "0.0.0.0:5432->5432/tcp" or "5432/tcp" or ""
func extractPort(ports string, defaultPort int) int {
	if ports == "" {
		return defaultPort
	}
	// Try to find the host port in format like 0.0.0.0:5432->5432/tcp
	parts := strings.Split(ports, "->")
	if len(parts) >= 1 {
		hostPart := parts[0]
		// Extract the last number after ":"
		if idx := strings.LastIndex(hostPart, ":"); idx >= 0 {
			portStr := hostPart[idx+1:]
			var port int
			if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 {
				return port
			}
		}
	}
	return defaultPort
}

var discoverDryRun bool

func init() {
	resourceCmd.AddCommand(resourceDiscoverCmd)
	resourceDiscoverCmd.Flags().BoolVar(&discoverDryRun, "dry-run", false, "Show discovered containers without registering")
}
