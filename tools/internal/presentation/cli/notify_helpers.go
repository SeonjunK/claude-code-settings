package cli

import (
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/notifier"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/session"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/config"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider/slack"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider/telegram"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider/webhook"
)

// buildNotifier creates a Notifier from vibe config.
// Returns nil if no providers are enabled.
func buildNotifier(vc *config.VibeConfig) *notifier.Notifier {
	if vc == nil || !vc.HasAnyProvider() {
		return nil
	}

	reg := provider.NewRegistry()

	n := vc.Notifications

	if n.Telegram.Enabled && n.Telegram.BotToken != "" && n.Telegram.ChatID != "" {
		reg.Add(telegram.NewSender(n.Telegram.BotToken, n.Telegram.ChatID))
	}

	if n.Slack.Enabled && n.Slack.WebhookURL != "" {
		reg.Add(slack.NewWebhookSender(n.Slack.WebhookURL, n.Slack.Channel))
	}

	if n.Webhook.Enabled && n.Webhook.URL != "" {
		reg.Add(webhook.NewGenericSender(n.Webhook.URL, n.Webhook.Headers))
	}

	if len(reg.Providers()) == 0 {
		return nil
	}

	return notifier.New(reg, vc.Notifications.Events.Filters)
}

// enrichEvent fills session data into a notification event.
func enrichEvent(event *notification.Event, sess *session.Session) {
	event.Goal = sess.Goal.Description
	event.Iteration = sess.Iteration
	event.MaxIter = sess.MaxIterations
	event.TeamName = sess.TeamName
	event.Metrics = &notification.MetricsSnapshot{
		TasksCompleted:  sess.Metrics.TasksCompleted,
		TasksPending:    sess.Metrics.TasksPending,
		TasksInProgress: sess.Metrics.TasksInProgress,
		TotalToolCalls:  sess.Metrics.TotalToolCalls,
	}
}

// dispatchAndWait dispatches a notification event and waits for completion.
// Safe to call when notif is nil.
func dispatchAndWait(n *notifier.Notifier, event notification.Event) {
	if n == nil {
		return
	}
	wait := n.Dispatch(event)
	wait()
}

// newEvent creates a new event with common fields filled.
func newEvent(sessionID string, eventType notification.EventType, summary string) notification.Event {
	return notification.Event{
		Type:      eventType,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Summary:   summary,
	}
}

// loadSessionFromPath loads a session from the given path.
func loadSessionFromPath(path string) (*session.Session, error) {
	return session.LoadSession(path)
}
