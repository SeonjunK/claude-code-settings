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

For Go files: uses gofmt
For Python files: uses ruff format (if available) or black`,
		Run: func(cmd *cobra.Command, args []string) {
			deps.Log.Debug("post-tool-use: command started")

			// Try to read stdin for hook input
			stdinData, stdinErr := storage.ReadStdin()

			var filePath string

			if stdinErr == nil && len(stdinData) > 0 {
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

			files := args
			if filePath != "" {
				files = []string{filePath}
			}
			if len(files) == 0 {
				files = []string{"."}
			}

			deps.Log.Debug("post-tool-use: formatting", "files", files)

			goFiles := []string{}
			pyFiles := []string{}

			for _, f := range files {
				if endsWith(f, ".go") {
					goFiles = append(goFiles, f)
				} else if endsWith(f, ".py") {
					pyFiles = append(pyFiles, f)
				}
			}

			if filePath != "" && !endsWith(filePath, ".go") && !endsWith(filePath, ".py") {
				deps.Log.Debug("post-tool-use: not a Go/Python file, skipping", "file", filePath)
				os.Exit(0)
			}

			hasErrors := false

			if len(goFiles) > 0 {
				if err := formatGoFiles(goFiles, check, diff); err != nil {
					deps.Log.Error("post-tool-use: Go format error", "err", err, "files", goFiles)
					fmt.Fprintf(os.Stderr, "Go format error: %v\n", err)
					hasErrors = true
				}
			}

			if len(pyFiles) > 0 || (filePath == "" && len(args) == 0) {
				if err := formatPythonFiles(pyFiles, check, diff); err != nil {
					deps.Log.Error("post-tool-use: Python format error", "err", err, "files", pyFiles)
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

			deps.Log.Debug("post-tool-use: completed")
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
