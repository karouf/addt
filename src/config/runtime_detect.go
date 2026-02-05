package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DetectContainerRuntime automatically detects which container runtime to use
// Priority: explicit ADDT_PROVIDER > Podman (if available) > Docker (if running) > Podman (default)
func DetectContainerRuntime() string {
	// If explicitly set, use that
	if provider := os.Getenv("ADDT_PROVIDER"); provider != "" {
		return provider
	}

	// Check if Podman is available (preferred - no daemon required)
	if isPodmanAvailable() {
		return "podman"
	}

	// Check if Docker is available and running
	if isDockerRunning() {
		return "docker"
	}

	// Default to podman (will offer to install if not available)
	return "podman"
}

// EnsureContainerRuntime ensures a container runtime is available
// Downloads Podman automatically if needed (unless Docker is explicitly selected)
func EnsureContainerRuntime() (string, error) {
	// If Docker is explicitly selected, use it without auto-download
	if provider := os.Getenv("ADDT_PROVIDER"); provider == "docker" {
		if !isDockerRunning() {
			return "", fmt.Errorf("Docker is explicitly selected but not running")
		}
		return "docker", nil
	}

	// Check if Podman binary exists (even if machine not running)
	podmanPath := GetPodmanPath()
	if podmanPath != "" {
		// Podman binary exists - ensure machine is running on macOS
		if runtime.GOOS == "darwin" {
			if err := ensurePodmanMachine(podmanPath); err != nil {
				return "", fmt.Errorf("failed to start Podman machine: %w", err)
			}
		}
		return "podman", nil
	}

	// Check if Docker is available as fallback
	if isDockerRunning() {
		return "docker", nil
	}

	// Neither available - auto-download Podman
	fmt.Println("No container runtime found. Downloading Podman...")
	if err := DownloadPodman(); err != nil {
		return "", fmt.Errorf("failed to download Podman: %w", err)
	}

	// Verify it works now
	if isPodmanAvailable() {
		return "podman", nil
	}

	return "", fmt.Errorf("Podman downloaded but not working")
}

// isDockerRunning checks if Docker daemon is running
func isDockerRunning() bool {
	// First check if docker command exists
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return false
	}

	// Check if daemon is responsive
	cmd := exec.Command(dockerPath, "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// isPodmanAvailable checks if Podman is available and functional
// Checks both system Podman and bundled Podman
// On macOS, also verifies that a machine is running
func isPodmanAvailable() bool {
	podmanPath := GetPodmanPath()
	if podmanPath == "" {
		return false
	}

	// Check version works
	cmd := exec.Command(podmanPath, "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if cmd.Run() != nil {
		return false
	}

	// On macOS, we need a machine running
	if runtime.GOOS == "darwin" {
		return isPodmanMachineRunning(podmanPath)
	}

	return true
}

// isPodmanMachineRunning checks if a Podman machine is running (macOS)
func isPodmanMachineRunning(podmanPath string) bool {
	cmd := exec.Command(podmanPath, "machine", "list", "--format", "{{.Running}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "true")
}

// EnsurePodmanMachineRunning ensures the Podman machine is running (macOS only)
// This should be called before any Podman commands that require the machine
func EnsurePodmanMachineRunning() error {
	if runtime.GOOS != "darwin" {
		return nil // Only needed on macOS
	}

	podmanPath := GetPodmanPath()
	if podmanPath == "" {
		return fmt.Errorf("podman not found")
	}

	return ensurePodmanMachine(podmanPath)
}

// GetPodmanPath returns the path to Podman binary (system or bundled)
func GetPodmanPath() string {
	// First check system Podman (installed via Homebrew, package manager, etc.)
	if path, err := exec.LookPath("podman"); err == nil {
		return path
	}

	// On macOS, the bundled binary is remote-only and can't run machines
	// So we only use it on Linux
	if runtime.GOOS == "darwin" {
		return ""
	}

	// Check bundled Podman (Linux only)
	bundledPath := GetBundledPodmanPath()
	if bundledPath != "" {
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath
		}
	}

	return ""
}

// GetRuntimeInfo returns information about the detected runtime
func GetRuntimeInfo() (runtime string, version string, extras []string) {
	runtime = DetectContainerRuntime()

	switch runtime {
	case "docker":
		version = getDockerVersion()
	case "podman":
		version = getPodmanVersion()
		if hasPasta() {
			extras = append(extras, "pasta")
		}
	}

	return runtime, version, extras
}

func getDockerVersion() string {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getPodmanVersion() string {
	podmanPath := GetPodmanPath()
	if podmanPath == "" {
		return "unknown"
	}
	// Use --version flag which works without daemon connection
	cmd := exec.Command(podmanPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	// Parse "podman version X.Y.Z" -> "X.Y.Z"
	version := strings.TrimSpace(string(output))
	return strings.TrimPrefix(version, "podman version ")
}

func hasPasta() bool {
	_, err := exec.LookPath("pasta")
	return err == nil
}
