package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/presentation/cli"
)

func main() {
	deps := cli.InitDeps()
	defer deps.Close()
	root := &cobra.Command{Use: "vibe-guard", Short: "Guard hook for file/command access control"}
	root.AddCommand(cli.NewPreToolUseCmd(deps))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
