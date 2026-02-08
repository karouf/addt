package extensions

import (
	"encoding/json"
)

// ExtensionMount represents a mount configuration for an extension
type ExtensionMount struct {
	Source string `yaml:"source" json:"source"`
	Target string `yaml:"target" json:"target"`
}

// ExtensionFlag represents a CLI flag for an extension
type ExtensionFlag struct {
	Flag        string `yaml:"flag" json:"flag"`
	Description string `yaml:"description" json:"description"`
	EnvVar      string `yaml:"env_var,omitempty" json:"env_var,omitempty"` // Set this env var to "true" when flag is present
}

// Entrypoint can be either a string or an array of strings
// Examples:
//
//	entrypoint: claude
//	entrypoint: ["bash", "-i"]
type Entrypoint []string

// UnmarshalYAML handles both string and array formats
func (e *Entrypoint) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try string first
	var str string
	if err := unmarshal(&str); err == nil {
		*e = []string{str}
		return nil
	}

	// Try array
	var arr []string
	if err := unmarshal(&arr); err != nil {
		return err
	}
	*e = arr
	return nil
}

// UnmarshalJSON handles both string and array formats
func (e *Entrypoint) UnmarshalJSON(data []byte) error {
	// Try string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*e = []string{str}
		return nil
	}

	// Try array
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*e = arr
	return nil
}

// MarshalJSON outputs as array
func (e Entrypoint) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(e))
}

// Command returns the command (first element)
func (e Entrypoint) Command() string {
	if len(e) == 0 {
		return ""
	}
	return e[0]
}

// Args returns the arguments (elements after the first)
func (e Entrypoint) Args() []string {
	if len(e) <= 1 {
		return []string{}
	}
	return e[1:]
}

// ExtensionAuthConfig holds auth settings in extension config.yaml
type ExtensionAuthConfig struct {
	Autologin bool   `yaml:"autologin" json:"autologin"` // Automatically handle authentication on first launch
	Method    string `yaml:"method" json:"method"`       // How to authenticate: native, env, auto
}

// ExtensionCfgSection holds the config: section in extension config.yaml
type ExtensionCfgSection struct {
	Automount bool             `yaml:"automount" json:"automount"` // Auto-mount extension config directories
	Readonly  bool             `yaml:"readonly" json:"readonly"`   // Mount extension config directories as read-only
	Mounts    []ExtensionMount `yaml:"mounts" json:"mounts,omitempty"`
}

// ExtensionConfig represents the config.yaml structure for extension source files
// Used when reading extension configs from embedded filesystem or local ~/.addt/extensions/
type ExtensionConfig struct {
	Name             string              `yaml:"name" json:"name"`
	Description      string              `yaml:"description" json:"description"`
	Entrypoint       Entrypoint          `yaml:"entrypoint" json:"entrypoint"`
	DefaultVersion   string              `yaml:"default_version" json:"default_version,omitempty"`
	Auth             ExtensionAuthConfig `yaml:"auth" json:"auth"`
	Config           ExtensionCfgSection `yaml:"config" json:"config"`
	Dependencies     []string            `yaml:"dependencies" json:"dependencies,omitempty"`
	EnvVars          []string            `yaml:"env_vars" json:"env_vars,omitempty"`
	OtelVars         []string            `yaml:"otel_vars" json:"otel_vars,omitempty"` // OpenTelemetry env vars; supports "VAR" or "VAR=default"
	Flags            []ExtensionFlag     `yaml:"flags" json:"flags,omitempty"`
	CredentialScript string              `yaml:"credential_script,omitempty" json:"credential_script,omitempty"` // Script to run on host for credentials
	IsLocal          bool                `yaml:"-" json:"-"`                                                     // Runtime flag, not serialized
}

// ExtensionAuthMetadata holds auth settings in extensions.json inside Docker images
type ExtensionAuthMetadata struct {
	Autologin *bool  `json:"autologin,omitempty"` // true = auto login on first launch
	Method    string `json:"method,omitempty"`    // native, env, auto
}

// ExtensionCfgSectionMetadata holds the config: section in extensions.json
type ExtensionCfgSectionMetadata struct {
	Automount *bool            `json:"automount,omitempty"` // true = auto mount config dirs, nil or false = disabled
	Readonly  *bool            `json:"readonly,omitempty"`  // true = mount config dirs as read-only
	Mounts    []ExtensionMount `json:"mounts,omitempty"`
}

// ExtensionMetadata represents metadata for an installed extension inside a Docker image
// Used when reading extensions.json from built Docker images
type ExtensionMetadata struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Entrypoint  Entrypoint                   `json:"entrypoint"`
	Auth        *ExtensionAuthMetadata       `json:"auth,omitempty"`
	Config      *ExtensionCfgSectionMetadata `json:"config,omitempty"`
	Flags       []ExtensionFlag              `json:"flags,omitempty"`
	EnvVars     []string                     `json:"env_vars,omitempty"`
	OtelVars    []string                     `json:"otel_vars,omitempty"` // OpenTelemetry env vars; supports "VAR" or "VAR=default"
}

// ExtensionsJSONConfig represents the extensions.json file structure inside Docker images
type ExtensionsJSONConfig struct {
	Extensions map[string]ExtensionMetadata `json:"extensions"`
}

// ExtensionMountWithName includes the extension name for mount filtering
type ExtensionMountWithName struct {
	Source          string
	Target          string
	ExtensionName   string
	ConfigAutomount *bool // from extension-level config.automount
	ConfigReadonly  *bool // from extension-level config.readonly
}
