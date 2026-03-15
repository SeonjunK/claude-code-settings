package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/hook"
	"github.com/SeonjunK/claude-code-settings/tools/internal/domain/notification"
)

// NewPreToolUseCmd creates the pre-tool-use hook command.
func NewPreToolUseCmd(deps *Deps) *cobra.Command {
	var tool string

	cmd := &cobra.Command{
		Use:   "pre-tool-use",
		Short: "Handle pre-tool-use hook for guard checks",
		Long: `Handle pre-tool-use hook by checking guard rules before tool execution.

Reads hook input from stdin and outputs deny decision if blocked.

Input stdin:
  {"session_id": "...", "tool_name": "Read", "tool_input": {"file_path": "..."}}
  {"session_id": "...", "tool_name": "Bash", "tool_input": {"command": "rm -rf ..."}}

Flags:
  --tool Read   Check file read access
  --tool Write  Check file write access
  --tool Bash   Check bash command

Output (blocked):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny"}, "systemMessage": "..."}
Output (allowed):
  (empty, exit 0)`,
		Run: func(cmd *cobra.Command, args []string) {
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil || len(stdinData) == 0 {
				os.Exit(0)
			}

			var input struct {
				ToolName  string                 `json:"tool_name"`
				ToolInput map[string]interface{} `json:"tool_input"`
			}
			if err := json.Unmarshal(stdinData, &input); err != nil {
				os.Exit(0)
			}

			// No vibe.json or no guard rules - allow all
			if deps.VibeConf == nil || !deps.VibeConf.HasGuard() {
				os.Exit(0)
			}

			guard := &deps.VibeConf.Guard

			t := tool
			if t == "" {
				t = input.ToolName
			}

			var blockMsg string

			switch t {
			case "Read":
				filePath := extractStringField(input.ToolInput, "file_path")
				if filePath != "" {
					blockMsg = hook.CheckFileAccess(guard, "read", filePath)
				}
			case "Write", "Edit", "MultiEdit":
				filePath := extractStringField(input.ToolInput, "file_path")
				if filePath == "" {
					filePath = extractStringField(input.ToolInput, "path")
				}
				if filePath != "" {
					blockMsg = hook.CheckFileAccess(guard, "write", filePath)
				}
			case "Bash":
				command := extractStringField(input.ToolInput, "command")
				if command != "" {
					blockMsg = hook.CheckBashCommand(guard, command)
				}
			}

			if blockMsg != "" {
				output := hook.DenyOutput(blockMsg)
				data, _ := json.Marshal(output)
				fmt.Println(string(data))

				// Notify guard deny
				event := newEvent(deps.Cfg.SessionID, notification.EventGuardDeny, blockMsg)
				event.Details = map[string]string{"tool": t}
				dispatchAndWait(deps.Notif, event)

				os.Exit(2) // non-zero to trigger deny
			}

			os.Exit(0)
		},
	}

	cmd.Flags().StringVar(&tool, "tool", "", "tool type: Read, Write, or Bash")
	return cmd
}

func extractStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
