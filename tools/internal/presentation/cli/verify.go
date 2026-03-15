package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/application/hook"
	"github.com/neurumaru/blueprint-vibe/claude-plugin/internal/infrastructure/storage"
)

// verifyCmd represents the verify command.
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Run verification checks",
	Long: `Run verification checks for the project.

Runs linting, type checking, and tests to ensure code quality.

Examples:
  claude-code-hooks verify              # Run all verifications
  claude-code-hooks verify --lint       # Run linting only
  claude-code-hooks verify --test       # Run tests only
  claude-code-hooks verify --typecheck  # Run type checking only`,
	Run: runVerify,
}

var (
	verifyLint      bool
	verifyTest      bool
	verifyTypecheck bool
	verifyAll       bool
	verifyHookMode  bool
)

func init() {
	verifyCmd.Flags().BoolVar(&verifyLint, "lint", false, "run linting only")
	verifyCmd.Flags().BoolVar(&verifyTest, "test", false, "run tests only")
	verifyCmd.Flags().BoolVar(&verifyTypecheck, "typecheck", false, "run type checking only")
	verifyCmd.Flags().BoolVarP(&verifyAll, "all", "a", true, "run all verifications (default)")
	verifyCmd.Flags().BoolVar(&verifyHookMode, "hook", false, "run as Stop hook (reads stdin)")

	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) {
	// Detect hook mode: check if stdin has JSON or --hook flag
	stdinData, _ := storage.ReadStdin()
	isHookMode := verifyHookMode || (len(stdinData) > 0 && len(stdinData) < 10000 && stdinData[0] == '{')

	if isHookMode {
		runVerifyHook(stdinData)
		return
	}

	// Standalone mode (existing behavior)
	runStandaloneVerify()
}

func runVerifyHook(stdinData []byte) {
	projectDir := cfg.ProjectDir

	// Parse transcript path from stdin
	var input struct {
		TranscriptPath string `json:"transcript_path"`
		SessionID      string `json:"session_id"`
		Cwd            string `json:"cwd"`
	}
	if len(stdinData) > 0 {
		_ = json.Unmarshal(stdinData, &input)
	}

	if input.Cwd != "" && projectDir == "" {
		projectDir = input.Cwd
	}

	// Parse transcript to find changed files
	delta := parseTranscript(input.TranscriptPath, projectDir)

	if !delta.HasChanges {
		output := hook.SystemMessageOutput("✓ No code or hook changes to verify")
		data, _ := json.Marshal(output)
		fmt.Println(string(data))
		return
	}

	// Run verification pipeline
	decision, reason, message := runVerificationPipeline(delta, projectDir)

	var output map[string]interface{}
	if decision == "block" {
		output = hook.StopBlockOutput(reason, message)
	} else {
		output = hook.SystemMessageOutput(message)
	}

	data, _ := json.Marshal(output)
	fmt.Println(string(data))

	if decision == "block" {
		os.Exit(1)
	}
}

// transcriptDelta holds information about files changed in the transcript.
type transcriptDelta struct {
	Files       []string
	HasChanges  bool
	PythonFiles []string
	ShellFiles  []string
	JSONFiles   []string
}

