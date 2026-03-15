package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// sessionStartCmd represents the session-start hook command.
var sessionStartCmd = &cobra.Command{
	Use:   "session-start",
	Short: "Handle session-start hook",
	Long: `Handle session-start hook when called from a Claude Code hook.

Validates the environment (jq, uv, guard.json) and outputs additionalContext.

Input stdin: {"session_id": "...", "source": "startup|resume|...", ...}

Output (always exit 0):
  {"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "jq=ok uv=ok guard.json=ok"}}`,
	Run: runSessionStart,
}

func init() {
	rootCmd.AddCommand(sessionStartCmd)
}

func runSessionStart(cmd *cobra.Command, args []string) {
	// Read stdin (ignore parse errors - always continue)
	stdinData, _ := io.ReadAll(os.Stdin)
	_ = stdinData // input parsing not needed for env validation

	projectDir := cfg.ProjectDir

	var checks []string

	// Check jq
	if _, err := exec.LookPath("jq"); err != nil {
		checks = append(checks, "jq=missing(warning)")
	} else {
		checks = append(checks, "jq=ok")
	}

	// Check uv
	if _, err := exec.LookPath("uv"); err != nil {
		checks = append(checks, "uv=missing(warning)")
	} else {
		checks = append(checks, "uv=ok")
	}

	// Check guard.json
	guardPaths := []string{
		filepath.Join(projectDir, ".claude", "guard.json"),
		filepath.Join(projectDir, "guard.json"),
	}
	guardFound := false
	for _, p := range guardPaths {
		if _, err := os.Stat(p); err == nil {
			guardFound = true
			break
		}
	}
	if guardFound {
		checks = append(checks, "guard.json=ok")
	} else {
		checks = append(checks, "guard.json=missing(warning)")
	}

	context := strings.Join(checks, " ")

	output := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":     "SessionStart",
			"additionalContext": context,
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal output: %v\n", err)
		os.Exit(0)
	}

	fmt.Println(string(data))
	os.Exit(0)
}
