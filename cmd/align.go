package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennytrytek/ktrai/internal/detect"
	"github.com/kennytrytek/ktrai/internal/scaffold"
	"github.com/spf13/cobra"
)

var alignCmd = &cobra.Command{
	Use:   "align [directory]",
	Short: "Restructure a repository to the AI agent context layout",
	Long: `align moves files into the .agent/ directory layout without modifying
their contents. Missing files are created with default content; existing
files are only moved to their canonical location.

  - .agent/context/, .agent/rules/, and .agent/skills/ are created if absent.
  - AGENTS.md, CLAUDE.md, .cursor/rules/, and .claude/rules/ are migrated
    into .agent/ and replaced with symlinks when they exist as real files.
  - Skills from .claude/skills/ and .cursor/skills/ are migrated
    into .agent/skills/ and replaced with symlinks. Duplicate skill names
    across sources cause an error — resolve manually before re-running.
  - Configurations for unsupported AI tools are removed
    (e.g. .github/copilot-instructions.md).
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
	skillsDir := filepath.Join(agentDir, "skills")

	for _, d := range []string{contextDir, rulesDir, skillsDir} {
		if err := scaffold.EnsureDir(d); err != nil {
			return err
		}
	}
	fmt.Println("→ ensured .agent/context/, .agent/rules/, and .agent/skills/")

	// Migrate any real (non-symlink) files/dirs into .agent/ before symlinking.
	if err := migrateExisting(root, contextDir, rulesDir, skillsDir); err != nil {
		return err
	}

	// Remove configurations for unsupported AI tools.
	if err := removeUnsupportedConfigs(root); err != nil {
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
	if err := scaffold.WriteIfNotExists(
		filepath.Join(skillsDir, "update-agents-md", "SKILL.md"),
		scaffold.UpdateAgentsMdSkill,
	); err != nil {
		return err
	}
	if err := scaffold.WriteIfNotExists(
		filepath.Join(rulesDir, "update-ci-review.md"),
		scaffold.UpdateCiReviewRule,
	); err != nil {
		return err
	}
	if err := scaffold.WriteIfNotExists(
		filepath.Join(skillsDir, "ci-review", "SKILL.md"),
		scaffold.CiReviewSkill,
	); err != nil {
		return err
	}
	if err := scaffold.WriteIfNotExists(
		filepath.Join(skillsDir, "update-ci-review", "SKILL.md"),
		scaffold.UpdateCiReviewSkill,
	); err != nil {
		return err
	}
	lang := detect.Detect(root)
	filePatterns := lang.ReviewFilePatterns()
	excludedPatterns := lang.ReviewExcludedPatterns()
	githubWorkflowsDir := filepath.Join(root, ".github", "workflows")
	if err := scaffold.WriteIfNotExists(
		filepath.Join(githubWorkflowsDir, "ai-review.yml"),
		scaffold.AiReviewWorkflow(filePatterns, excludedPatterns),
	); err != nil {
		return err
	}
	if err := scaffold.WriteIfNotExists(
		filepath.Join(githubWorkflowsDir, "ai-review-scheduled.yml"),
		scaffold.AiReviewScheduledWorkflow(filePatterns, excludedPatterns),
	); err != nil {
		return err
	}
	fmt.Printf("→ detected language: %s\n", lang)
	fmt.Println("→ created any missing context, rule, skill, and workflow files")

	if err := wireToolSymlinks(root, agentDir, rulesDir, contextDir, skillsDir); err != nil {
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
func wireToolSymlinks(root, agentDir, rulesDir, contextDir, skillsDir string) error {
	_ = agentDir

	relRulesCursor, err := filepath.Rel(filepath.Join(root, ".cursor"), rulesDir)
	if err != nil {
		return err
	}
	relSkillsCursor, err := filepath.Rel(filepath.Join(root, ".cursor"), skillsDir)
	if err != nil {
		return err
	}
	relRulesClaude, err := filepath.Rel(filepath.Join(root, ".claude"), rulesDir)
	if err != nil {
		return err
	}
	relSkillsClaude, err := filepath.Rel(filepath.Join(root, ".claude"), skillsDir)
	if err != nil {
		return err
	}
	relContextFromRoot, err := filepath.Rel(root, filepath.Join(contextDir, "AGENTS.md"))
	if err != nil {
		return err
	}

	type link struct {
		linkPath string
		target   string
	}
	links := []link{
		{filepath.Join(root, ".cursor", "rules"), relRulesCursor},
		{filepath.Join(root, ".cursor", "skills"), relSkillsCursor},
		{filepath.Join(root, "AGENTS.md"), relContextFromRoot},
		{filepath.Join(root, ".claude", "rules"), relRulesClaude},
		{filepath.Join(root, ".claude", "skills"), relSkillsClaude},
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

// migrateExisting moves any real (non-symlink) context files, rule
// directories, and skill directories into the .agent/ layout and replaces
// them with symlinks.
func migrateExisting(root, contextDir, rulesDir, skillsDir string) error {
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

	// Migrate skills from tool-specific locations into .agent/skills/.
	// .claude/skills/ takes priority; .cursor/skills/ and .skills/ follow.
	// .skills/ is checked for migration only — it is never re-created afterward.
	// Duplicate skill names across sources cause an error.
	type skillSource struct {
		dir  string
		name string // human-readable label for error messages
	}
	sources := []skillSource{
		{filepath.Join(root, ".claude", "skills"), ".claude/skills/"},
		{filepath.Join(root, ".cursor", "skills"), ".cursor/skills/"},
		{filepath.Join(root, ".skills"), ".skills/"},
	}

	// Track which skill names are already in .agent/skills/ and which source
	// they came from so we can produce precise collision errors.
	claimedBy := map[string]string{}

	// Seed claimedBy with any skills already present in .agent/skills/.
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			claimedBy[e.Name()] = ".agent/skills/ (pre-existing)"
		}
	}

	var collisions []string
	for _, src := range sources {
		// Detect collisions before migrating.
		info, err := os.Lstat(src.dir)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			continue // does not exist or already a symlink — skip
		}
		entries, err := os.ReadDir(src.dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.Type()&os.ModeSymlink != 0 {
				continue // individual per-skill symlinks are left in place
			}
			if prior, exists := claimedBy[e.Name()]; exists {
				collisions = append(collisions,
					fmt.Sprintf("  skill %q from %s conflicts with %s", e.Name(), src.name, prior))
			}
		}
	}
	if len(collisions) > 0 {
		return fmt.Errorf(
			"skill name collisions detected — resolve before re-running:\n%s",
			strings.Join(collisions, "\n"),
		)
	}

	for _, src := range sources {
		migrated, err := scaffold.MigrateSkillsDir(src.dir, skillsDir)
		if err != nil {
			return fmt.Errorf("migrating skills from %s: %w", src.name, err)
		}
		for _, name := range migrated {
			claimedBy[name] = src.name
			fmt.Printf("  → moved skill %q from %s to .agent/skills/\n", name, src.name)
		}

		// Remove any per-skill symlinks left behind by MigrateSkillsDir and
		// then remove the now-empty directory so EnsureSymlink can replace the
		// whole thing with a single directory symlink.
		if err := scaffold.CollapseSkillSymlinks(src.dir, skillsDir); err != nil {
			return fmt.Errorf("collapsing skill symlinks in %s: %w", src.name, err)
		}
	}

	if len(claimedBy) > 0 {
		fmt.Println("→ migrated skills into .agent/skills/")
	}

	return nil
}

// unsupportedConfigs lists paths (relative to the repo root) of configuration
// files for AI tools that are not supported. Add future entries here.
var unsupportedConfigs = []string{
	".github/copilot-instructions.md",
}

// removeUnsupportedConfigs deletes configuration files for unsupported AI
// tools. Only regular files are removed; symlinks and directories are left
// untouched. Each removal is printed so the change is visible in the output.
func removeUnsupportedConfigs(root string) error {
	for _, rel := range unsupportedConfigs {
		path := filepath.Join(root, rel)
		info, err := os.Lstat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("lstat %s: %w", path, err)
		}
		// Only remove regular files; leave symlinks and directories to the user.
		if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing %s: %w", rel, err)
		}
		fmt.Printf("→ removed unsupported config: %s\n", rel)
	}
	return nil
}

func printAlignNextSteps(_ string) {
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .agent/context/AGENTS.md — fill in purpose and module roles.")
	fmt.Println("  2. Add skills to .agent/skills/ — each skill is a directory with a SKILL.md.")
	fmt.Println("  3. Review .github/workflows/ai-review.yml and ai-review-scheduled.yml —")
	fmt.Println("     update file-patterns and the reusable workflow version pin as needed.")
	fmt.Println("  4. Commit .agent/, .github/workflows/, and the symlinks (AGENTS.md, CLAUDE.md,")
	fmt.Println("     .cursor/rules, .cursor/skills, .claude/rules, .claude/skills).")
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
