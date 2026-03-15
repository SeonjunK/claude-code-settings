package telegram

import (
	"context"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// Sender implements provider.Provider for Telegram push notifications.
type Sender struct {
	client *Client
	chatID string
}

// NewSender creates a Telegram notification sender.
func NewSender(botToken, chatID string) *Sender {
	return &Sender{
		client: NewClient(botToken),
		chatID: chatID,
	}
}

// Name returns the provider name.
func (s *Sender) Name() string {
	return "telegram"
}

// Send formats and delivers the event as a Telegram message.
func (s *Sender) Send(ctx context.Context, event notification.Event) error {
	text := FormatEvent(event)
	return s.client.SendMessage(ctx, s.chatID, text, "HTML")
}
