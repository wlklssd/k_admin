package kadmin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSystemConfigPathUsesProjectRoot(t *testing.T) {
	root := systemConfigProjectRoot()
	want := filepath.Join(root, "data", "system_config.json")
	if got := resolveSystemConfigPath("", root); got != want {
		t.Fatalf("default system config path = %q, want %q", got, want)
	}
	if !filepath.IsAbs(want) {
		t.Fatalf("system config path is not absolute: %q", want)
	}
}

func TestSystemConfigRelativeEnvironmentPathUsesProjectRoot(t *testing.T) {
	root := systemConfigProjectRoot()
	want := filepath.Join(root, "custom", "settings.json")
	if got := resolveSystemConfigPath(filepath.Join("custom", "settings.json"), root); got != want {
		t.Fatalf("relative system config path = %q, want %q", got, want)
	}
}

func TestSystemConfigAbsoluteEnvironmentPathIsPreserved(t *testing.T) {
	want := filepath.Join(t.TempDir(), "settings.json")
	if got := resolveSystemConfigPath(want, `D:\unrelated`); got != want {
		t.Fatalf("absolute system config path = %q, want %q", got, want)
	}
}

func TestFindKAdminProjectRootFromNestedDirectory(t *testing.T) {
	root := systemConfigProjectRoot()
	nested := filepath.Join(root, "internal", "kadmin")
	if got := findKAdminProjectRoot(nested); got != root {
		t.Fatalf("project root = %q, want %q", got, root)
	}
}

func TestSystemConfigPathDoesNotFollowWorkingDirectory(t *testing.T) {
	originalWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	t.Setenv(systemConfigPathEnv, "")
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	want := filepath.Join(systemConfigProjectRoot(), "data", "system_config.json")
	if got := systemConfigPath(); got != want {
		t.Fatalf("system config path = %q, want %q", got, want)
	}
}

func TestSystemConfigReadWriteUsesExplicitPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "system_config.json")
	t.Setenv(systemConfigPathEnv, path)
	store := &Store{}
	want := map[string]string{
		"auth.default_password":         "changed",
		"auth.default_username":         "admin",
		"security.captcha_enabled":      "true",
		"ui.theme_mode":                 "dark",
		"custom.relative_path_verified": "yes",
	}
	if err := store.writeSystemConfig(want); err != nil {
		t.Fatalf("write system config: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat system config: %v", err)
	}
	got, err := store.readSystemConfig()
	if err != nil {
		t.Fatalf("read system config: %v", err)
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("config %s = %q, want %q", key, got[key], value)
		}
	}
}
