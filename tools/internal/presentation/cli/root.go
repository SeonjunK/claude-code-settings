// Package cli provides CLI commands using Cobra.
package cli

import (
	"path/filepath"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/notifier"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/config"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/logger"
)

// Deps holds shared dependencies for all CLI commands.
type Deps struct {
	Cfg      *config.Config
	VibeConf *config.VibeConfig
	Notif    *notifier.Notifier
	Log      *logger.Logger
}

// InitDeps creates and initializes shared dependencies.
func InitDeps() *Deps {
	cfg := config.Load()

	// Ensure sessions directory exists
	_ = cfg.EnsureSessionsDir()

	// Initialize logger: .claude/logs/{session-id}.log
	logsDir := filepath.Join(cfg.ProjectDir, ".claude", "logs")
	log, err := logger.New(logsDir, cfg.SessionID)
	if err != nil {
		log = logger.Nop()
	}

	// Load unified vibe.json config (optional)
	vibeConf, _ := config.LoadVibeConfig(cfg.ProjectDir)

	var n *notifier.Notifier
	if vibeConf != nil && vibeConf.HasAnyProvider() {
		n = buildNotifier(vibeConf)
	}

	log.Info("deps initialized",
		"project_dir", cfg.ProjectDir,
		"session_id", cfg.SessionID,
	)

	return &Deps{
		Cfg:      cfg,
		VibeConf: vibeConf,
		Notif:    n,
		Log:      log,
	}
}

// Close releases resources held by Deps.
func (d *Deps) Close() {
	if d.Log != nil {
		d.Log.Close()
	}
}
