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

Reads hook input from stdin and outputs JSON response.
Exit codes:
  0 - allow command
  1 - block command
  2 - error`,
		Run: func(cmd *cobra.Command, args []string) {
			// Read input from stdin
			input, err := storage.ReadStdin()
			if err != nil {
				fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to read stdin"))
				os.Exit(2)
			}

			// Parse input
			parsed, err := hook.ParseUserPromptSubmitInput(input)
			if err != nil {
				fmt.Println(hook.AllowOutput())
				os.Exit(0)
			}

			// Use session ID from config if not in input
			if parsed.SessionID == "" {
				parsed.SessionID = deps.Cfg.SessionID
			}

			// Handle using application layer
			handler := hook.NewUserPromptSubmitHandler(deps.Cfg.SessionsDir)
			output, err := handler.Handle(parsed, generateSessionID)
			if err != nil {
				fmt.Fprintln(os.Stderr, hook.ErrorOutput(err.Error()))
				os.Exit(2)
			}

			// Output JSON
			jsonOutput, err := output.ToJSON()
			if err != nil {
				fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to serialize output"))
				os.Exit(2)
			}

			fmt.Println(string(jsonOutput))

			// Notify if this was a /team-loops command
			if strings.HasPrefix(parsed.Prompt, "/team-loops") {
				event := newEvent(deps.Cfg.SessionID, notification.EventSessionStart, "Team-loops session created")
				event.Details = map[string]string{"prompt": parsed.Prompt}
				dispatchAndWait(deps.Notif, event)
			}

			// Exit with appropriate code
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
