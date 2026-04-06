package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadStoredConfig(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("MONOLOGUE_CONFIG_DIR", filepath.Join(configRoot, "monologue"))
	t.Setenv("MONOLOGUE_API_BASE_URL", "")
	t.Setenv("MONOLOGUE_API_TOKEN", "")

	path, err := Save(StoredConfig{
		BaseURL: "https://example.monologue.test",
		Token:   "mono_pat_saved",
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	expectedPath := filepath.Join(configRoot, "monologue", "config.json")
	if path != expectedPath {
		t.Fatalf("unexpected path: got %q want %q", path, expectedPath)
	}

	cfg, err := Load("", "")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.BaseURL != "https://example.monologue.test" {
		t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
	}
	if cfg.Token != "mono_pat_saved" {
		t.Fatalf("unexpected token: %q", cfg.Token)
	}
	if cfg.Path != expectedPath {
		t.Fatalf("unexpected config path: %q", cfg.Path)
	}
}

func TestLoadPrefersFlagsThenEnvironmentThenStoredConfig(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("MONOLOGUE_CONFIG_DIR", filepath.Join(configRoot, "monologue"))
	if _, err := Save(StoredConfig{
		BaseURL: "https://stored.monologue.test",
		Token:   "mono_pat_stored",
	}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	t.Setenv("MONOLOGUE_API_BASE_URL", "https://env.monologue.test")
	t.Setenv("MONOLOGUE_API_TOKEN", "mono_pat_env")

	cfg, err := Load("https://flag.monologue.test", "mono_pat_flag")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.BaseURL != "https://flag.monologue.test" {
		t.Fatalf("unexpected base URL: %q", cfg.BaseURL)
	}
	if cfg.Token != "mono_pat_flag" {
		t.Fatalf("unexpected token: %q", cfg.Token)
	}
}
