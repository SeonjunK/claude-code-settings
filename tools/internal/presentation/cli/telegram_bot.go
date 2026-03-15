package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/provider/telegram"
)

// NewTelegramBotCmd creates the telegram-bot command.
func NewTelegramBotCmd(deps *Deps) *cobra.Command {
	var (
		token  string
		chatID string
	)

	cmd := &cobra.Command{
		Use:   "telegram-bot",
		Short: "Start Telegram bot (long-polling)",
		Long: `Start a long-running Telegram bot that responds to commands.

Commands:
  /status  — Current session status
  /tasks   — Task list summary
  /stop    — Stop active session
  /help    — Show available commands

The bot reads session state from .claude/sessions/ files.
Configure botToken and chatId in vibe.json.

Examples:
  vibe-loops telegram-bot
  vibe-loops telegram-bot --token $BOT_TOKEN --chat-id $CHAT_ID`,
		Run: func(cmd *cobra.Command, args []string) {
			deps.Log.Info("telegram-bot: starting")

			t := token
			c := chatID

			if t == "" && deps.VibeConf != nil {
				t = deps.VibeConf.Notifications.Telegram.BotToken
			}
			if c == "" && deps.VibeConf != nil {
				c = deps.VibeConf.Notifications.Telegram.ChatID
			}

			if t == "" || c == "" {
				deps.Log.Error("telegram-bot: missing token or chat_id")
				fmt.Fprintln(os.Stderr, "Telegram botToken and chatId are required.")
				fmt.Fprintln(os.Stderr, "Set them in vibe.json or use --token and --chat-id flags.")
				os.Exit(1)
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			deps.Log.Info("telegram-bot: running", "chat_id", c)
			bot := telegram.NewBot(t, c, deps.Cfg.SessionsDir, deps.Cfg.ProjectDir)
			if err := bot.Run(ctx); err != nil && err != context.Canceled {
				deps.Log.Error("telegram-bot: error", "err", err)
				fmt.Fprintf(os.Stderr, "Bot error: %v\n", err)
				os.Exit(1)
			}
			deps.Log.Info("telegram-bot: stopped")
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Telegram bot token (overrides vibe.json)")
	cmd.Flags().StringVar(&chatID, "chat-id", "", "Telegram chat ID (overrides vibe.json)")
	return cmd
}
