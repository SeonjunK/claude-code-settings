package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/session"
)

// NewStopSessionCmd creates the stop-session command.
func NewStopSessionCmd(deps *Deps) *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "stop-session",
		Short: "Stop the current session",
		Long:  `Stop the current active session.`,
		Run: func(cmd *cobra.Command, args []string) {
			sessionPath, err := getActiveSessionPath(deps.Cfg.SessionsDir)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "No active session to stop: %v\n", err)
				return
			}

			sess, err := session.LoadSession(sessionPath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading session: %v\n", err)
				return
			}

			sess.Active = false

			if err := session.SaveSession(sessionPath, sess); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error stopping session: %v\n", err)
				return
			}

			fmt.Printf("Session stopped successfully!\n")
			fmt.Printf("  ID: %s\n", sess.SessionID)
			fmt.Printf("  Reason: %s\n", reason)
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "completed", "reason for stopping")
	return cmd
}
