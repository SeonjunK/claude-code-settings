package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/application/hook"
	"github.com/SeonjunK/claude-code-settings/tools/internal/infrastructure/storage"
)

// NewPostToolUseCmd creates the post-tool-use hook command.
func NewPostToolUseCmd(deps *Deps) *cobra.Command {
	var (
		check bool
		diff  bool
	)

	cmd := &cobra.Command{
		Use:   "post-tool-use [files...]",
		Short: "Handle post-tool-use hook for formatting",
		Long: `Handle post-tool-use hook by formatting files using appropriate formatters.

This command is designed to be called from Claude Code's post-tool-use hook.
It reads hook input from stdin and formats modified files.

For Go files: uses gofmt
For Python files: uses ruff format (if available) or black

Examples:
  vibe-format post-tool-use              # Format from stdin (hook mode)
  vibe-format post-tool-use file.go      # Format specific file`,
		Run: func(cmd *cobra.Command, args []string) {
			// Try to read stdin for hook input
			stdinData, stdinErr := storage.ReadStdin()

			var filePath string

			if stdinErr == nil && len(stdinData) > 0 {
				// Parse hook input from stdin
				var input struct {
					ToolName  string                 `json:"tool_name"`
					ToolInput map[string]interface{} `json:"tool_input"`
				}
				if err := json.Unmarshal(stdinData, &input); err == nil {
					if fp, ok := input.ToolInput["file_path"]; ok {
						if s, ok := fp.(string); ok {
							filePath = s
						}
					}
				}
			}

			// Fallback to args if no stdin file path
			files := args
			if filePath != "" {
				files = []string{filePath}
			}
			if len(files) == 0 {
				files = []string{"."}
			}

			// Group by file type
			goFiles := []string{}
			pyFiles := []string{}

			for _, f := range files {
				if endsWith(f, ".go") {
					goFiles = append(goFiles, f)
				} else if endsWith(f, ".py") {
					pyFiles = append(pyFiles, f)
				}
			}

			// If stdin specified a single non-Go/Python file - nothing to format
			if filePath != "" && !endsWith(filePath, ".go") && !endsWith(filePath, ".py") {
				os.Exit(0)
			}

			hasErrors := false

			// Format Go files
			if len(goFiles) > 0 {
				if err := formatGoFiles(goFiles, check, diff); err != nil {
					fmt.Fprintf(os.Stderr, "Go format error: %v\n", err)
					hasErrors = true
				}
			}

			// Format Python files
			if len(pyFiles) > 0 || (filePath == "" && len(args) == 0) {
				if err := formatPythonFiles(pyFiles, check, diff); err != nil {
					fmt.Fprintf(os.Stderr, "Python format error: %v\n", err)
					hasErrors = true
				}
			}

			if hasErrors {
				output := hook.SystemMessageOutput(fmt.Sprintf("⚠ Format failed for %v", files))
				data, _ := json.Marshal(output)
				fmt.Println(string(data))
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&check, "check", false, "check if files are formatted (exit 1 if not)")
	cmd.Flags().BoolVar(&diff, "diff", false, "display diff instead of modifying files")
	return cmd
}

func formatGoFiles(files []string, check, diff bool) error {
	if len(files) == 0 {
		files = []string{"."}
	}

	args := []string{"-w"}
	if check {
		args = []string{"-l"}
	}
	if diff {
		args = []string{"-d"}
	}
	args = append(args, files...)

	cmd := exec.Command("gofmt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func formatPythonFiles(files []string, check, diff bool) error {
	if len(files) == 0 {
		files = []string{"."}
	}

	// Try ruff first
	if _, err := exec.LookPath("ruff"); err == nil {
		args := []string{"format"}
		if check {
			args = append(args, "--check")
		}
		if diff {
			args = append(args, "--diff")
		}
		args = append(args, files...)

		cmd := exec.Command("ruff", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Fallback to black
	if _, err := exec.LookPath("black"); err == nil {
		args := []string{}
		if check {
			args = append(args, "--check")
		}
		if diff {
			args = append(args, "--diff")
		}
		args = append(args, files...)

		cmd := exec.Command("black", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println("No Python formatter found (ruff or black)")
	return nil
}

func endsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}
