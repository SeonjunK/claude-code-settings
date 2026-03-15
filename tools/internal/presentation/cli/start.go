package cli

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/domain/session"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start <goal>",
	Short: "Start a new team-loops session",
	Long: `Start a new team-loops session with the specified goal.

Example:
  claude-code-hooks start "인증 경계 강화" --max-iterations 10 --max-parallel 3`,
	Args: cobra.ExactArgs(1),
	Run:  runStart,
}

var (
	startMaxIterations int
	startMaxParallel   int
)

func init() {
	startCmd.Flags().IntVarP(&startMaxIterations, "max-iterations", "i", session.DefaultMaxIterations, "maximum number of iterations")
	startCmd.Flags().IntVarP(&startMaxParallel, "max-parallel", "p", session.DefaultMaxParallelTeammates, "maximum parallel agents")

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) {
	goal := args[0]
	sessionID := uuid.New().String()

	opts := []session.Option{
		session.WithMaxIterations(startMaxIterations),
		session.WithMaxParallelTeammates(startMaxParallel),
	}

	sess := session.CreateSession(goal, sessionID, opts...)

	sessionPath := cfg.GetSessionPath(sessionID)
	if err := session.SaveSession(sessionPath, sess); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error creating session: %v\n", err)
		return
	}

	fmt.Printf("Session started successfully!\n")
	fmt.Printf("  ID: %s\n", sess.SessionID)
	fmt.Printf("  Goal: %s\n", sess.Goal.Description)
	fmt.Printf("  Max Iterations: %d\n", sess.MaxIterations)
	fmt.Printf("  Max Parallel: %d\n", sess.Control.MaxParallelTeammates)
	fmt.Printf("  Session File: %s\n", sessionPath)
}
