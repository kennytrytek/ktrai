package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ktrai",
	Short: "AI agent context scaffolding for software repositories",
	Long: `ktrai installs a tool-agnostic AI agent context layer into any
Go, Python, TypeScript, or Java/Kotlin repository.

It creates a .agent/ directory containing:
  - AGENTS.md        agent-optimised orientation document
  - symbols.md       ctags-driven symbol index (all function/class signatures)
  - rules/           always-apply and on-demand rules for Cursor and Claude

Subcommands:
  init         Initialise a repository with AI scaffolding
  restructure  Migrate an existing repository to the .agent/ layout
  gen-symbols  Convert ctags JSON (stdin) to a Markdown symbol index (stdout)`,
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
