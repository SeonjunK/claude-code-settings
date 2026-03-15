// Package hook provides hook handlers for Claude Code hooks.
package hook

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/config"
)

// CheckFileAccess checks whether a file access should be blocked.
// mode should be "read" or "write".
// Returns a non-empty block message if access should be denied, or "" to allow.
func CheckFileAccess(guard *config.GuardConfig, mode, filePath string) string {
	var patterns []string
	switch mode {
	case "read":
		patterns = guard.Read.BlockedPatterns
	case "write":
		patterns = guard.Write.BlockedPatterns
	default:
		return ""
	}

	baseName := filepath.Base(filePath)

	// Try basename match first (e.g. ".env*", "*.key")
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, baseName)
		if matched {
			return fmt.Sprintf("File access blocked: %s (matched %s)", baseName, pattern)
		}
	}

	// Try relative path match (e.g. "secrets/*")
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filePath)
		if matched {
			return fmt.Sprintf("File access blocked: %s (matched %s)", baseName, pattern)
		}
	}

	return ""
}

// CheckBashCommand checks whether a bash command should be blocked.
// Returns a non-empty block message if the command should be denied, or "" to allow.
func CheckBashCommand(guard *config.GuardConfig, command string) string {
	trimmed := strings.TrimSpace(command)

	// Check exact match against blockedCommands
	for _, blocked := range guard.Bash.BlockedCommands {
		if trimmed == blocked {
			return fmt.Sprintf("Command blocked: %s is blocked by guard policy", trimmed)
		}
	}

	// Check contains match against blockedPatterns
	for _, pattern := range guard.Bash.BlockedPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Sprintf("Command blocked: matches pattern %s", pattern)
		}
	}

	return ""
}

// DenyOutput creates a PreToolUse deny output conforming to Claude Code hook schema.
func DenyOutput(message string) map[string]any {
	return map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "deny",
		},
		"systemMessage": message,
	}
}
