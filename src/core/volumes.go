package core

import (
	"github.com/jedi4ever/addt/provider"
)

// BuildVolumes creates volume mounts from the configuration
func BuildVolumes(cfg *provider.Config, cwd string) []provider.VolumeMount {
	var volumes []provider.VolumeMount

	// Mount working directory if automount is enabled (default: true)
	if cfg.WorkdirAutomount {
		volumes = append(volumes, provider.VolumeMount{
			Source:   cwd,
			Target:   "/workspace",
			ReadOnly: cfg.WorkdirReadonly,
		})
	}

	return volumes
}
