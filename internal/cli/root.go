package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/config"
	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/store"
	jsonfile "github.com/anomalyco/gtd-cli/internal/store/jsonfile"
	"github.com/anomalyco/gtd-cli/internal/store/sqlite"
	"github.com/anomalyco/gtd-cli/internal/util"
)

var (
	cfgFile string
	profile string
	backend string
	dbPath  string
	dataDir string
	format  string
	pretty  bool
	quiet   bool
)

type App struct {
	Config  *config.Config
	Store   store.Store
	Clock   util.Clock
	IDGen   util.IDGenerator
	Version string
	Backend string
}

func (a *App) Policy() core.Policy {
	return core.Policy{
		RequireProjectWhenLeavingInbox: a.Config.ActiveProfile.Policy.RequireProjectWhenLeavingInbox,
		RequireContextWhenLeavingInbox: a.Config.ActiveProfile.Policy.RequireContextWhenLeavingInbox,
		AutoNextOnMoveFromInbox:        a.Config.ActiveProfile.Policy.AutoNextOnMoveFromInbox,
		AutoNextOnInboxProcess:         a.Config.ActiveProfile.Policy.AutoNextOnInboxProcess,
	}
}

var rootCmd = &cobra.Command{
	Use:   "gtd-cli",
	Short: "A Getting Things Done CLI tool",
	Long: `gtd-cli is a command-line tool for managing tasks and projects
using the Getting Things Done (GTD) methodology.

All commands output JSON. Use --help on any command for more information.`,
}

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/gtd-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "config profile to use")
	rootCmd.PersistentFlags().StringVar(&backend, "backend", "", "storage backend (sqlite|json)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "sqlite database path")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "json backend data directory")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "output format (json|ndjson)")
	rootCmd.PersistentFlags().BoolVar(&pretty, "pretty", false, "pretty-print JSON output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", true, "suppress non-JSON output")
}

func initConfig() {
}

func newApp(version string) (*App, error) {
	var cfg *config.Config
	var err error

	if cfgFile != "" {
		cfg, err = config.Load(cfgFile, profile)
		if err != nil {
			cfg = config.Defaults()
		}
	} else {
		cfg = config.Defaults()
	}

	be := backend
	if be == "" {
		be = cfg.ActiveProfile.Backend
	}

	clock := util.RealClock{}
	idGen := util.NewRealIDGenerator(clock)

	app := &App{
		Config:  cfg,
		Clock:   clock,
		IDGen:   idGen,
		Version: version,
		Backend: be,
	}

	switch be {
	case "sqlite":
		path := dbPath
		if path == "" {
			path = cfg.ActiveProfile.SQLite.Path
		}
		app.Store, err = sqlite.New(path, idGen)
		if err != nil {
			return nil, fmt.Errorf("create sqlite store: %w", err)
		}
	case "json":
		dir := dataDir
		if dir == "" {
			dir = cfg.ActiveProfile.JSON.DataDir
		}
		app.Store, err = jsonfile.New(dir, idGen, clock)
		if err != nil {
			return nil, fmt.Errorf("create json store: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown backend: %s", be)
	}

	if err := app.Store.Migrate(context.Background()); err != nil {
		app.Store.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return app, nil
}
