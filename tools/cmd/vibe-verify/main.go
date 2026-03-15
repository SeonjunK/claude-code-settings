package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/presentation/cli"
)

func main() {
	deps := cli.InitDeps()
	root := &cobra.Command{Use: "vibe-verify", Short: "Verification pipeline and environment checks"}
	root.AddCommand(cli.NewSessionStartCmd(deps))
	root.AddCommand(cli.NewVerifyCmd(deps))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
