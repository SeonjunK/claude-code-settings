package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/application/hook"
)

// preToolUseCmd represents the pre-tool-use hook command.
var preToolUseCmd = &cobra.Command{
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
	Run: runPreToolUse,
}

var preToolUseTool string

func init() {
	preToolUseCmd.Flags().StringVar(&preToolUseTool, "tool", "", "tool type: Read, Write, or Bash")
	rootCmd.AddCommand(preToolUseCmd)
}

func runPreToolUse(cmd *cobra.Command, args []string) {
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

	projectDir := cfg.ProjectDir

	guardCfg, err := hook.LoadGuardConfig(projectDir)
	if err != nil {
		// guard.json not found - allow all
		os.Exit(0)
	}

	tool := preToolUseTool
	if tool == "" {
		tool = input.ToolName
	}

	var blockMsg string

	switch tool {
	case "Read":
		filePath := extractStringField(input.ToolInput, "file_path")
		if filePath != "" {
			blockMsg = hook.CheckFileAccess(guardCfg, "read", filePath)
		}
	case "Write", "Edit", "MultiEdit":
		filePath := extractStringField(input.ToolInput, "file_path")
		if filePath == "" {
			filePath = extractStringField(input.ToolInput, "path")
		}
		if filePath != "" {
			blockMsg = hook.CheckFileAccess(guardCfg, "write", filePath)
		}
	case "Bash":
		command := extractStringField(input.ToolInput, "command")
		if command != "" {
			blockMsg = hook.CheckBashCommand(guardCfg, command)
		}
	}

	if blockMsg != "" {
		output := hook.DenyOutput(blockMsg)
		data, _ := json.Marshal(output)
		fmt.Println(string(data))
		os.Exit(2) // non-zero to trigger deny
	}

	os.Exit(0)
}

func extractStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
