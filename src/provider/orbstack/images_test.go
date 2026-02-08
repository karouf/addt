package orbstack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtAssetsHash_ChangesWhenExtraExtensionModified(t *testing.T) {
	// Scenario: user edits a setup.sh in their extra extensions dir,
	// and expects the image hash to change so a rebuild is triggered.

	extDir := t.TempDir()
	extName := filepath.Join(extDir, "myext")
	os.MkdirAll(extName, 0755)
	os.WriteFile(filepath.Join(extName, "setup.sh"), []byte("echo hello"), 0644)

	t.Setenv("ADDT_EXTENSIONS_DIR", extDir)

	p := &OrbStackProvider{}

	hash1 := p.extAssetsHash()

	// Modify the extension file
	os.WriteFile(filepath.Join(extName, "setup.sh"), []byte("echo changed"), 0644)

	hash2 := p.extAssetsHash()

	if hash1 == hash2 {
		t.Errorf("expected hash to change after modifying extra extension file, got same hash: %s", hash1)
	}
}

func TestExtAssetsHash_ChangesWhenLocalExtensionModified(t *testing.T) {
	// Scenario: user edits an extension under ~/.addt/extensions/,
	// and expects the image hash to change so a rebuild is triggered.

	addtHome := t.TempDir()
	localExtsDir := filepath.Join(addtHome, "extensions", "myext")
	os.MkdirAll(localExtsDir, 0755)
	os.WriteFile(filepath.Join(localExtsDir, "install.sh"), []byte("apt-get install foo"), 0644)

	t.Setenv("ADDT_HOME", addtHome)
	// Ensure ADDT_EXTENSIONS_DIR doesn't interfere
	t.Setenv("ADDT_EXTENSIONS_DIR", "")

	p := &OrbStackProvider{}

	hash1 := p.extAssetsHash()

	// Modify the extension file
	os.WriteFile(filepath.Join(localExtsDir, "install.sh"), []byte("apt-get install bar"), 0644)

	hash2 := p.extAssetsHash()

	if hash1 == hash2 {
		t.Errorf("expected hash to change after modifying local extension file, got same hash: %s", hash1)
	}
}

func TestExtAssetsHash_StableWhenNoExtensionDirs(t *testing.T) {
	// Scenario: no local or extra extensions exist, hash should be stable.

	t.Setenv("ADDT_HOME", t.TempDir())
	t.Setenv("ADDT_EXTENSIONS_DIR", "")

	p := &OrbStackProvider{}

	hash1 := p.extAssetsHash()
	hash2 := p.extAssetsHash()

	if hash1 != hash2 {
		t.Errorf("expected stable hash with no extension dirs, got %s and %s", hash1, hash2)
	}
}

func TestExtAssetsHash_ChangesWhenNewFileAdded(t *testing.T) {
	// Scenario: user adds a new file to their extra extensions dir.

	extDir := t.TempDir()
	extName := filepath.Join(extDir, "myext")
	os.MkdirAll(extName, 0755)
	os.WriteFile(filepath.Join(extName, "setup.sh"), []byte("echo hello"), 0644)

	t.Setenv("ADDT_EXTENSIONS_DIR", extDir)

	p := &OrbStackProvider{}

	hash1 := p.extAssetsHash()

	// Add a new file
	os.WriteFile(filepath.Join(extName, "credentials.sh"), []byte("export TOKEN=abc"), 0644)

	hash2 := p.extAssetsHash()

	if hash1 == hash2 {
		t.Errorf("expected hash to change after adding new extension file, got same hash: %s", hash1)
	}
}
