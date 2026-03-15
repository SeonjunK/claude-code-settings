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

Output (blocked):
  {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny"}, "systemMessage": "..."}
Output (allowed):
  (empty, exit 0)`,
		Run: func(cmd *cobra.Command, args []string) {
			deps.Log.Debug("pre-tool-use: command started", "tool_flag", tool)

			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil || len(stdinData) == 0 {
				deps.Log.Debug("pre-tool-use: no stdin, allowing")
				os.Exit(0)
			}

			var input struct {
				ToolName  string                 `json:"tool_name"`
				ToolInput map[string]interface{} `json:"tool_input"`
			}
			if err := json.Unmarshal(stdinData, &input); err != nil {
				deps.Log.Warn("pre-tool-use: invalid json stdin", "err", err)
				os.Exit(0)
			}

			deps.Log.Debug("pre-tool-use: parsed input", "tool_name", input.ToolName)

			// No vibe.json or no guard rules - allow all
			if deps.VibeConf == nil || !deps.VibeConf.HasGuard() {
				deps.Log.Debug("pre-tool-use: no guard config, allowing")
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
				deps.Log.Warn("pre-tool-use: DENIED", "tool", t, "reason", blockMsg)
				output := hook.DenyOutput(blockMsg)
				data, _ := json.Marshal(output)
				fmt.Println(string(data))

				event := newEvent(deps.Cfg.SessionID, notification.EventGuardDeny, blockMsg)
				event.Details = map[string]string{"tool": t}
				dispatchAndWait(deps.Notif, event)

				os.Exit(2)
			}

			deps.Log.Debug("pre-tool-use: allowed", "tool", t)
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
