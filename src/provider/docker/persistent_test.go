package docker

import (
	"strings"
	"testing"

	"github.com/jedi4ever/addt/provider"
)

// createPersistentUnitProvider creates a DockerProvider for unit tests
func createPersistentUnitProvider(workdir, extensions string) *DockerProvider {
	return &DockerProvider{
		config: &provider.Config{
			Workdir:    workdir,
			Extensions: extensions,
		},
		tempDirs: []string{},
	}
}

func TestGenerateContainerName_Format(t *testing.T) {
	prov := createPersistentUnitProvider("/home/user/myproject", "claude")
	name := prov.GenerateContainerName()

	if !strings.HasPrefix(name, "addt-persistent-myproject-") {
		t.Errorf("GenerateContainerName() = %q, want prefix 'addt-persistent-myproject-'", name)
	}

	// Should have 8-char hex hash suffix
	parts := strings.Split(name, "-")
	hash := parts[len(parts)-1]
	if len(hash) != 8 {
		t.Errorf("Hash suffix %q should be 8 chars", hash)
	}
}

func TestGenerateContainerName_Consistent(t *testing.T) {
	prov := createPersistentUnitProvider("/home/user/project", "claude")

	name1 := prov.GenerateContainerName()
	name2 := prov.GenerateContainerName()

	if name1 != name2 {
		t.Errorf("GenerateContainerName() not consistent: %q != %q", name1, name2)
	}
}

func TestGenerateContainerName_SameWorkdirSameExtensions(t *testing.T) {
	prov1 := createPersistentUnitProvider("/home/user/project", "claude")
	prov2 := createPersistentUnitProvider("/home/user/project", "claude")

	if prov1.GenerateContainerName() != prov2.GenerateContainerName() {
		t.Error("Same workdir + same extensions should produce same name")
	}
}

func TestGenerateContainerName_DifferentExtensions(t *testing.T) {
	workdir := "/home/user/project"

	name1 := createPersistentUnitProvider(workdir, "claude").GenerateContainerName()
	name2 := createPersistentUnitProvider(workdir, "codex").GenerateContainerName()
	name3 := createPersistentUnitProvider(workdir, "claude,codex").GenerateContainerName()

	if name1 == name2 {
		t.Errorf("Different extensions should produce different names: %q", name1)
	}
	if name1 == name3 {
		t.Errorf("Different extensions should produce different names: %q", name1)
	}
	if name2 == name3 {
		t.Errorf("Different extensions should produce different names: %q", name2)
	}
}

func TestGenerateContainerName_DifferentWorkdirs(t *testing.T) {
	name1 := createPersistentUnitProvider("/home/user/project-a", "claude").GenerateContainerName()
	name2 := createPersistentUnitProvider("/home/user/project-b", "claude").GenerateContainerName()

	if name1 == name2 {
		t.Errorf("Different workdirs should produce different names: %q", name1)
	}
}

func TestGenerateContainerName_ExtensionOrderIndependent(t *testing.T) {
	name1 := createPersistentUnitProvider("/tmp/test", "claude,codex").GenerateContainerName()
	name2 := createPersistentUnitProvider("/tmp/test", "codex,claude").GenerateContainerName()

	if name1 != name2 {
		t.Errorf("Extension order should not matter: %q != %q", name1, name2)
	}
}

func TestGenerateContainerName_ExtensionWhitespace(t *testing.T) {
	name1 := createPersistentUnitProvider("/tmp/test", "claude, codex").GenerateContainerName()
	name2 := createPersistentUnitProvider("/tmp/test", "claude,codex").GenerateContainerName()

	if name1 != name2 {
		t.Errorf("Whitespace in extensions should not matter: %q != %q", name1, name2)
	}
}

func TestGenerateContainerName_EmptyExtensions(t *testing.T) {
	name := createPersistentUnitProvider("/tmp/test", "").GenerateContainerName()

	if !strings.HasPrefix(name, "addt-persistent-test-") {
		t.Errorf("Empty extensions should still work: %q", name)
	}
}

func TestGenerateContainerName_SpecialChars(t *testing.T) {
	testCases := []struct {
		workdir    string
		wantPrefix string
	}{
		{"/home/user/My Project!", "addt-persistent-my-project-"},
		{"/home/user/UPPERCASE", "addt-persistent-uppercase-"},
		{"/home/user/dots.and.more", "addt-persistent-dots-and-more-"},
		{"/home/user/under_scores", "addt-persistent-under-scores-"},
		{"/home/user/---dashes---", "addt-persistent-dashes-"},
	}

	for _, tc := range testCases {
		t.Run(tc.workdir, func(t *testing.T) {
			name := createPersistentUnitProvider(tc.workdir, "claude").GenerateContainerName()
			if !strings.HasPrefix(name, tc.wantPrefix) {
				t.Errorf("GenerateContainerName(%q) = %q, want prefix %q", tc.workdir, name, tc.wantPrefix)
			}
		})
	}
}

