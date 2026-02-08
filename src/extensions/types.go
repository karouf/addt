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

// ExtensionConfig represents the config.yaml structure for extension source files
// Used when reading extension configs from embedded filesystem or local ~/.addt/extensions/
type ExtensionConfig struct {
	Name             string           `yaml:"name" json:"name"`
	Description      string           `yaml:"description" json:"description"`
	Entrypoint       Entrypoint       `yaml:"entrypoint" json:"entrypoint"`
	DefaultVersion   string           `yaml:"default_version" json:"default_version,omitempty"`
	AutoMount        bool             `yaml:"auto_mount" json:"auto_mount"`
	Autotrust        bool             `yaml:"autotrust" json:"autotrust"`       // Trust the /workspace directory on first launch
	AutoLogin        bool             `yaml:"auto_login" json:"auto_login"`     // Automatically handle authentication on first launch
	LoginMethod      string           `yaml:"login_method" json:"login_method"` // How to authenticate: native, env, auto
	Dependencies     []string         `yaml:"dependencies" json:"dependencies,omitempty"`
	EnvVars          []string         `yaml:"env_vars" json:"env_vars,omitempty"`
	OtelVars         []string         `yaml:"otel_vars" json:"otel_vars,omitempty"` // OpenTelemetry env vars; supports "VAR" or "VAR=default"
	Mounts           []ExtensionMount `yaml:"mounts" json:"mounts,omitempty"`
	Flags            []ExtensionFlag  `yaml:"flags" json:"flags,omitempty"`
	CredentialScript string           `yaml:"credential_script,omitempty" json:"credential_script,omitempty"` // Script to run on host for credentials
	IsLocal          bool             `yaml:"-" json:"-"`                                                     // Runtime flag, not serialized
}

// ExtensionMetadata represents metadata for an installed extension inside a Docker image
// Used when reading extensions.json from built Docker images
type ExtensionMetadata struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Entrypoint  Entrypoint       `json:"entrypoint"`
	AutoMount   *bool            `json:"auto_mount,omitempty"`   // true = auto mount, nil or false = disabled (default)
	Autotrust   *bool            `json:"autotrust,omitempty"`    // true = trust /workspace on first launch
	AutoLogin   *bool            `json:"auto_login,omitempty"`   // true = auto login on first launch
	LoginMethod string           `json:"login_method,omitempty"` // native, env, auto
	Mounts      []ExtensionMount `json:"mounts,omitempty"`
	Flags       []ExtensionFlag  `json:"flags,omitempty"`
	EnvVars     []string         `json:"env_vars,omitempty"`
	OtelVars    []string         `json:"otel_vars,omitempty"` // OpenTelemetry env vars; supports "VAR" or "VAR=default"
}

// ExtensionsJSONConfig represents the extensions.json file structure inside Docker images
type ExtensionsJSONConfig struct {
	Extensions map[string]ExtensionMetadata `json:"extensions"`
}

// ExtensionMountWithName includes the extension name for mount filtering
type ExtensionMountWithName struct {
	Source        string
	Target        string
	ExtensionName string
	AutoMount     *bool // from extension level
}