func parseTranscript(transcriptPath, projectDir string) transcriptDelta {
	var delta transcriptDelta

	if transcriptPath == "" {
		return delta
	}

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return delta
	}

	// Parse transcript as JSON array
	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return delta
	}

	fileSet := make(map[string]bool)

	for _, entry := range entries {
		var msg struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content []struct {
					Type  string `json:"type"`
					Name  string `json:"name"`
					Input struct {
						FilePath string `json:"file_path"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal(entry, &msg); err != nil {
			continue
		}

		if msg.Type != "assistant" || msg.Message.Role != "assistant" {
			continue
		}

		for _, content := range msg.Message.Content {
			if content.Type != "tool_use" {
				continue
			}
			name := content.Name
			if name != "Write" && name != "Edit" && name != "MultiEdit" {
				continue
			}
			fp := content.Input.FilePath
			if fp == "" {
				continue
			}
			// Make relative
			if projectDir != "" && len(fp) > len(projectDir) && fp[:len(projectDir)] == projectDir {
				fp = fp[len(projectDir)+1:]
			}
			fileSet[fp] = true
		}
	}

	for fp := range fileSet {
		delta.Files = append(delta.Files, fp)
		if endsWith(fp, ".py") {
			delta.PythonFiles = append(delta.PythonFiles, fp)
		} else if endsWith(fp, ".sh") {
			delta.ShellFiles = append(delta.ShellFiles, fp)
		} else if endsWith(fp, ".json") {
			delta.JSONFiles = append(delta.JSONFiles, fp)
		}
	}

	delta.HasChanges = len(delta.Files) > 0
	return delta
}

func runVerificationPipeline(delta transcriptDelta, projectDir string) (decision, reason, message string) {
	// 1. Shell syntax check
	for _, fp := range delta.ShellFiles {
		target := fp
		if !endsWith(target, "/") && target[0] != '/' {
			target = projectDir + "/" + fp
		}
		if _, err := os.Stat(target); err != nil {
			continue
		}
		if err := exec.Command("bash", "-n", target).Run(); err != nil {
			return "block", "Shell syntax failed", fmt.Sprintf("⚠ Shell syntax failed for %s.", fp)
		}
	}

	// 2. JSON validation
	for _, fp := range delta.JSONFiles {
		target := fp
		if target[0] != '/' {
			target = projectDir + "/" + fp
		}
		if _, err := os.Stat(target); err != nil {
			continue
		}
		data, err := os.ReadFile(target)
		if err != nil {
			continue
		}
		if !json.Valid(data) {
			return "block", "JSON validation failed", fmt.Sprintf("⚠ JSON validation failed for %s.", fp)
		}
	}

	// 3. Python checks (only if python files changed)
	if len(delta.PythonFiles) == 0 {
		return "approve", "", "✓ Validation passed for changed shell/JSON files"
	}

	uvCmd := func(args ...string) error {
		c := exec.Command("uv", args...)
		c.Dir = projectDir
		c.Stdout = nil
		c.Stderr = nil
		return c.Run()
	}

	if err := uvCmd("run", "ruff", "format", "."); err != nil {
		return "block", "Format failed", "⚠ Format failed. Run `uv run ruff format .` to see details."
	}

	if err := uvCmd("run", "ruff", "check", ".", "--fix"); err != nil {
		return "block", "Lint failed", "⚠ Lint failed. Run `uv run ruff check . --fix` to see details."
	}

	if err := uvCmd("run", "mypy", "apps", "packages/python", "--config-file", "pyproject.toml"); err != nil {
		return "block", "Type check failed", "⚠ Type check failed. Run `uv run mypy apps packages/python --config-file pyproject.toml` to see details."
	}

	// pytest
	c := exec.Command("uv", "run", "pytest", "--no-cov", "-q")
	c.Dir = projectDir
	testOut, _ := c.CombinedOutput()
	if strings.Contains(string(testOut), "FAILED") {
		return "block", "Tests failed", "⚠ Tests failed. Run `uv run pytest` to see details."
	}

	return "approve", "", "✓ Validation passed (shell/json checks + Python checks)"
}

func runStandaloneVerify() {
	// Original standalone logic
	runAll := !verifyLint && !verifyTest && !verifyTypecheck

	failed := false

	if runAll || verifyLint {
		if err := runLint(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Lint failed: %v\n", err)
			failed = true
		} else {
			fmt.Println("✅ Lint passed")
		}
	}

	if runAll || verifyTypecheck {
		if err := runTypeCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Type check failed: %v\n", err)
			failed = true
		} else {
			fmt.Println("✅ Type check passed")
		}
	}

	if runAll || verifyTest {
		if err := runTests(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Tests failed: %v\n", err)
			failed = true
		} else {
			fmt.Println("✅ Tests passed")
		}
	}

	if failed {
		os.Exit(1)
	}

	fmt.Println("\n✨ All verifications passed!")
}

func runLint() error {
	// Go: go vet
	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "vet", "./...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Python: ruff check
	if _, err := exec.LookPath("ruff"); err == nil {
		cmd := exec.Command("ruff", "check", ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run() // Non-fatal for Python lint
	}

	return nil
}

func runTypeCheck() error {
	// Go: no separate type check needed (go vet covers most)
	// Python: mypy
	if _, err := exec.LookPath("mypy"); err == nil {
		cmd := exec.Command("mypy", "src")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run() // Non-fatal for mypy
	}

	return nil
}

func runTests() error {
	// Go tests
	if _, err := exec.LookPath("go"); err == nil {
		cmd := exec.Command("go", "test", "./...")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Python tests
	if _, err := exec.LookPath("uv"); err == nil {
		cmd := exec.Command("uv", "run", "pytest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run() // Non-fatal for Python tests
	}

	return nil
}
