package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/kennytrytek/ktrai/internal/scaffold"
	"github.com/spf13/cobra"
)

var restructureCmd = &cobra.Command{
	Use:   "restructure [directory]",
	Short: "Migrate an existing repository to the .agent/ layout",
	Long: `restructure detects existing AI context files (AGENTS.md, CLAUDE.md,
.cursor/rules/, .claude/rules/) that are not already symlinks into .agent/,
moves them into .agent/, creates symlinks at the original locations, and then
runs the same initialisation logic as 'ktrai init' for anything missing.

If directory is omitted the current working directory is used.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRestructure,
}

func init() {
	rootCmd.AddCommand(restructureCmd)
	restructureCmd.Flags().StringVarP(&flagLanguage, "language", "l", "",
		"Override language detection (go, python, typescript, java)")
}

func runRestructure(_ *cobra.Command, args []string) error {
	root, err := resolveRoot(args)
	if err != nil {
		return err
	}

	agentDir := filepath.Join(root, ".agent")
	contextDir := filepath.Join(agentDir, "context")
	rulesDir := filepath.Join(agentDir, "rules")

	fmt.Println("→ migrating existing context files into .agent/")

	// Files that belong in .agent/context/
	contextFiles := []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "CLAUDE.md"),
	}
	for _, f := range contextFiles {
		if err := scaffold.MoveToAgent(f, contextDir); err != nil {
			return fmt.Errorf("migrating %s: %w", f, err)
		}
	}

	// Directories that become .agent/rules/ (their contents are merged in).
	ruleDirs := []string{
		filepath.Join(root, ".cursor", "rules"),
		filepath.Join(root, ".claude", "rules"),
	}
	for _, d := range ruleDirs {
		if err := scaffold.MoveToAgent(d, rulesDir); err != nil {
			return fmt.Errorf("migrating %s: %w", d, err)
		}
	}

	fmt.Println("→ running init for any missing scaffold files")

	// Delegate to init for everything that is still missing.
	return runInit(nil, args)
}
