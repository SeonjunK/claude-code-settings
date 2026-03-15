// Package hook provides hook handlers for Claude Code hooks.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GuardConfig represents the guard.json configuration.
type GuardConfig struct {
	Read struct {
		BlockedPatterns []string `json:"blockedPatterns"`
	} `json:"read"`
	Write struct {
		BlockedPatterns []string `json:"blockedPatterns"`
	} `json:"write"`
	Bash struct {
		BlockedCommands []string `json:"blockedCommands"`
		BlockedPatterns []string `json:"blockedPatterns"`
	} `json:"bash"`
}

// LoadGuardConfig loads guard configuration from the project directory.
// Looks for guard.json in:
//  1. $projectDir/.claude/guard.json (legacy path)
//  2. $projectDir/guard.json (plugin path)
func LoadGuardConfig(projectDir string) (*GuardConfig, error) {
	paths := []string{
		filepath.Join(projectDir, ".claude", "guard.json"),
		filepath.Join(projectDir, "guard.json"),
	}

	var data []byte
	var lastErr error
	for _, path := range paths {
		data, lastErr = os.ReadFile(path)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("guard.json not found in %s: %w", projectDir, lastErr)
	}

	var cfg GuardConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse guard.json: %w", err)
	}
	return &cfg, nil
}

// CheckFileAccess checks whether a file access should be blocked.
// mode should be "read" or "write".
// Returns a non-empty block message if access should be denied, or "" to allow.
func CheckFileAccess(cfg *GuardConfig, mode, filePath string) string {
	var patterns []string
	switch mode {
	case "read":
		patterns = cfg.Read.BlockedPatterns
	case "write":
		patterns = cfg.Write.BlockedPatterns
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
func CheckBashCommand(cfg *GuardConfig, command string) string {
	trimmed := strings.TrimSpace(command)

	// Check exact match against blockedCommands
	for _, blocked := range cfg.Bash.BlockedCommands {
		if trimmed == blocked {
			return fmt.Sprintf("Command blocked: %s is blocked by guard policy", trimmed)
		}
	}

	// Check contains match against blockedPatterns
	for _, pattern := range cfg.Bash.BlockedPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Sprintf("Command blocked: matches pattern %s", pattern)
		}
	}

	return ""
}

// DenyOutput creates a PreToolUse deny output conforming to Claude Code hook schema.
func DenyOutput(message string) map[string]interface{} {
	return map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "deny",
		},
		"systemMessage": message,
	}
}
