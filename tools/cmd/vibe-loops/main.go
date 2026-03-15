package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/SeonjunK/claude-code-settings/tools/internal/presentation/cli"
)

func main() {
	deps := cli.InitDeps()
	root := &cobra.Command{Use: "vibe-loops", Short: "Team-loops session management, notifications, and bot"}
	root.AddCommand(cli.NewStopCmd(deps))
	root.AddCommand(cli.NewUserPromptSubmitCmd(deps))
	root.AddCommand(cli.NewStartCmd(deps))
	root.AddCommand(cli.NewStatusCmd(deps))
	root.AddCommand(cli.NewStopSessionCmd(deps))
	root.AddCommand(cli.NewNotifyCmd(deps))
	root.AddCommand(cli.NewTelegramBotCmd(deps))
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
