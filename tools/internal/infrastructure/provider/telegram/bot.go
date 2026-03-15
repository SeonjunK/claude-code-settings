package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/session"
)

// Bot is a long-polling Telegram bot for status queries.
type Bot struct {
	client     *Client
	chatID     string
	chatIDInt  int64
	sessDir    string
	projectDir string
	offset     int
}

// NewBot creates a new Telegram bot.
func NewBot(token, chatID, sessDir, projectDir string) *Bot {
	chatIDInt, _ := strconv.ParseInt(chatID, 10, 64)
	return &Bot{
		client:     NewClientWithLongPoll(token, 30*time.Second),
		chatID:     chatID,
		chatIDInt:  chatIDInt,
		sessDir:    sessDir,
		projectDir: projectDir,
	}
}

// Run starts the long-polling loop. Blocks until ctx is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	fmt.Fprintf(os.Stderr, "[telegram-bot] Started (chat_id=%s)\n", b.chatID)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "[telegram-bot] Shutting down...")
			return ctx.Err()
		default:
			updates, err := b.client.GetUpdates(ctx, b.offset, 30)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				fmt.Fprintf(os.Stderr, "[telegram-bot] getUpdates error: %v\n", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, u := range updates {
				b.offset = u.UpdateID + 1
				b.handleUpdate(ctx, u)
			}
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	// Only respond to the configured chat
	if update.Message.Chat.ID != b.chatIDInt {
		return
	}

	text := strings.TrimSpace(update.Message.Text)
	cmd := strings.Split(text, " ")[0]

	switch cmd {
	case "/status":
		b.handleStatus(ctx)
	case "/tasks":
		b.handleTasks(ctx)
	case "/stop":
		b.handleStop(ctx)
	case "/help", "/start":
		b.handleHelp(ctx)
	}
}

func (b *Bot) handleStatus(ctx context.Context) {
	sess, _, err := b.findActiveSession()
	if err != nil {
		_ = b.client.SendMessage(ctx, b.chatID, FormatNoActiveSession(), "HTML")
		return
	}

	metrics := &notification.MetricsSnapshot{
		TasksCompleted:  sess.Metrics.TasksCompleted,
		TasksPending:    sess.Metrics.TasksPending,
		TasksInProgress: sess.Metrics.TasksInProgress,
		TotalToolCalls:  sess.Metrics.TotalToolCalls,
	}

	var teammates []TeammateInfo
	for _, t := range sess.Teammates {
		teammates = append(teammates, TeammateInfo{Name: t.Name, Status: t.Status})
	}

	text := FormatSessionStatus(
		sess.SessionID, sess.TeamName, sess.Goal.Description,
		sess.Active, sess.Iteration, sess.MaxIterations,
		metrics, teammates,
	)

	keyboard := [][]InlineKeyboardButton{
		{
			{Text: "\U0001F504 Refresh", CallbackData: "refresh_status"},
			{Text: "\u23F9 Stop Session", CallbackData: "stop_session"},
		},
	}

	_ = b.client.SendMessageWithKeyboard(ctx, b.chatID, text, "HTML", keyboard)
}

func (b *Bot) handleTasks(ctx context.Context) {
	sess, _, err := b.findActiveSession()
	if err != nil {
		_ = b.client.SendMessage(ctx, b.chatID, FormatNoActiveSession(), "HTML")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>\U0001F4CB Task Summary</b>\n\n")
	fmt.Fprintf(&sb, "<b>Completed:</b> %d\n", sess.Metrics.TasksCompleted)
	fmt.Fprintf(&sb, "<b>In Progress:</b> %d\n", sess.Metrics.TasksInProgress)
	fmt.Fprintf(&sb, "<b>Pending:</b> %d\n", sess.Metrics.TasksPending)
	fmt.Fprintf(&sb, "<b>Total Tool Calls:</b> %d\n", sess.Metrics.TotalToolCalls)

	if len(sess.Teammates) > 0 {
		fmt.Fprintf(&sb, "\n<b>Teammate Status:</b>\n")
		for _, t := range sess.Teammates {
			fmt.Fprintf(&sb, "  \u2022 %s (%s): %s\n", t.Name, t.SubagentType, t.Status)
		}
	}

	_ = b.client.SendMessage(ctx, b.chatID, sb.String(), "HTML")
}

func (b *Bot) handleStop(ctx context.Context) {
	sess, path, err := b.findActiveSession()
	if err != nil {
		_ = b.client.SendMessage(ctx, b.chatID, FormatNoActiveSession(), "HTML")
		return
	}

	sess.Active = false
	if err := session.SaveSession(path, sess); err != nil {
		_ = b.client.SendMessage(ctx, b.chatID,
			fmt.Sprintf("\u274C Failed to stop session: %v", err), "HTML")
		return
	}

	_ = b.client.SendMessage(ctx, b.chatID,
		fmt.Sprintf("\u2705 Session <code>%s</code> stopped.", shortBotID(sess.SessionID)), "HTML")
}

func (b *Bot) handleHelp(ctx context.Context) {
	_ = b.client.SendMessage(ctx, b.chatID, FormatHelp(), "HTML")
}

func (b *Bot) handleCallback(ctx context.Context, cb *CallbackQuery) {
	_ = b.client.AnswerCallbackQuery(ctx, cb.ID, "")

	switch cb.Data {
	case "refresh_status":
		b.handleStatus(ctx)
	case "stop_session":
		b.handleStop(ctx)
	}
}

func (b *Bot) findActiveSession() (*session.Session, string, error) {
	entries, err := os.ReadDir(b.sessDir)
	if err != nil {
		return nil, "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".local.md") {
			path := filepath.Join(b.sessDir, entry.Name())
			sess, err := session.LoadSession(path)
			if err != nil {
				continue
			}
			if sess.Active {
				return sess, path, nil
			}
		}
	}

	return nil, "", fmt.Errorf("no active session")
}

func shortBotID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
