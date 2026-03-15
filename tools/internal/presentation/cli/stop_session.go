package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/domain/session"
)

// stopSessionCmd represents the stop-session command.
var stopSessionCmd = &cobra.Command{
	Use:   "stop-session",
	Short: "Stop the current session",
	Long:  `Stop the current active session.`,
	Run:   runStopSession,
}

var stopSessionReason string

func init() {
	stopSessionCmd.Flags().StringVarP(&stopSessionReason, "reason", "r", "completed", "reason for stopping")
	rootCmd.AddCommand(stopSessionCmd)
}

func runStopSession(cmd *cobra.Command, args []string) {
	sessionPath, err := getActiveSessionPath()
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
	fmt.Printf("  Reason: %s\n", stopSessionReason)
}
