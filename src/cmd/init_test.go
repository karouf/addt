package cmd

import (
	"os"
	"path/filepath"
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestDetectProjectType_NodeJS(t *testing.T) {
	// Create temp dir with package.json
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.WriteFile("package.json", []byte(`{"name": "test"}`), 0644)

	project := detectProjectType()

	if project.Language != "Node.js" {
		t.Errorf("expected Node.js, got %s", project.Language)
	}
	if project.PackageFile != "package.json" {
		t.Errorf("expected package.json, got %s", project.PackageFile)
	}
}

func TestDetectProjectType_Python(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.WriteFile("requirements.txt", []byte("flask==2.0"), 0644)

	project := detectProjectType()

	if project.Language != "Python" {
		t.Errorf("expected Python, got %s", project.Language)
	}
	if project.PackageFile != "requirements.txt" {
		t.Errorf("expected requirements.txt, got %s", project.PackageFile)
	}
}

func TestDetectProjectType_Go(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.WriteFile("go.mod", []byte("module test\n\ngo 1.21"), 0644)

	project := detectProjectType()

	if project.Language != "Go" {
		t.Errorf("expected Go, got %s", project.Language)
	}
	if project.PackageFile != "go.mod" {
		t.Errorf("expected go.mod, got %s", project.PackageFile)
	}
}

func TestDetectProjectType_Rust(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.WriteFile("Cargo.toml", []byte("[package]\nname = \"test\""), 0644)

	project := detectProjectType()

	if project.Language != "Rust" {
		t.Errorf("expected Rust, got %s", project.Language)
	}
}

func TestDetectProjectType_WithGit(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.MkdirAll(".git", 0755)
	os.WriteFile("package.json", []byte(`{"name": "test"}`), 0644)

	project := detectProjectType()

	if !project.HasGit {
		t.Error("expected HasGit to be true")
	}
}

func TestDetectProjectType_WithGitHub(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	os.MkdirAll(".git", 0755)
	os.WriteFile(filepath.Join(".git", "config"), []byte("[remote \"origin\"]\n\turl = git@github.com:user/repo.git"), 0644)

	project := detectProjectType()

	if !project.HasGitHub {
		t.Error("expected HasGitHub to be true")
	}
}

func TestConfigureDefaults(t *testing.T) {
	project := ProjectType{
		Language:  "Node.js",
		HasGit:    true,
		HasGitHub: true,
	}

	config := &InitConfig{}
	configureDefaults(config, project)

	if config.Extensions != "claude" {
		t.Errorf("expected claude extension, got %s", config.Extensions)
	}

	if config.Persistent == nil || *config.Persistent {
		t.Error("expected persistent to be false")
	}

	if config.Firewall == nil || config.Firewall.Enabled == nil || !*config.Firewall.Enabled {
		t.Error("expected firewall to be enabled")
	}

	if config.Firewall.Mode != "strict" {
		t.Errorf("expected strict firewall mode, got %s", config.Firewall.Mode)
	}

	if config.SSH == nil || config.SSH.ForwardKeys == nil || !*config.SSH.ForwardKeys {
		t.Error("expected SSH forward keys to be true")
	}
	if config.SSH.ForwardMode != "proxy" {
		t.Errorf("expected proxy SSH forward mode, got %s", config.SSH.ForwardMode)
	}

	if config.GitHub == nil || config.GitHub.ForwardToken == nil || !*config.GitHub.ForwardToken {
		t.Error("expected GitHub.ForwardToken to be true")
	}
	if config.GitHub.TokenSource != "gh_auth" {
		t.Errorf("expected GitHub.TokenSource to be gh_auth, got %s", config.GitHub.TokenSource)
	}

	if config.NodeVersion != "22" {
		t.Errorf("expected Node version 22, got %s", config.NodeVersion)
	}

	// Check firewall allowed domains
	foundNpm := false
	for _, d := range config.Firewall.Allowed {
		if d == "registry.npmjs.org" {
			foundNpm = true
		}
	}
	if !foundNpm {
		t.Error("expected npm registry in allowed domains")
	}
}

func TestConfigureDefaults_Go(t *testing.T) {
	project := ProjectType{
		Language: "Go",
	}

	config := &InitConfig{}
	configureDefaults(config, project)

	if config.GoVersion != "1.24" {
		t.Errorf("expected Go version 1.24, got %s", config.GoVersion)
	}

	// Check firewall allowed domains
	foundProxy := false
	for _, d := range config.Firewall.Allowed {
		if d == "proxy.golang.org" {
			foundProxy = true
		}
	}
	if !foundProxy {
		t.Error("expected Go proxy in allowed domains")
	}
}

func TestGetDefaultAllowedDomains(t *testing.T) {
	testCases := []struct {
		name     string
		project  ProjectType
		expected []string
	}{
		{
			name:    "Node.js project",
			project: ProjectType{Language: "Node.js"},
			expected: []string{
				"api.anthropic.com",
				"api.openai.com",
				"registry.npmjs.org",
			},
		},
		{
			name:    "Python project",
			project: ProjectType{Language: "Python"},
			expected: []string{
				"api.anthropic.com",
				"pypi.org",
			},
		},
		{
			name:    "GitHub project",
			project: ProjectType{HasGitHub: true},
			expected: []string{
				"api.anthropic.com",
				"api.github.com",
				"github.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			domains := getDefaultAllowedDomains(tc.project)
			for _, expected := range tc.expected {
				found := false
				for _, got := range domains {
					if got == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected domain %s not found in %v", expected, domains)
				}
			}
		})
	}
}

func TestWriteConfig(tt *testing.T) {
	dir := tt.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	t := true
	f := false
	config := &InitConfig{
		Extensions: "claude",
		Persistent: &f,
		Firewall: &cfgtypes.FirewallSettings{
			Enabled: &t,
			Mode:    "strict",
			Allowed: []string{"api.anthropic.com"},
		},
		SSH: &cfgtypes.SSHSettings{ForwardKeys: &t, ForwardMode: "proxy"},
	}

	err := writeConfig(config)
	if err != nil {
		tt.Fatalf("writeConfig failed: %v", err)
	}

	// Check file was created
	data, err := os.ReadFile(".addt.yaml")
	if err != nil {
		tt.Fatalf("failed to read .addt.yaml: %v", err)
	}

	content := string(data)
	if content == "" {
		tt.Error("config file is empty")
	}

	// Check header
	if !contains(content, "# addt configuration") {
		tt.Error("expected header comment")
	}

	// Check content
	if !contains(content, "extensions: claude") {
		tt.Error("expected extensions field")
	}
	if !contains(content, "enabled: true") {
		tt.Error("expected firewall enabled field")
	}
	if !contains(content, "forward_keys: true") {
		tt.Error("expected forward_keys field")
	}
	if !contains(content, "forward_mode: proxy") {
		tt.Error("expected forward_mode field")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Create a file
	os.WriteFile("test.txt", []byte("test"), 0644)

	if !fileExists("test.txt") {
		t.Error("expected fileExists to return true")
	}

	if fileExists("nonexistent.txt") {
		t.Error("expected fileExists to return false for nonexistent file")
	}
}

func TestDirExists(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Create a directory
	os.MkdirAll("testdir", 0755)

	if !dirExists("testdir") {
		t.Error("expected dirExists to return true")
	}

	if dirExists("nonexistent") {
		t.Error("expected dirExists to return false for nonexistent dir")
	}

	// File is not a directory
	os.WriteFile("testfile", []byte("test"), 0644)
	if dirExists("testfile") {
		t.Error("expected dirExists to return false for file")
	}
}

func TestContainsGitHubRemote(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Test with no .git
	if containsGitHubRemote() {
		t.Error("expected false without .git")
	}

	// Test with .git but no github
	os.MkdirAll(".git", 0755)
	os.WriteFile(filepath.Join(".git", "config"), []byte("[remote \"origin\"]\n\turl = git@gitlab.com:user/repo.git"), 0644)
	if containsGitHubRemote() {
		t.Error("expected false without github.com")
	}

	// Test with github
	os.WriteFile(filepath.Join(".git", "config"), []byte("[remote \"origin\"]\n\turl = git@github.com:user/repo.git"), 0644)
	if !containsGitHubRemote() {
		t.Error("expected true with github.com")
	}
}

func TestDetectCredentials(t *testing.T) {
	// Save original env
	origAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	origOpenAIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		os.Setenv("ANTHROPIC_API_KEY", origAnthropicKey)
		os.Setenv("OPENAI_API_KEY", origOpenAIKey)
	}()

	// Clear env
	os.Setenv("ANTHROPIC_API_KEY", "")
	os.Setenv("OPENAI_API_KEY", "")

	creds := detectCredentials()
	if creds["anthropic"] {
		t.Error("expected anthropic to be false without key")
	}

	// Set anthropic key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	creds = detectCredentials()
	if !creds["anthropic"] {
		t.Error("expected anthropic to be true with key")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
