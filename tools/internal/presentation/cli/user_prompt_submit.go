package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/hook"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/storage"
)

// NewUserPromptSubmitCmd creates the user-prompt-submit hook command.
func NewUserPromptSubmitCmd(deps *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "user-prompt-submit",
		Short: "Handle user-prompt-submit hook",
		Long: `Handle user-prompt-submit hook when called from a Claude Code hook.

This command is triggered when a user submits a prompt.
It checks if the prompt is a /team-loops command and initializes the session.

Exit codes:
  0 - allow command
  1 - block command
  2 - error`,
		Run: func(cmd *cobra.Command, args []string) {
			deps.Log.Debug("user-prompt-submit: command started")

			input, err := storage.ReadStdin()
			if err != nil {
				deps.Log.Error("user-prompt-submit: stdin read failed", "err", err)
				fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to read stdin"))
				os.Exit(2)
			}

			parsed, err := hook.ParseUserPromptSubmitInput(input)
			if err != nil {
				deps.Log.Warn("user-prompt-submit: parse failed, allowing", "err", err)
				fmt.Println(hook.AllowOutput())
				os.Exit(0)
			}

			if parsed.SessionID == "" {
				parsed.SessionID = deps.Cfg.SessionID
			}

			deps.Log.Debug("user-prompt-submit: parsed",
				"session_id", parsed.SessionID,
				"prompt_prefix", truncate(parsed.Prompt, 80),
			)

			handler := hook.NewUserPromptSubmitHandler(deps.Cfg.SessionsDir)
			output, err := handler.Handle(parsed, generateSessionID)
			if err != nil {
				deps.Log.Error("user-prompt-submit: handler failed", "err", err)
				fmt.Fprintln(os.Stderr, hook.ErrorOutput(err.Error()))
				os.Exit(2)
			}

			jsonOutput, err := output.ToJSON()
			if err != nil {
				deps.Log.Error("user-prompt-submit: marshal failed", "err", err)
				fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to serialize output"))
				os.Exit(2)
			}

			fmt.Println(string(jsonOutput))

			deps.Log.Info("user-prompt-submit: result", "action", output.Action)

			// Notify if this was a /team-loops command
			if strings.HasPrefix(parsed.Prompt, "/team-loops") {
				event := newEvent(deps.Cfg.SessionID, notification.EventSessionStart, "Team-loops session created")
				event.Details = map[string]string{"prompt": parsed.Prompt}
				dispatchAndWait(deps.Notif, event)
				deps.Log.Info("user-prompt-submit: team-loops notification sent")
			}

			switch output.Action {
			case "block":
				os.Exit(1)
			case "allow":
				os.Exit(0)
			default:
				os.Exit(2)
			}
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
