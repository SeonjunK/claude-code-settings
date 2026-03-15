// Package slack provides Slack incoming webhook notification.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// WebhookSender implements provider.Provider for Slack incoming webhooks.
type WebhookSender struct {
	webhookURL string
	channel    string
	client     *http.Client
}

// NewWebhookSender creates a Slack webhook sender.
func NewWebhookSender(webhookURL, channel string) *WebhookSender {
	return &WebhookSender{
		webhookURL: webhookURL,
		channel:    channel,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Name returns the provider name.
func (s *WebhookSender) Name() string {
	return "slack"
}

// Send formats and delivers the event via Slack webhook.
func (s *WebhookSender) Send(ctx context.Context, event notification.Event) error {
	blocks := buildBlocks(event)

	payload := map[string]any{
		"blocks": blocks,
	}
	if s.channel != "" {
		payload["channel"] = s.channel
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func buildBlocks(event notification.Event) []any {
	header := map[string]any{
		"type": "header",
		"text": map[string]any{
			"type":  "plain_text",
			"text":  event.Type.Emoji() + " " + event.Type.Label(),
			"emoji": true,
		},
	}

	var fields []any
	if event.SessionID != "" {
		fields = append(fields, mrkdwnField("*Session*", shortID(event.SessionID)))
	}
	if event.Goal != "" {
		fields = append(fields, mrkdwnField("*Goal*", event.Goal))
	}
	if event.MaxIter > 0 {
		fields = append(fields, mrkdwnField("*Iteration*", fmt.Sprintf("%d/%d", event.Iteration, event.MaxIter)))
	}
	if event.Metrics != nil {
		m := event.Metrics
		fields = append(fields, mrkdwnField("*Tasks*",
			fmt.Sprintf("%d done, %d active, %d pending", m.TasksCompleted, m.TasksInProgress, m.TasksPending)))
	}

	section := map[string]any{
		"type": "section",
		"text": map[string]any{
			"type": "mrkdwn",
			"text": event.Summary,
		},
	}
	if len(fields) > 0 {
		section["fields"] = fields
	}

	context := map[string]any{
		"type": "context",
		"elements": []any{
			map[string]any{
				"type": "mrkdwn",
				"text": fmt.Sprintf("_%s_", event.Timestamp.Format(time.RFC3339)),
			},
		},
	}

	return []any{header, section, context}
}

func mrkdwnField(label, value string) any {
	return map[string]any{
		"type": "mrkdwn",
		"text": fmt.Sprintf("%s\n%s", label, value),
	}
}

func shortID(id string) string {
	if len(id) > 8 {
		return "`" + id[:8] + "`"
	}
	return "`" + id + "`"
}
