// Package agentsmd reads an existing AGENTS.md (or drafts one from a module
// scan) and rewrites it following the update-agents-md rule conventions:
// purpose, module table, conventions, do-not-edit list — nothing else.
package agentsmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kennytrytek/ktrai/internal/detect"
)

// Content holds the parsed sections of an AGENTS.md.
type Content struct {
	// ProjectName is derived from the repo root directory name.
	ProjectName string
	// Purpose is the one-sentence description of what the project does.
	Purpose string
	// Modules is the list of (file, role) pairs for the module table.
	Modules []Module
	// Conventions are the bullet points an agent must follow when writing code.
	Conventions []string
	// DoNotEdit lists paths that must not be overwritten.
	DoNotEdit []string
}

// Module represents a row in the module table.
type Module struct {
	File string
	Role string
}

// Load reads an existing AGENTS.md at path and returns a Content struct.
// It preserves the purpose, modules table, conventions, and do-not-edit list
// while stripping everything else (function names, diagrams, etc.).
func Load(path string) (*Content, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parse(string(data), filepath.Base(filepath.Dir(path))), nil
}

// Draft builds a Content from a module scan of root when no AGENTS.md exists.
func Draft(root string, lang detect.Language) *Content {
	projectName := filepath.Base(root)
	modules := scanModules(root, lang)
	return &Content{
		ProjectName: projectName,
		Purpose:     fmt.Sprintf("%s — TODO: add a one-sentence description of this project.", projectName),
		Modules:     modules,
		Conventions: []string{
			"TODO: add project-specific coding conventions here.",
		},
		DoNotEdit: []string{
			".agent/context/symbols.md",
		},
	}
}

// Render serialises Content back to Markdown following the update-agents-md rule.
func Render(c *Content) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s — agent context\n\n", c.ProjectName))
	sb.WriteString(c.Purpose + "\n\n")
	sb.WriteString("Symbol index (all signatures): `.agent/context/symbols.md`\n\n")

	if len(c.Modules) > 0 {
		sb.WriteString("## Modules\n")
		sb.WriteString("| File | Role |\n")
		sb.WriteString("|---|---|\n")
		for _, m := range c.Modules {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", m.File, m.Role))
		}
		sb.WriteString("\n")
	}

	if len(c.Conventions) > 0 {
		sb.WriteString("## Conventions\n")
		for _, cv := range c.Conventions {
			sb.WriteString(fmt.Sprintf("- %s\n", cv))
		}
		sb.WriteString("\n")
	}

	if len(c.DoNotEdit) > 0 {
		sb.WriteString("## Do not edit\n")
		parts := make([]string, len(c.DoNotEdit))
		for i, p := range c.DoNotEdit {
			parts[i] = fmt.Sprintf("`%s`", p)
		}
		sb.WriteString(strings.Join(parts, ", ") + "\n")
	}

	return sb.String()
}

// parse extracts structured sections from raw Markdown.
// It is intentionally permissive — anything not recognised is dropped.
func parse(raw, defaultProjectName string) *Content {
	c := &Content{}
	scanner := bufio.NewScanner(strings.NewReader(raw))

	type section int
	const (
		secNone section = iota
		secModules
		secConventions
		secDoNotEdit
	)

	cur := secNone
	inTable := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// H1 → project name + purpose (rest of the line after "# NAME — ")
		if strings.HasPrefix(trimmed, "# ") {
			rest := strings.TrimPrefix(trimmed, "# ")
			if idx := strings.Index(rest, " — "); idx != -1 {
				c.ProjectName = strings.TrimSpace(rest[:idx])
			} else {
				c.ProjectName = strings.TrimSpace(rest)
			}
			cur = secNone
			inTable = false
			continue
		}

		// Purpose line: the first non-empty, non-heading line after H1
		// (before the first H2).
		if c.Purpose == "" && !strings.HasPrefix(trimmed, "#") && trimmed != "" &&
			!strings.HasPrefix(trimmed, "Symbol index") {
			c.Purpose = trimmed
			continue
		}

		// H2 section headers
		switch {
		case strings.HasPrefix(trimmed, "## Modules"):
			cur = secModules
			inTable = false
			continue
		case strings.HasPrefix(trimmed, "## Conventions"):
			cur = secConventions
			inTable = false
			continue
		case strings.HasPrefix(trimmed, "## Do not edit"):
			cur = secDoNotEdit
			inTable = false
			continue
		case strings.HasPrefix(trimmed, "## "):
			cur = secNone
			inTable = false
			continue
		}

		switch cur {
		case secModules:
			if strings.HasPrefix(trimmed, "|") {
				inTable = true
				// skip header and separator rows
				if strings.Contains(trimmed, "---|") || strings.Contains(trimmed, "File") {
					continue
				}
				parts := strings.Split(trimmed, "|")
				if len(parts) >= 3 {
					file := strings.Trim(strings.TrimSpace(parts[1]), "`")
					role := strings.TrimSpace(parts[2])
					if file != "" && role != "" {
						c.Modules = append(c.Modules, Module{File: file, Role: role})
					}
				}
			} else if inTable && trimmed == "" {
				cur = secNone
			}

		case secConventions:
			if strings.HasPrefix(trimmed, "- ") {
				c.Conventions = append(c.Conventions, strings.TrimPrefix(trimmed, "- "))
			}

		case secDoNotEdit:
			if trimmed != "" {
				// Can be a comma-separated list of `path` entries or a single bullet.
				entries := strings.Split(trimmed, ",")
				for _, e := range entries {
					e = strings.Trim(strings.TrimSpace(e), "`")
					e = strings.TrimPrefix(e, "- ")
					e = strings.Trim(e, "`")
					if e != "" {
						c.DoNotEdit = append(c.DoNotEdit, e)
					}
				}
			}
		}
	}

	if c.ProjectName == "" {
		c.ProjectName = defaultProjectName
	}
	if c.Purpose == "" {
		c.Purpose = fmt.Sprintf("%s — TODO: add a one-sentence description.", c.ProjectName)
	}

	return c
}

// scanModules walks root looking for source files to populate the module table.
func scanModules(root string, lang detect.Language) []Module {
	var modules []Module
	seen := map[string]bool{}

	var exts []string
	switch lang {
	case detect.Go:
		exts = []string{".go"}
	case detect.Python:
		exts = []string{".py"}
	case detect.TypeScript:
		exts = []string{".ts", ".tsx", ".js", ".jsx"}
	case detect.Java:
		exts = []string{".java", ".kt"}
	default:
		exts = []string{".go", ".py", ".ts", ".java", ".kt"}
	}

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		name := info.Name()
		// Skip hidden dirs, vendor, node_modules, test dirs.
		if info.IsDir() && (strings.HasPrefix(name, ".") ||
			name == "vendor" || name == "node_modules" ||
			name == "testdata" || name == "__pycache__") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(name)
		for _, e := range exts {
			if ext == e {
				rel, _ := filepath.Rel(root, path)
				if !seen[rel] && !isTestFile(rel, lang) {
					seen[rel] = true
					modules = append(modules, Module{
						File: rel,
						Role: "TODO: describe the role of this module.",
					})
				}
				break
			}
		}
		return nil
	})

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].File < modules[j].File
	})
	return modules
}

func isTestFile(rel string, lang detect.Language) bool {
	base := filepath.Base(rel)
	switch lang {
	case detect.Go:
		return strings.HasSuffix(base, "_test.go")
	case detect.Python:
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	case detect.TypeScript:
		return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.")
	case detect.Java:
		return strings.HasSuffix(base, "Test.java") || strings.HasSuffix(base, "Test.kt")
	}
	return false
}
