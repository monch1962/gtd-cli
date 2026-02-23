package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `
profiles:
  default:
    backend: sqlite
    sqlite:
      path: "~/.local/share/gtd-cli/gtd.sqlite"
    json:
      data_dir: "~/.local/share/gtd-cli/jsondb"
    output:
      format: json
      pretty: false
    policy:
      require_project_when_leaving_inbox: false
      require_context_when_leaving_inbox: false
      auto_next_on_move_from_inbox: true
      auto_next_on_inbox_process: true
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ActiveProfile.Backend != "sqlite" {
		t.Errorf("Backend = %s, want sqlite", cfg.ActiveProfile.Backend)
	}

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".local/share/gtd-cli/gtd.sqlite")
	if cfg.ActiveProfile.SQLite.Path != expectedPath {
		t.Errorf("SQLite.Path = %s, want %s", cfg.ActiveProfile.SQLite.Path, expectedPath)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load should fail for missing file")
	}
}

func TestLoad_DefaultProfile(t *testing.T) {
	content := `
profiles:
  default:
    backend: json
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ActiveProfile.Backend != "json" {
		t.Errorf("Backend = %s, want json", cfg.ActiveProfile.Backend)
	}
}

func TestLoad_NamedProfile(t *testing.T) {
	content := `
profiles:
  default:
    backend: sqlite
  work:
    backend: json
    json:
      data_dir: "/work/gtd"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath, "work")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ActiveProfile.Backend != "json" {
		t.Errorf("Backend = %s, want json", cfg.ActiveProfile.Backend)
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if cfg.ActiveProfile.Backend != "sqlite" {
		t.Errorf("Default backend = %s, want sqlite", cfg.ActiveProfile.Backend)
	}
	if cfg.ActiveProfile.Output.Format != "json" {
		t.Errorf("Default format = %s, want json", cfg.ActiveProfile.Output.Format)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.expected {
				t.Errorf("expandPath(%s) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}