func TestGenerateContainerName_LongDirname(t *testing.T) {
	prov := createPersistentUnitProvider("/home/user/this-is-a-very-long-directory-name-that-exceeds-limit", "claude")
	name := prov.GenerateContainerName()

	// Remove prefix "addt-persistent-" and hash suffix "-xxxxxxxx"
	trimmed := strings.TrimPrefix(name, "addt-persistent-")
	dirPart := trimmed[:strings.LastIndex(trimmed, "-")]

	if len(dirPart) > 20 {
		t.Errorf("Directory part %q exceeds 20 chars (len=%d)", dirPart, len(dirPart))
	}
}

func TestGenerateContainerName_EmptyWorkdir(t *testing.T) {
	// Empty workdir falls back to os.Getwd()
	prov := createPersistentUnitProvider("", "claude")
	name := prov.GenerateContainerName()

	if !strings.HasPrefix(name, "addt-persistent-") {
		t.Errorf("Empty workdir should still produce valid name: %q", name)
	}
}

func TestGenerateEphemeralName_Format(t *testing.T) {
	prov := createPersistentUnitProvider("/tmp", "claude")
	name := prov.GenerateEphemeralName()

	if !strings.HasPrefix(name, "addt-") {
		t.Errorf("Ephemeral name should start with 'addt-': %q", name)
	}

	if strings.HasPrefix(name, "addt-persistent-") {
		t.Errorf("Ephemeral name should NOT start with 'addt-persistent-': %q", name)
	}
}

func TestGenerateEphemeralName_ContainsPID(t *testing.T) {
	prov := createPersistentUnitProvider("/tmp", "claude")
	name := prov.GenerateEphemeralName()

	// Should contain a PID (digits at end)
	parts := strings.Split(name, "-")
	lastPart := parts[len(parts)-1]
	for _, c := range lastPart {
		if c < '0' || c > '9' {
			t.Errorf("Ephemeral name should end with PID, got %q in %q", lastPart, name)
			break
		}
	}
}

func TestGeneratePersistentName_IsAlias(t *testing.T) {
	prov := createPersistentUnitProvider("/tmp/test", "claude")

	name1 := prov.GenerateContainerName()
	name2 := prov.GeneratePersistentName()

	if name1 != name2 {
		t.Errorf("GeneratePersistentName() should equal GenerateContainerName(): %q != %q", name1, name2)
	}
}

func TestIsPersistentContainer(t *testing.T) {
	testCases := []struct {
		name string
		want bool
	}{
		{"addt-persistent-myproject-abc12345", true},
		{"addt-persistent-test-xyz", true},
		{"addt-persistent-a-b", true},
		{"addt-persistent-", true},
		{"addt-20240101-123456-1234", false},
		{"addt-test-123", false},
		{"some-other-container", false},
		{"addt-", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPersistentContainer(tc.name)
			if got != tc.want {
				t.Errorf("IsPersistentContainer(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestIsEphemeralContainer(t *testing.T) {
	testCases := []struct {
		name string
		want bool
	}{
		{"addt-20240101-123456-1234", true},
		{"addt-test-123", true},
		{"addt-something", true},
		{"addt-persistent-myproject-abc12345", false},
		{"some-other-container", false},
		{"docker-container", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEphemeralContainer(tc.name)
			if got != tc.want {
				t.Errorf("IsEphemeralContainer(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestGetContainerWorkdir(t *testing.T) {
	testCases := []struct {
		name string
		want string
	}{
		{"addt-persistent-myproject-abc12345", "myproject"},
		{"addt-persistent-test-xyz", "test"},
		{"addt-persistent-my-long-name-abc", "my-long-name"},
		{"addt-persistent-a-b-c-d", "a-b-c"},
		{"addt-20240101-123456-1234", ""},
		{"some-other-container", ""},
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetContainerWorkdir(tc.name)
			if got != tc.want {
				t.Errorf("GetContainerWorkdir(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestGetContainerWorkdir_NoHash(t *testing.T) {
	// Edge case: persistent name with no hash separator
	got := GetContainerWorkdir("addt-persistent-onlyname")
	if got != "onlyname" {
		t.Errorf("GetContainerWorkdir with no hash = %q, want 'onlyname'", got)
	}
}

func TestIsPersistentAndEphemeral_MutuallyExclusive(t *testing.T) {
	names := []string{
		"addt-persistent-test-abc",
		"addt-20240101-123456-1234",
		"addt-something",
		"other-container",
	}

	for _, name := range names {
		isPersistent := IsPersistentContainer(name)
		isEphemeral := IsEphemeralContainer(name)

		if isPersistent && isEphemeral {
			t.Errorf("Name %q is both persistent and ephemeral", name)
		}
	}
}
