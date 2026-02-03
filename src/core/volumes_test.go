package core

import (
	"testing"

	"github.com/jedi4ever/addt/provider"
)

func TestBuildVolumes_AutomountEnabled(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: true,
	}

	volumes := BuildVolumes(cfg, "/home/user/project")

	if len(volumes) != 1 {
		t.Fatalf("Expected 1 volume, got %d", len(volumes))
	}

	if volumes[0].Source != "/home/user/project" {
		t.Errorf("Volume source = %q, want '/home/user/project'", volumes[0].Source)
	}

	if volumes[0].Target != "/workspace" {
		t.Errorf("Volume target = %q, want '/workspace'", volumes[0].Target)
	}

	if volumes[0].ReadOnly {
		t.Error("Volume should not be read-only")
	}
}

func TestBuildVolumes_AutomountDisabled(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: false,
	}

	volumes := BuildVolumes(cfg, "/home/user/project")

	if len(volumes) != 0 {
		t.Errorf("Expected 0 volumes when automount disabled, got %d", len(volumes))
	}
}

func TestBuildVolumes_DifferentPaths(t *testing.T) {
	cfg := &provider.Config{
		WorkdirAutomount: true,
	}

	testPaths := []string{
		"/home/user/project",
		"/tmp/test",
		"/var/data",
	}

	for _, path := range testPaths {
		t.Run(path, func(t *testing.T) {
			volumes := BuildVolumes(cfg, path)

			if len(volumes) != 1 {
				t.Fatalf("Expected 1 volume, got %d", len(volumes))
			}

			if volumes[0].Source != path {
				t.Errorf("Volume source = %q, want %q", volumes[0].Source, path)
			}
		})
	}
}
