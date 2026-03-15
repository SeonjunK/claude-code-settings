package cli

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// NewNotifyCmd creates the notify command.
func NewNotifyCmd(deps *Deps) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "notify <event_type>",
		Short: "Send a notification event manually",
		Long: `Send a notification event to configured providers.

Event types: session_start, session_stop, iteration,
             verify_pass, verify_fail, task_complete,
             task_blocked, guard_deny

Examples:
  vibe-loops notify verify_pass -m "All tests passed"
  vibe-loops notify session_start -m "New session"`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if deps.Notif == nil {
				fmt.Fprintln(os.Stderr, "[notify] No notification providers configured. Enable providers in vibe.json.")
				os.Exit(0)
			}

			eventType := notification.EventType(args[0])

			// Validate event type
			if !slices.Contains(notification.AllEventTypes(), eventType) {
				fmt.Fprintf(os.Stderr, "Unknown event type: %s\n", args[0])
				fmt.Fprintln(os.Stderr, "Valid types: session_start, session_stop, iteration, verify_pass, verify_fail, task_complete, task_blocked, guard_deny")
				os.Exit(1)
			}

			summary := message
			if summary == "" {
				summary = eventType.Label()
			}

			event := newEvent(deps.Cfg.SessionID, eventType, summary)

			// Try to enrich from active session
			if sessionPath, err := getActiveSessionPath(deps.Cfg.SessionsDir); err == nil {
				if sess, err := loadSessionFromPath(sessionPath); err == nil {
					enrichEvent(&event, sess)
				}
			}

			deps.Notif.DispatchSync(event)
			fmt.Printf("Notification sent: %s %s\n", eventType.Emoji(), summary)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "notification message")
	return cmd
}
