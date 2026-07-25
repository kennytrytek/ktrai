package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kennytrytek/ktrai/internal/agentsmd"
	"github.com/kennytrytek/ktrai/internal/detect"
	"github.com/kennytrytek/ktrai/internal/makefile"
	"github.com/kennytrytek/ktrai/internal/scaffold"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialise a repository with AI agent context scaffolding",
	Long: `init creates the .agent/ directory layout, writes AGENTS.md and
symbols.md, installs Cursor and Claude rule symlinks, and (optionally)
injects gen-context / prep targets into an existing Makefile.

If directory is omitted the current working directory is used.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var flagLanguage string

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&flagLanguage, "language", "l", "",
		"Override language detection (go, python, typescript, java)")
}

func runInit(_ *cobra.Command, args []string) error {
	root, err := resolveRoot(args)
	if err != nil {
		return err
	}

	// 1. Detect language.
	lang := resolveLanguage(root, flagLanguage)
	fmt.Printf("→ detected language: %s\n", lang)

	// 2. Load or draft AGENTS.md content.
	var content *agentsmd.Content
	agentsPath := filepath.Join(root, "AGENTS.md")
	if fileExists(agentsPath) {
		fmt.Println("→ found existing AGENTS.md — using as source material")
		content, err = agentsmd.Load(agentsPath)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("→ no AGENTS.md found — drafting from module scan")
		content = agentsmd.Draft(root, lang)
	}

	// 3. Create .agent/ directories.
	agentDir := filepath.Join(root, ".agent")
	contextDir := filepath.Join(agentDir, "context")
	rulesDir := filepath.Join(agentDir, "rules")
	for _, d := range []string{contextDir, rulesDir} {
		if err := scaffold.EnsureDir(d); err != nil {
			return err
		}
	}
	fmt.Println("→ created .agent/context/ and .agent/rules/")

	// 4. Write .agent/context/AGENTS.md.
	rendered := agentsmd.Render(content)
	if err := scaffold.WriteFile(filepath.Join(contextDir, "AGENTS.md"), rendered); err != nil {
		return err
	}
	fmt.Println("→ wrote .agent/context/AGENTS.md")

	// 5. Write rule files.
	if err := scaffold.WriteFile(
		filepath.Join(rulesDir, "codebase-map.md"),
		scaffold.CodemapRule,
	); err != nil {
		return err
	}
	if err := scaffold.WriteFile(
		filepath.Join(rulesDir, "update-agents-md.md"),
		scaffold.UpdateAgentsMdRule,
	); err != nil {
		return err
	}
	fmt.Println("→ wrote .agent/rules/codebase-map.md and update-agents-md.md")

	// 6. Generate symbols.md (or write placeholder).
	symbolsPath := filepath.Join(contextDir, "symbols.md")
	if err := generateSymbols(root, lang, symbolsPath); err != nil {
		return err
	}

	// 7. Wire up Cursor and Claude symlinks.
	if err := wireToolSymlinks(root, agentDir, rulesDir, contextDir); err != nil {
		return err
	}

	// 8. Inject Makefile targets.
	makefilePath := filepath.Join(root, "Makefile")
	if err := makefile.Inject(makefilePath, lang); err != nil {
		return err
	}
	if fileExists(makefilePath) {
		fmt.Println("→ updated Makefile with gen-context and prep targets")
	}

	fmt.Println("\n✓ AI scaffolding complete.")
	printNextSteps(root)
	return nil
}

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

func resolveLanguage(root, override string) detect.Language {
	if override != "" {
		switch strings.ToLower(override) {
		case "go":
			return detect.Go
		case "python", "py":
			return detect.Python
		case "typescript", "ts":
			return detect.TypeScript
		case "java", "kotlin", "java/kotlin":
			return detect.Java
		}
	}
	return detect.Detect(root)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// generateSymbols runs ctags piped through gen-symbols if universal-ctags
// with JSON support is available; otherwise writes a placeholder.
func generateSymbols(root string, lang detect.Language, outPath string) error {
	ktrai, err := os.Executable()
	if err != nil {
		ktrai = "ktrai"
	}

	ctagsBin, ctagsOK := findUniversalCtags()
	if !ctagsOK {
		fmt.Println("→ universal-ctags not found — writing symbols placeholder")
		return scaffold.WriteFile(outPath, scaffold.SymbolsPlaceholder)
	}

	langFlag := lang.CtagsLanguages()
	ctagsArgs := []string{"--output-format=json", "--fields=+KSZnte", "-R", "."}
	if langFlag != "" {
		ctagsArgs = append([]string{fmt.Sprintf("--languages=%s", langFlag)}, ctagsArgs...)
	}

	ctagsCmd := exec.Command(ctagsBin, ctagsArgs...)
	ctagsCmd.Dir = root
	ctagsOut, err := ctagsCmd.Output()
	if err != nil {
		fmt.Printf("→ ctags error (%v) — writing symbols placeholder\n", err)
		return scaffold.WriteFile(outPath, scaffold.SymbolsPlaceholder)
	}

	ktraiCmd := exec.Command(ktrai, "gen-symbols")
	ktraiCmd.Stdin = strings.NewReader(string(ctagsOut))
	out, err := ktraiCmd.Output()
	if err != nil {
		fmt.Printf("→ gen-symbols error (%v) — writing symbols placeholder\n", err)
		return scaffold.WriteFile(outPath, scaffold.SymbolsPlaceholder)
	}

	if err := scaffold.WriteFile(outPath, string(out)); err != nil {
		return err
	}
	fmt.Println("→ generated .agent/context/symbols.md")
	return nil
}

// findUniversalCtags returns the path to universal-ctags if it supports JSON output.
func findUniversalCtags() (string, bool) {
	for _, bin := range []string{"ctags", "/usr/local/bin/ctags", "/opt/homebrew/bin/ctags"} {
		path, err := exec.LookPath(bin)
		if err != nil {
			continue
		}
		out, err := exec.Command(path, "--list-output-formats").Output()
		if err != nil {
			continue
		}
		if strings.Contains(string(out), "json") {
			return path, true
		}
	}
	return "", false
}

// wireToolSymlinks creates the Cursor and Claude symlinks.
func wireToolSymlinks(root, agentDir, rulesDir, contextDir string) error {
	_ = agentDir // kept for clarity; symlink targets are relative

	// Compute relative paths from each symlink location to the real directory.
	// .cursor/rules -> ../../.agent/rules  (relative from .cursor/)
	// AGENTS.md -> .agent/context/AGENTS.md  (relative from root)
	// .claude/rules -> ../../.agent/rules
	// CLAUDE.md -> .agent/context/AGENTS.md

	type link struct {
		linkPath string
		target   string
	}

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
	fmt.Println("→ created Cursor and Claude symlinks")
	return nil
}

func printNextSteps(root string) {
	makefilePath := filepath.Join(root, "Makefile")
	hasMakefile := fileExists(makefilePath)

	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .agent/context/AGENTS.md — fill in purpose and module roles.")
	if !hasMakefile {
		fmt.Println("  2. Add a gen-context target to your build tool:")
		fmt.Println("       ctags --output-format=json --fields=+KSZnte -R . \\")
		fmt.Println("         | ktrai gen-symbols > .agent/context/symbols.md")
	} else {
		fmt.Println("  2. Run `make gen-context` to (re)generate the symbol index.")
	}
	fmt.Println("  3. Commit .agent/ and the symlinks (AGENTS.md, CLAUDE.md,")
	fmt.Println("     .cursor/rules, .claude/rules).")
}
