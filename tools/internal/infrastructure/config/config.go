// Package config provides configuration management using Viper.
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration.
type Config struct {
	ProjectDir  string `mapstructure:"project_dir"`
	SessionsDir string `mapstructure:"sessions_dir"`
	SessionID   string `mapstructure:"session_id"`
}

// Load reads configuration from environment variables and defaults.
func Load() *Config {
	v := viper.New()

	// Environment variable bindings
	v.BindEnv("project_dir", "CLAUDE_PROJECT_DIR")
	v.BindEnv("session_id", "CLAUDE_CODE_SESSION_ID")

	// Set defaults
	cwd, _ := os.Getwd()
	v.SetDefault("project_dir", cwd)

	// Read config
	projectDir := v.GetString("project_dir")
	v.SetDefault("sessions_dir", filepath.Join(projectDir, ".claude", "sessions"))

	return &Config{
		ProjectDir:  v.GetString("project_dir"),
		SessionsDir: v.GetString("sessions_dir"),
		SessionID:   v.GetString("session_id"),
	}
}

// GetSessionPath returns the path to a session file.
func (c *Config) GetSessionPath(sessionID string) string {
	if sessionID == "" {
		sessionID = c.SessionID
	}
	return filepath.Join(c.SessionsDir, sessionID+".local.md")
}

// GetActiveSessionPath returns the path to the current active session.
func (c *Config) GetActiveSessionPath() string {
	return c.GetSessionPath("")
}

// EnsureSessionsDir ensures the sessions directory exists.
func (c *Config) EnsureSessionsDir() error {
	return os.MkdirAll(c.SessionsDir, 0755)
}
