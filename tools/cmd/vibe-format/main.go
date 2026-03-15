package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/presentation/cli"
)

func main() {
	deps := cli.InitDeps()
	root := &cobra.Command{Use: "vibe-format", Short: "Auto-format hook for Go/Python files"}
	root.AddCommand(cli.NewPostToolUseCmd(deps))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
