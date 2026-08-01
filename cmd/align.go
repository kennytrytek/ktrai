package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kennytrytek/ktrai/internal/scaffold"
	"github.com/spf13/cobra"
)

var alignCmd = &cobra.Command{
	Use:   "align [directory]",
	Short: "Restructure a repository to the AI agent context layout",
	Long: `align moves files into the .agent/ directory layout without modifying
their contents. Missing files are created with default content; existing
files are only moved to their canonical location.

  - .agent/context/ and .agent/rules/ are created if absent.
  - AGENTS.md, CLAUDE.md, .cursor/rules/, and .claude/rules/ are migrated
    into .agent/ and replaced with symlinks when they exist as real files.
  - A default AGENTS.md is created under .agent/context/ if none exists.
  - The update-agents-md rule is added to .agent/rules/ if absent.
  - Cursor and Claude symlinks are always kept up to date.

If directory is omitted the current working directory is used.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAlign,
}

func init() {
	rootCmd.AddCommand(alignCmd)
}

func runAlign(_ *cobra.Command, args []string) error {
	root, err := resolveRoot(args)
	if err != nil {
		return err
	}

	agentDir := filepath.Join(root, ".agent")
	contextDir := filepath.Join(agentDir, "context")
	rulesDir := filepath.Join(agentDir, "rules")

	for _, d := range []string{contextDir, rulesDir} {
		if err := scaffold.EnsureDir(d); err != nil {
			return err
		}
	}
	fmt.Println("→ ensured .agent/context/ and .agent/rules/")

	// Migrate any real (non-symlink) files/dirs into .agent/ before symlinking.
	if err := migrateExisting(root, contextDir, rulesDir); err != nil {
		return err
	}

	// Create files only when absent — never overwrite existing content.
	if err := scaffold.WriteIfNotExists(
		filepath.Join(contextDir, "AGENTS.md"),
		defaultAgentsMD,
	); err != nil {
		return err
	}
	if err := scaffold.WriteIfNotExists(
		filepath.Join(rulesDir, "update-agents-md.md"),
		scaffold.UpdateAgentsMdRule,
	); err != nil {
		return err
	}
	fmt.Println("→ created any missing context and rule files")

	if err := wireToolSymlinks(root, agentDir, rulesDir, contextDir); err != nil {
		return err
	}

	fmt.Println("\n✓ Repository aligned.")
	printAlignNextSteps(root)
	return nil
}

// resolveRoot returns the absolute path for the target directory.
func resolveRoot(args []string) (string, error) {
	if len(args) == 1 {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return "", fmt.Errorf("resolving path: %w", err)
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return cwd, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// wireToolSymlinks creates the Cursor and Claude symlinks.
func wireToolSymlinks(root, agentDir, rulesDir, contextDir string) error {
	_ = agentDir

	relRules, err := filepath.Rel(filepath.Join(root, ".cursor"), rulesDir)
	if err != nil {
		return err
	}
	relContextFromRoot, err := filepath.Rel(root, filepath.Join(contextDir, "AGENTS.md"))
	if err != nil {
		return err
	}
	relRulesClaude, err := filepath.Rel(filepath.Join(root, ".claude"), rulesDir)
	if err != nil {
		return err
	}

	type link struct {
		linkPath string
		target   string
	}
	links := []link{
		{filepath.Join(root, ".cursor", "rules"), relRules},
		{filepath.Join(root, "AGENTS.md"), relContextFromRoot},
		{filepath.Join(root, ".claude", "rules"), relRulesClaude},
		{filepath.Join(root, "CLAUDE.md"), relContextFromRoot},
	}

	for _, l := range links {
		if err := scaffold.EnsureSymlink(l.linkPath, l.target); err != nil {
			return err
		}
	}
	fmt.Println("→ ensured Cursor and Claude symlinks")
	return nil
}

// migrateExisting moves any real (non-symlink) context files and rule
// directories into the .agent/ layout and replaces them with symlinks.
func migrateExisting(root, contextDir, rulesDir string) error {
	for _, f := range []string{
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "CLAUDE.md"),
	} {
		if err := scaffold.MoveToAgent(f, contextDir); err != nil {
			return fmt.Errorf("migrating %s: %w", f, err)
		}
	}
	for _, d := range []string{
		filepath.Join(root, ".cursor", "rules"),
		filepath.Join(root, ".claude", "rules"),
	} {
		if err := scaffold.MoveToAgent(d, rulesDir); err != nil {
			return fmt.Errorf("migrating %s: %w", d, err)
		}
	}
	return nil
}

func printAlignNextSteps(_ string) {
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .agent/context/AGENTS.md — fill in purpose and module roles.")
	fmt.Println("  2. Commit .agent/ and the symlinks (AGENTS.md, CLAUDE.md,")
	fmt.Println("     .cursor/rules, .claude/rules).")
}

const defaultAgentsMD = `# Project — agent context

TODO: add a one-sentence description of this project.

## Modules
| File | Role |
|---|---|

## Conventions
- TODO: add naming, style, and error-handling rules here.

## Commit pre-flight
Run in order before every commit:
1. ` + "`TODO: formatter command`" + ` — auto-fixes style
2. ` + "`TODO: linter command`" + ` — must pass with zero warnings
3. ` + "`TODO: test command`" + ` — preferred; mark as optional if slow

## Notes
<!-- reserved for human annotation — agents must not modify this section -->
`
