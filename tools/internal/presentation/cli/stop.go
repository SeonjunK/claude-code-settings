package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/hook"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/session"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/storage"
)

// NewStopCmd creates the stop hook command.
func NewStopCmd(deps *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Handle stop hook",
		Long: `Handle stop hook when called from a Claude Code hook.

This command is triggered when a Claude Code session is about to stop.
It manages session iteration and determines whether to continue or complete.

Reads hook input from stdin and outputs JSON response.
Exit codes:
  0 - allow stop
  1 - block stop (session continues)
  2 - error`,
		Run: func(cmd *cobra.Command, args []string) {
			// Read input from stdin
			input, err := storage.ReadStdin()
			if err != nil {
				fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to read stdin"))
				os.Exit(2)
			}

			// Parse input
			parsed, err := hook.ParseStopInput(input)
			if err != nil {
				fmt.Fprintln(os.Stderr, hook.ErrorOutput(err.Error()))
				os.Exit(2)
			}

			// Use session ID from config if not in input
			if parsed.SessionID == "" {
				parsed.SessionID = deps.Cfg.SessionID
			}

			// Handle using application layer
			handler := hook.NewStopHandler(deps.Cfg.SessionsDir)
			output, err := handler.Handle(parsed)
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

			// Notify based on action
			if deps.Notif != nil {
				var event notification.Event
				switch output.Action {
				case "block":
					event = newEvent(deps.Cfg.SessionID, notification.EventIteration, output.Message)
				case "allow":
					event = newEvent(deps.Cfg.SessionID, notification.EventSessionStop, "Session completed")
				}
				// Try to enrich with session data
				if parsed.SessionID != "" {
					sessionPath := filepath.Join(deps.Cfg.SessionsDir, parsed.SessionID+".local.md")
					if sess, err := session.LoadSession(sessionPath); err == nil {
						enrichEvent(&event, sess)
					}
				}
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

// generateSessionID creates a unique session ID.
func generateSessionID() string {
	return uuid.New().String()
}
