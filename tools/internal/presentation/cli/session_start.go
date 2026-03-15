package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// NewSessionStartCmd creates the session-start hook command.
func NewSessionStartCmd(deps *Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "session-start",
		Short: "Handle session-start hook",
		Long: `Handle session-start hook when called from a Claude Code hook.

Validates the environment (jq, uv, vibe.json) and outputs additionalContext.

Input stdin: {"session_id": "...", "source": "startup|resume|...", ...}

Output (always exit 0):
  {"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "jq=ok uv=ok vibe.json=ok"}}`,
		Run: func(cmd *cobra.Command, args []string) {
			deps.Log.Info("session-start: command started")

			// Read stdin (ignore parse errors - always continue)
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				deps.Log.Warn("session-start: stdin read error", "err", err)
			} else {
				deps.Log.Debug("session-start: stdin received", "size", len(stdinData))
			}

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

			// Check vibe.json
			if deps.VibeConf != nil {
				checks = append(checks, "vibe.json=ok")
			} else {
				checks = append(checks, "vibe.json=missing(warning)")
			}

			context := strings.Join(checks, " ")
			deps.Log.Info("session-start: env checks done", "result", context)

			output := map[string]any{
				"hookSpecificOutput": map[string]any{
					"hookEventName":     "SessionStart",
					"additionalContext": context,
				},
			}

			data, err := json.Marshal(output)
			if err != nil {
				deps.Log.Error("session-start: marshal failed", "err", err)
				fmt.Fprintf(os.Stderr, "failed to marshal output: %v\n", err)
				os.Exit(0)
			}

			fmt.Println(string(data))

			// Notify session start
			event := newEvent(deps.Cfg.SessionID, notification.EventSessionStart, "Session started: "+context)
			dispatchAndWait(deps.Notif, event)
			deps.Log.Info("session-start: completed")

			os.Exit(0)
		},
	}
}
