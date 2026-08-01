package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ktrai",
	Short: "AI scaffolding for software repositories",
	Long: `ktrai is AI scaffolding for software repositories, as desired by KTry.

It manages a .agent/ directory containing:
  - AGENTS.md   agent-optimised orientation document
  - rules/      always-apply and on-demand rules for Cursor and Claude

Subcommands:
  align  Set up or migrate a repository into the .agent/ layout`,
	Version: version,
}

// version is set at build time via -ldflags.
var version = "dev"

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
