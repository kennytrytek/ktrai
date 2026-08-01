package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ktrai",
	Short: "AI agent context scaffolding for software repositories",
	Long: `ktrai installs a tool-agnostic AI agent context layer into any repository.

It manages a .agent/ directory containing:
  - AGENTS.md        agent-optimised orientation document
  - symbols.md       ctags-driven symbol index (all function/class signatures)
  - rules/           always-apply and on-demand rules for Cursor and Claude

Subcommands:
  align  Move and restructure a repository into the AI agent context layout`,
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
