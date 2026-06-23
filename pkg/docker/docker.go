package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// ComposeExec runs a docker compose command.
func ComposeExec(composeFile string, args ...string) (string, error) {
	cmdArgs := []string{"compose", "-f", composeFile}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("docker", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), fmt.Errorf("docker compose %s: %w\n%s", strings.Join(args, " "), err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// DockerExec runs a docker command.
func DockerExec(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(output)), fmt.Errorf("docker %s: %w\n%s", strings.Join(args, " "), err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// ContainerRunning checks if a container with the given name is running.
func ContainerRunning(containerName string) (bool, error) {
	output, err := DockerExec("ps", "--filter", fmt.Sprintf("name=%s", containerName), "--filter", "status=running", "--format", "{{.Names}}")
	if err != nil {
		return false, err
	}
	return strings.Contains(output, containerName), nil
}

// ContainerStatus returns the status of a container.
func ContainerStatus(containerName string) (string, error) {
	output, err := DockerExec("ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}\t{{.Status}}\t{{.Image}}")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(output) == "" {
		return "missing", nil
	}
	return output, nil
}

// HealthCheck runs docker inspect to get health status.
func HealthCheck(containerName string) (string, error) {
	output, err := DockerExec("inspect", "--format", "{{.State.Health.Status}}", containerName)
	if err != nil {
		return "unknown", nil
	}
	if output == "" || strings.Contains(output, "<nil>") {
		return "no health check", nil
	}
	return output, nil
}

// IsDockerAvailable checks if Docker binary exists and daemon is running.
func IsDockerAvailable() (bool, string) {
	_, err := exec.LookPath("docker")
	if err != nil {
		return false, "docker binary not found in PATH"
	}

	output, err := DockerExec("info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return false, fmt.Sprintf("Docker daemon not running: %v", err)
	}

	return true, fmt.Sprintf("Docker version %s", strings.TrimSpace(output))
}

// IsDockerComposeAvailable checks if Docker Compose plugin is available.
func IsDockerComposeAvailable() (bool, string) {
	output, err := ComposeExec("version")
	if err != nil {
		return false, "Docker Compose not available"
	}
	return true, strings.TrimSpace(output)
}

// PsStatus parses docker ps output for all containers matching filter.
func PsStatus(composeFile string) (string, error) {
	return ComposeExec(composeFile, "ps")
}
