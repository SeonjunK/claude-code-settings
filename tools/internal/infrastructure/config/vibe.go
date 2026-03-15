package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VibeConfig represents the unified vibe.json configuration.
type VibeConfig struct {
	Guard         GuardConfig         `json:"guard"`
	Notifications NotificationsConfig `json:"notifications"`
}

// GuardConfig holds file and command access control rules.
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

// NotificationsConfig holds notification provider settings.
type NotificationsConfig struct {
	Telegram TelegramConfig `json:"telegram"`
	Slack    SlackConfig    `json:"slack"`
	Webhook  WebhookConfig  `json:"webhook"`
	Events   EventsConfig   `json:"events"`
}

// TelegramConfig holds Telegram notification settings.
type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}

// SlackConfig holds Slack notification settings.
type SlackConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl"`
	Channel    string `json:"channel"`
}

// WebhookConfig holds generic webhook settings.
type WebhookConfig struct {
	Enabled bool              `json:"enabled"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// EventsConfig holds event routing rules.
type EventsConfig struct {
	Filters map[string][]string `json:"filters"`
}

// LoadVibeConfig loads vibe.json from the project directory.
// Search order:
//  1. $projectDir/.claude/vibe.json
//  2. $projectDir/vibe.json
//
// Returns nil, nil if the file does not exist (not an error).
func LoadVibeConfig(projectDir string) (*VibeConfig, error) {
	paths := []string{
		filepath.Join(projectDir, ".claude", "vibe.json"),
		filepath.Join(projectDir, "vibe.json"),
	}

	var raw []byte
	var lastErr error
	for _, path := range paths {
		raw, lastErr = os.ReadFile(path)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		if os.IsNotExist(lastErr) {
			return nil, nil
		}
		return nil, nil
	}

	// Interpolate environment variables before parsing.
	expanded := os.ExpandEnv(string(raw))

	var cfg VibeConfig
	if err := json.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse vibe.json: %w", err)
	}

	return &cfg, nil
}

// HasGuard returns true if any guard rules are configured.
func (c *VibeConfig) HasGuard() bool {
	return len(c.Guard.Read.BlockedPatterns) > 0 ||
		len(c.Guard.Write.BlockedPatterns) > 0 ||
		len(c.Guard.Bash.BlockedCommands) > 0 ||
		len(c.Guard.Bash.BlockedPatterns) > 0
}

// HasAnyProvider returns true if at least one notification provider is enabled.
func (c *VibeConfig) HasAnyProvider() bool {
	n := c.Notifications
	return n.Telegram.Enabled || n.Slack.Enabled || n.Webhook.Enabled
}

// EnabledProviderNames returns the names of enabled providers.
func (c *VibeConfig) EnabledProviderNames() []string {
	n := c.Notifications
	var names []string
	if n.Telegram.Enabled {
		names = append(names, "telegram")
	}
	if n.Slack.Enabled {
		names = append(names, "slack")
	}
	if n.Webhook.Enabled {
		names = append(names, "webhook")
	}
	return names
}

// ProvidersForEvent returns the provider names configured for a given event type.
// If no filters are configured, returns all enabled providers.
func (c *VibeConfig) ProvidersForEvent(eventType string) []string {
	if c.Notifications.Events.Filters != nil {
		if providers, ok := c.Notifications.Events.Filters[eventType]; ok {
			return providers
		}
	}
	return c.EnabledProviderNames()
}
