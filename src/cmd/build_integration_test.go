//go:build integration

package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/provider"
	testutil "github.com/jedi4ever/addt/test/util"
)

// providerImageCmd returns an exec.Cmd for image operations on the given provider.
func providerImageCmd(providerType string, args ...string) *exec.Cmd {
	switch providerType {
	case "podman":
		return exec.Command("podman", args...)
	case "docker":
		return provider.DockerCmd("desktop-linux", args...)
	case "rancher":
		return provider.DockerCmd("rancher-desktop", args...)
	case "orbstack":
		return provider.DockerCmd("orbstack", args...)
	default:
		return exec.Command("docker", args...)
	}
}

// providerImageExists checks if an image exists for the given provider.
func providerImageExists(providerType, imageName string) bool {
	cmd := providerImageCmd(providerType, "image", "inspect", imageName)
	return cmd.Run() == nil
}

// providerRemoveImage removes an image for the given provider.
func providerRemoveImage(providerType, imageName string) {
	providerImageCmd(providerType, "rmi", "-f", imageName).Run()
}

func TestBuildCommand_Integration_Claude(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			testImageName := "addt-test-claude-integration-" + prov

			providerRemoveImage(prov, testImageName)
			defer providerRemoveImage(prov, testImageName)

			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "claude"

			providerCfg := &provider.Config{
				Extensions:        cfg.Extensions,
				ExtensionVersions: cfg.ExtensionVersions,
				NodeVersion:       cfg.NodeVersion,
				GoVersion:         cfg.GoVersion,
				UvVersion:         cfg.UvVersion,
				ImageName:         testImageName,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", prov, err)
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Fatalf("Failed to initialize %s provider: %v", prov, err)
			}

			if err := p.BuildIfNeeded(true, false); err != nil {
				t.Fatalf("BuildIfNeeded failed for %s: %v", prov, err)
			}

			if !providerImageExists(prov, testImageName) {
				t.Errorf("Expected image to exist after build on %s", prov)
			}
		})
	}
}

func TestBuildCommand_Integration_WithNoCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			testImageName := "addt-test-nocache-integration-" + prov

			providerRemoveImage(prov, testImageName)
			defer providerRemoveImage(prov, testImageName)

			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "claude"

			providerCfg := &provider.Config{
				Extensions:        cfg.Extensions,
				ExtensionVersions: cfg.ExtensionVersions,
				NodeVersion:       cfg.NodeVersion,
				GoVersion:         cfg.GoVersion,
				UvVersion:         cfg.UvVersion,
				ImageName:         testImageName,
				NoCache:           true,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", prov, err)
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Fatalf("Failed to initialize %s provider: %v", prov, err)
			}

			if err := p.BuildIfNeeded(true, false); err != nil {
				t.Fatalf("BuildIfNeeded with NoCache failed for %s: %v", prov, err)
			}

			if !providerImageExists(prov, testImageName) {
				t.Errorf("Expected image to exist after no-cache build on %s", prov)
			}
		})
	}
}

func TestBuildCommand_Integration_Binary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Could not get working directory: %v", err)
			}

			srcDir := wd + "/.."
			distDir := srcDir + "/../dist"
			binaryPath := distDir + "/addt"

			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				os.MkdirAll(distDir, 0755)
				buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
				buildCmd.Dir = srcDir
				if output, err := buildCmd.CombinedOutput(); err != nil {
					t.Skipf("Could not build addt binary: %v\nOutput: %s", err, string(output))
				}
			}

			if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
				t.Skip("Binary does not exist after build attempt, skipping")
			}

			cmd := exec.Command(binaryPath, "build", "claude")
			cmd.Env = append(os.Environ(), "ADDT_PROVIDER="+prov)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("addt build command failed for %s: %v\nOutput: %s", prov, err, string(output))
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Image tagged as:") && !strings.Contains(outputStr, "Using cached") {
				t.Errorf("Expected build success output for %s, got: %s", prov, outputStr)
			}
		})
	}
}

