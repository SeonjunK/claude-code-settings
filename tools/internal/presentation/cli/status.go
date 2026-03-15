package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/domain/session"
)

// statusCmd represents the status command.
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current session status",
	Long:  `Display the status of the current active session.`,
	Run:   runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	sessionPath, err := getActiveSessionPath()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "No active session: %v\n", err)
		return
	}

	sess, err := session.LoadSession(sessionPath)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error loading session: %v\n", err)
		return
	}

	fmt.Printf("Session Status\n")
	fmt.Printf("==============\n\n")
	fmt.Printf("ID: %s\n", sess.SessionID)
	fmt.Printf("Team Name: %s\n", sess.TeamName)
	fmt.Printf("Goal: %s\n", sess.Goal.Description)
	fmt.Printf("Active: %t\n", sess.Active)
	fmt.Printf("Iteration: %d/%d\n", sess.Iteration, sess.MaxIterations)
	fmt.Printf("Max Parallel: %d\n", sess.Control.MaxParallelTeammates)
	fmt.Printf("Started: %s\n", sess.StartedAt.Format("2006-01-02 15:04:05"))

	if len(sess.Teammates) > 0 {
		fmt.Printf("\nTeammates (%d):\n", len(sess.Teammates))
		for _, t := range sess.Teammates {
			fmt.Printf("  - %s (%s): %s\n", t.Name, t.SubagentType, t.Status)
		}
	}

	fmt.Printf("\nMetrics\n")
	fmt.Printf("-------\n")
	fmt.Printf("Tasks Completed: %d\n", sess.Metrics.TasksCompleted)
	fmt.Printf("Tasks Pending: %d\n", sess.Metrics.TasksPending)
	fmt.Printf("Tasks In Progress: %d\n", sess.Metrics.TasksInProgress)
	fmt.Printf("Total Tool Calls: %d\n", sess.Metrics.TotalToolCalls)
	if sess.Metrics.LastActivityAt != nil {
		fmt.Printf("Last Activity: %s\n", sess.Metrics.LastActivityAt.Format("2006-01-02 15:04:05"))
	}
}

// getActiveSessionPath finds the first active session file.
func getActiveSessionPath() (string, error) {
	entries, err := os.ReadDir(cfg.SessionsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read sessions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".local.md") {
			path := filepath.Join(cfg.SessionsDir, entry.Name())
			sess, err := session.LoadSession(path)
			if err != nil {
				continue
			}
			if sess.Active {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no active session found")
}
