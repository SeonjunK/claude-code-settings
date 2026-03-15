// Package cli provides CLI commands using Cobra.
package cli

import (
	"github.com/SeonjunK/claude-code-settings/tools/internal/application/notifier"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/config"
)

// Deps holds shared dependencies for all CLI commands.
type Deps struct {
	Cfg      *config.Config
	VibeConf *config.VibeConfig
	Notif    *notifier.Notifier
}

// InitDeps creates and initializes shared dependencies.
func InitDeps() *Deps {
	cfg := config.Load()

	// Ensure sessions directory exists
	_ = cfg.EnsureSessionsDir()

	// Load unified vibe.json config (optional)
	vibeConf, _ := config.LoadVibeConfig(cfg.ProjectDir)

	var n *notifier.Notifier
	if vibeConf != nil && vibeConf.HasAnyProvider() {
		n = buildNotifier(vibeConf)
	}

	return &Deps{
		Cfg:      cfg,
		VibeConf: vibeConf,
		Notif:    n,
	}
}
