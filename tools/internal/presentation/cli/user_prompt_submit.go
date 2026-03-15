package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/application/hook"
	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/infrastructure/storage"
)

// userPromptSubmitCmd represents the user-prompt-submit hook command.
var userPromptSubmitCmd = &cobra.Command{
	Use:   "user-prompt-submit",
	Short: "Handle user-prompt-submit hook",
	Long: `Handle user-prompt-submit hook when called from a Claude Code hook.

This command is triggered when a user submits a prompt.
It checks if the prompt is a /team-loops command and initializes the session.

Reads hook input from stdin and outputs JSON response.
Exit codes:
  0 - allow command
  1 - block command
  2 - error

Example hook configuration in .claude/settings.json:
  {
    "hooks": {
      "user-prompt-submit": ["claude-code-hooks user-prompt-submit"]
    }
  }`,
	Run: runUserPromptSubmit,
}

func init() {
	rootCmd.AddCommand(userPromptSubmitCmd)
}

func runUserPromptSubmit(cmd *cobra.Command, args []string) {
	// Read input from stdin
	input, err := storage.ReadStdin()
	if err != nil {
		fmt.Fprintln(os.Stderr, hook.ErrorOutput("failed to read stdin"))
		os.Exit(2)
	}

	// Parse input
	parsed, err := hook.ParseUserPromptSubmitInput(input)
	if err != nil {
		// If parsing fails, allow the command
		fmt.Println(hook.AllowOutput())
		os.Exit(0)
	}

	// Use session ID from config if not in input
	if parsed.SessionID == "" {
		parsed.SessionID = cfg.SessionID
	}

	// Handle using application layer
	handler := hook.NewUserPromptSubmitHandler(cfg.SessionsDir)
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

	// Exit with appropriate code
	switch output.Action {
	case "block":
		os.Exit(1)
	case "allow":
		os.Exit(0)
	default:
		os.Exit(2)
	}
}

// isTeamLoopsCommand checks if the prompt is a /team-loops command.
func isTeamLoopsCommand(prompt string) bool {
	return len(prompt) >= len("/team-loops") && prompt[:len("/team-loops")] == "/team-loops" &&
		(len(prompt) == len("/team-loops") || prompt[len("/team-loops")] == ' ')
}

// extractTeamLoopsGoal extracts the goal from a /team-loops prompt.
func extractTeamLoopsGoal(prompt string) string {
	if len(prompt) <= len("/team-loops") {
		return ""
	}
	goal := prompt[len("/team-loops"):]
	return strings.TrimSpace(goal)
}
