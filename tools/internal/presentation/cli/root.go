// Package cli provides CLI commands using Cobra.
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/infrastructure/config"
)

var (
	cfg *config.Config
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "claude-code-hooks",
	Short: "Claude Code hooks CLI for session management",
	Long: `Claude Code hooks CLI provides hook handlers and utilities
for managing Claude Code sessions and team orchestration.

Hook commands (called by Claude Code):
  claude-code-hooks stop                Handle stop hook
  claude-code-hooks user-prompt-submit Handle user-prompt-submit hook
  claude-code-hooks post-tool-use      Handle post-tool-use hook (formatting)

Utility commands:
  claude-code-hooks start <goal>      Start a new session
  claude-code-hooks status            Show session status
  claude-code-hooks stop-session      Stop current session
  claude-code-hooks verify            Run verification`,
}

// Execute runs the CLI.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cfg = config.Load()

	// Ensure sessions directory exists
	if err := cfg.EnsureSessionsDir(); err != nil {
		// Non-fatal, will be created on demand
	}
}
