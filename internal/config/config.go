package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultBackend = "json"

type Config struct {
	Profiles      map[string]Profile `yaml:"profiles"`
	ActiveProfile *Profile           `yaml:"-"`
	ProfileName   string             `yaml:"-"`
}

type Profile struct {
	Backend string       `yaml:"backend"`
	SQLite  SQLiteConfig `yaml:"sqlite"`
	JSON    JSONConfig   `yaml:"json"`
	Output  OutputConfig `yaml:"output"`
	Policy  PolicyConfig `yaml:"policy"`
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type JSONConfig struct {
	DataDir string `yaml:"data_dir"`
}

type OutputConfig struct {
	Format string `yaml:"format"`
	Pretty bool   `yaml:"pretty"`
}

type PolicyConfig struct {
	RequireProjectWhenLeavingInbox bool `yaml:"require_project_when_leaving_inbox"`
	RequireContextWhenLeavingInbox bool `yaml:"require_context_when_leaving_inbox"`
	AutoNextOnMoveFromInbox        bool `yaml:"auto_next_on_move_from_inbox"`
	AutoNextOnInboxProcess         bool `yaml:"auto_next_on_inbox_process"`
}

func Defaults() *Config {
	home, _ := os.UserHomeDir()
	defaultProfile := Profile{
		Backend: DefaultBackend,
		SQLite: SQLiteConfig{
			Path: filepath.Join(home, ".local/share/gtd-cli/gtd.sqlite"),
		},
		JSON: JSONConfig{
			DataDir: filepath.Join(home, ".local/share/gtd-cli/jsondb"),
		},
		Output: OutputConfig{
			Format: "json",
			Pretty: false,
		},
		Policy: PolicyConfig{
			RequireProjectWhenLeavingInbox: false,
			RequireContextWhenLeavingInbox: false,
			AutoNextOnMoveFromInbox:        true,
			AutoNextOnInboxProcess:         true,
		},
	}
	return &Config{
		Profiles: map[string]Profile{
			"default": defaultProfile,
		},
		ActiveProfile: &defaultProfile,
		ProfileName:   "default",
	}
}

func (c *Config) ApplyDefaults() {
	if len(c.Profiles) == 0 {
		c.Profiles = map[string]Profile{
			"default": {},
		}
	}

	for name, p := range c.Profiles {
		if p.Backend == "" {
			p.Backend = DefaultBackend
		}
		if p.SQLite.Path == "" {
			home, _ := os.UserHomeDir()
			p.SQLite.Path = filepath.Join(home, ".local/share/gtd-cli/gtd.sqlite")
		}
		if p.JSON.DataDir == "" {
			home, _ := os.UserHomeDir()
			p.JSON.DataDir = filepath.Join(home, ".local/share/gtd-cli/jsondb")
		}
		if p.Output.Format == "" {
			p.Output.Format = "json"
		}
		p.SQLite.Path = expandPath(p.SQLite.Path)
		p.JSON.DataDir = expandPath(p.JSON.DataDir)
		c.Profiles[name] = p
	}
}

func Load(path string, profile ...string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	cfg.ApplyDefaults()

	profileName := "default"
	if len(profile) > 0 && profile[0] != "" {
		profileName = profile[0]
	}
	cfg.ProfileName = profileName

	p, ok := cfg.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}
	cfg.ActiveProfile = &p

	return &cfg, nil
}

func (c *Config) SelectProfile(name string) error {
	p, ok := c.Profiles[name]
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	c.ActiveProfile = &p
	c.ProfileName = name
	return nil
}

func expandPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) Validate() error {
	if c.ActiveProfile == nil {
		return errors.New("no active profile selected")
	}
	if c.ActiveProfile.Backend != "sqlite" && c.ActiveProfile.Backend != "json" {
		return fmt.Errorf("invalid backend %q, must be sqlite or json", c.ActiveProfile.Backend)
	}
	return nil
}
