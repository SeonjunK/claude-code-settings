package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVibeConfig_FullConfig(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	jsonContent := `{
  "guard": {
    "read": { "blockedPatterns": [".env*", "*.pem"] },
    "write": { "blockedPatterns": [".env*"] },
    "bash": {
      "blockedCommands": ["rm -rf /"],
      "blockedPatterns": ["git push --force"]
    }
  },
  "notifications": {
    "telegram": { "enabled": true, "botToken": "test-token-123", "chatId": "12345" },
    "slack": { "enabled": false, "webhookUrl": "https://hooks.slack.com/test", "channel": "#general" },
    "webhook": { "enabled": true, "url": "https://example.com/hook", "headers": { "Authorization": "Bearer abc" } },
    "events": {
      "filters": {
        "session_start": ["telegram"],
        "guard_deny": ["telegram", "webhook"]
      }
    }
  }
}`
	if err := os.WriteFile(filepath.Join(claudeDir, "vibe.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write vibe.json: %v", err)
	}

	cfg, err := LoadVibeConfig(dir)
	if err != nil {
		t.Fatalf("LoadVibeConfig() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadVibeConfig() returned nil config")
	}

	// Guard
	if len(cfg.Guard.Read.BlockedPatterns) != 2 {
		t.Errorf("Guard.Read.BlockedPatterns length = %d, want 2", len(cfg.Guard.Read.BlockedPatterns))
	}
	if len(cfg.Guard.Bash.BlockedCommands) != 1 {
		t.Errorf("Guard.Bash.BlockedCommands length = %d, want 1", len(cfg.Guard.Bash.BlockedCommands))
	}

	// Telegram
	n := cfg.Notifications
	if !n.Telegram.Enabled {
		t.Error("Telegram.Enabled = false, want true")
	}
	if n.Telegram.BotToken != "test-token-123" {
		t.Errorf("Telegram.BotToken = %q, want %q", n.Telegram.BotToken, "test-token-123")
	}
	if n.Telegram.ChatID != "12345" {
		t.Errorf("Telegram.ChatID = %q, want %q", n.Telegram.ChatID, "12345")
	}

	// Slack
	if n.Slack.Enabled {
		t.Error("Slack.Enabled = true, want false")
	}

	// Webhook
	if !n.Webhook.Enabled {
		t.Error("Webhook.Enabled = false, want true")
	}
	if n.Webhook.URL != "https://example.com/hook" {
		t.Errorf("Webhook.URL = %q, want %q", n.Webhook.URL, "https://example.com/hook")
	}
	if n.Webhook.Headers["Authorization"] != "Bearer abc" {
		t.Errorf("Webhook.Headers[Authorization] = %q, want %q", n.Webhook.Headers["Authorization"], "Bearer abc")
	}

	// Events filters
	if providers, ok := n.Events.Filters["session_start"]; !ok {
		t.Error("Events.Filters missing session_start")
	} else if len(providers) != 1 || providers[0] != "telegram" {
		t.Errorf("Events.Filters[session_start] = %v, want [telegram]", providers)
	}
	if providers, ok := n.Events.Filters["guard_deny"]; !ok {
		t.Error("Events.Filters missing guard_deny")
	} else if len(providers) != 2 {
		t.Errorf("Events.Filters[guard_deny] length = %d, want 2", len(providers))
	}
}

func TestLoadVibeConfig_EnvVarInterpolation(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	jsonContent := `{
  "notifications": {
    "telegram": {
      "enabled": true,
      "botToken": "${VIBE_TEST_BOT_TOKEN}",
      "chatId": "${VIBE_TEST_CHAT_ID}"
    }
  }
}`
	if err := os.WriteFile(filepath.Join(claudeDir, "vibe.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write vibe.json: %v", err)
	}

	t.Setenv("VIBE_TEST_BOT_TOKEN", "env-token-value")
	t.Setenv("VIBE_TEST_CHAT_ID", "env-chat-999")

	cfg, err := LoadVibeConfig(dir)
	if err != nil {
		t.Fatalf("LoadVibeConfig() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadVibeConfig() returned nil config")
	}

	if cfg.Notifications.Telegram.BotToken != "env-token-value" {
		t.Errorf("Telegram.BotToken = %q, want %q (from env)", cfg.Notifications.Telegram.BotToken, "env-token-value")
	}
	if cfg.Notifications.Telegram.ChatID != "env-chat-999" {
		t.Errorf("Telegram.ChatID = %q, want %q (from env)", cfg.Notifications.Telegram.ChatID, "env-chat-999")
	}
}

