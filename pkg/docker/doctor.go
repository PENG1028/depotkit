package docker

import (
	"fmt"
	"os"

	"github.com/depotly/depotly/pkg/utils"
)

// CheckResult holds the result of a single doctor check.
type CheckResult struct {
	Name   string
	Status string // "pass", "fail", "warn"
	Detail string
}

// RunDoctorChecks runs all preflight checks.
func RunDoctorChecks() []CheckResult {
	var results []CheckResult

	// Check Docker binary
	dockerOk, dockerMsg := IsDockerAvailable()
	if dockerOk {
		results = append(results, CheckResult{Name: "Docker", Status: "pass", Detail: dockerMsg})
	} else {
		results = append(results, CheckResult{Name: "Docker", Status: "fail", Detail: "Docker is not available. Please install Docker Desktop or Docker Engine.\n  See: https://docs.docker.com/engine/install/"})
	}

	// Check Docker Compose
	composeOk, composeMsg := IsDockerComposeAvailable()
	if composeOk {
		results = append(results, CheckResult{Name: "Docker Compose", Status: "pass", Detail: composeMsg})
	} else {
		results = append(results, CheckResult{Name: "Docker Compose", Status: "fail", Detail: "Docker Compose is not available as a docker plugin."})
	}

	// Check current directory writable
	cwd, _ := os.Getwd()
	if utils.IsWritable(cwd) {
		results = append(results, CheckResult{Name: "Write permission", Status: "pass", Detail: fmt.Sprintf("Directory is writable: %s", cwd)})
	} else {
		results = append(results, CheckResult{Name: "Write permission", Status: "fail", Detail: fmt.Sprintf("Directory is not writable: %s", cwd)})
	}

	// Check .datadock can be created (implicitly checked by writable test above)
	if utils.IsWritable(cwd) {
		results = append(results, CheckResult{Name: ".datadock directory", Status: "pass", Detail: ".datadock can be created"})
	} else {
		results = append(results, CheckResult{Name: ".datadock directory", Status: "fail", Detail: fmt.Sprintf("Cannot create .datadock in %s", cwd)})
	}

	// Check port availability
	ports := utils.RequiredPorts()
	occupied := utils.CheckPorts(ports)
	if len(occupied) == 0 {
		results = append(results, CheckResult{Name: "Port availability", Status: "pass", Detail: "All required ports are available"})
	} else {
		portList := ""
		for _, p := range occupied {
			portList += fmt.Sprintf("%d ", p)
		}
		results = append(results, CheckResult{Name: "Port availability", Status: "warn", Detail: fmt.Sprintf("Ports already in use: %s", portList)})
	}

	return results
}