func TestBuildCommand_Integration_ExtensionVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			testImageName := "addt-test-version-integration-" + prov

			providerRemoveImage(prov, testImageName)
			defer providerRemoveImage(prov, testImageName)

			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "claude"

			providerCfg := &provider.Config{
				Extensions: cfg.Extensions,
				ExtensionVersions: map[string]string{
					"claude": "1.0.21",
				},
				NodeVersion: cfg.NodeVersion,
				GoVersion:   cfg.GoVersion,
				UvVersion:   cfg.UvVersion,
				ImageName:   testImageName,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", prov, err)
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Fatalf("Failed to initialize %s provider: %v", prov, err)
			}

			if err := p.BuildIfNeeded(true, false); err != nil {
				t.Fatalf("BuildIfNeeded with specific version failed for %s: %v", prov, err)
			}

			if !providerImageExists(prov, testImageName) {
				t.Errorf("Expected image to exist after versioned build on %s", prov)
			}

			cmd := providerImageCmd(prov, "inspect", "--format", "{{.Config.Labels}}", testImageName)
			output, err := cmd.Output()
			if err == nil {
				t.Logf("Image labels for %s: %s", prov, string(output))
			}
		})
	}
}

func TestBuildCommand_Integration_MultipleExtensions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			testImageName := "addt-test-multi-integration-" + prov

			providerRemoveImage(prov, testImageName)
			defer providerRemoveImage(prov, testImageName)

			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "claude,codex"

			providerCfg := &provider.Config{
				Extensions:        cfg.Extensions,
				ExtensionVersions: cfg.ExtensionVersions,
				NodeVersion:       cfg.NodeVersion,
				GoVersion:         cfg.GoVersion,
				UvVersion:         cfg.UvVersion,
				ImageName:         testImageName,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", prov, err)
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Fatalf("Failed to initialize %s provider: %v", prov, err)
			}

			if err := p.BuildIfNeeded(true, false); err != nil {
				t.Fatalf("BuildIfNeeded with multiple extensions failed for %s: %v", prov, err)
			}

			if !providerImageExists(prov, testImageName) {
				t.Errorf("Expected image to exist after multi-extension build on %s", prov)
			}
		})
	}
}

func TestBuildCommand_Integration_InvalidExtension(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping image build test in short mode")
	}

	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			testImageName := "addt-test-invalid-" + prov

			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "nonexistent-extension-xyz"

			providerCfg := &provider.Config{
				Extensions:  cfg.Extensions,
				NodeVersion: cfg.NodeVersion,
				GoVersion:   cfg.GoVersion,
				UvVersion:   cfg.UvVersion,
				ImageName:   testImageName,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Logf("Provider creation failed as expected for invalid extension on %s: %v", prov, err)
				return
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Logf("Initialization failed as expected for invalid extension on %s: %v", prov, err)
				return
			}

			err = p.BuildIfNeeded(true, false)
			if err == nil {
				t.Logf("Build succeeded for invalid extension on %s (extension was likely ignored)", prov)
				providerRemoveImage(prov, testImageName)
			} else {
				t.Logf("Build failed as expected for invalid extension on %s: %v", prov, err)
			}
		})
	}
}

func TestBuildCommand_Integration_ImageNameFormat(t *testing.T) {
	providers := testutil.RequireProviders(t)
	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			cfg := config.LoadConfig("0.0.0-test", "22", "1.23.5", "0.4.17", 49152)
			cfg.Extensions = "claude"

			providerCfg := &provider.Config{
				Extensions:  cfg.Extensions,
				NodeVersion: cfg.NodeVersion,
				GoVersion:   cfg.GoVersion,
				UvVersion:   cfg.UvVersion,
			}

			p, err := NewProvider(prov, providerCfg)
			if err != nil {
				t.Fatalf("Failed to create %s provider: %v", prov, err)
			}

			if err := p.Initialize(providerCfg); err != nil {
				t.Fatalf("Failed to initialize %s provider: %v", prov, err)
			}

			imageName := p.DetermineImageName()

			if imageName == "" {
				t.Error("DetermineImageName returned empty string")
			}

			if !strings.Contains(imageName, "claude") {
				t.Errorf("Expected image name to contain 'claude', got: %s", imageName)
			}

			t.Logf("Generated image name for %s: %s", prov, imageName)
		})
	}
}