func TestLoadVibeConfig_MissingFileReturnsNil(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadVibeConfig(dir)
	if err != nil {
		t.Fatalf("LoadVibeConfig() error: %v", err)
	}
	if cfg != nil {
		t.Errorf("LoadVibeConfig() = %v, want nil for missing file", cfg)
	}
}

func TestLoadVibeConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(claudeDir, "vibe.json"), []byte("{invalid json:::"), 0o644); err != nil {
		t.Fatalf("failed to write vibe.json: %v", err)
	}

	cfg, err := LoadVibeConfig(dir)
	if err == nil {
		t.Fatal("LoadVibeConfig() expected error for invalid JSON, got nil")
	}
	if cfg != nil {
		t.Errorf("LoadVibeConfig() returned non-nil config on error")
	}
}

func TestLoadVibeConfig_RootPath(t *testing.T) {
	dir := t.TempDir()

	// Place vibe.json at root (not .claude/)
	jsonContent := `{"guard": {"read": {"blockedPatterns": ["*.key"]}}}`
	if err := os.WriteFile(filepath.Join(dir, "vibe.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write vibe.json: %v", err)
	}

	cfg, err := LoadVibeConfig(dir)
	if err != nil {
		t.Fatalf("LoadVibeConfig() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadVibeConfig() returned nil for root-level vibe.json")
	}
	if len(cfg.Guard.Read.BlockedPatterns) != 1 {
		t.Errorf("Guard.Read.BlockedPatterns length = %d, want 1", len(cfg.Guard.Read.BlockedPatterns))
	}
}

func TestVibeConfig_HasAnyProvider(t *testing.T) {
	tests := []struct {
		name string
		cfg  VibeConfig
		want bool
	}{
		{name: "no providers", cfg: VibeConfig{}, want: false},
		{name: "telegram", cfg: VibeConfig{Notifications: NotificationsConfig{Telegram: TelegramConfig{Enabled: true}}}, want: true},
		{name: "slack", cfg: VibeConfig{Notifications: NotificationsConfig{Slack: SlackConfig{Enabled: true}}}, want: true},
		{name: "webhook", cfg: VibeConfig{Notifications: NotificationsConfig{Webhook: WebhookConfig{Enabled: true}}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.HasAnyProvider(); got != tt.want {
				t.Errorf("HasAnyProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVibeConfig_HasGuard(t *testing.T) {
	tests := []struct {
		name string
		cfg  VibeConfig
		want bool
	}{
		{name: "empty", cfg: VibeConfig{}, want: false},
		{name: "read patterns", cfg: VibeConfig{Guard: GuardConfig{Read: struct {
			BlockedPatterns []string `json:"blockedPatterns"`
		}{BlockedPatterns: []string{".env*"}}}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.HasGuard(); got != tt.want {
				t.Errorf("HasGuard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVibeConfig_ProvidersForEvent(t *testing.T) {
	cfg := VibeConfig{
		Notifications: NotificationsConfig{
			Telegram: TelegramConfig{Enabled: true},
			Slack:    SlackConfig{Enabled: true},
			Events: EventsConfig{
				Filters: map[string][]string{
					"session_start": {"telegram"},
				},
			},
		},
	}

	// Filtered event
	got := cfg.ProvidersForEvent("session_start")
	if len(got) != 1 || got[0] != "telegram" {
		t.Errorf("ProvidersForEvent(session_start) = %v, want [telegram]", got)
	}

	// Unfiltered event falls back to all enabled
	got = cfg.ProvidersForEvent("guard_deny")
	if len(got) != 2 {
		t.Errorf("ProvidersForEvent(guard_deny) = %v, want [telegram slack]", got)
	}
}
