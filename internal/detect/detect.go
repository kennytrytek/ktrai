// Package detect infers the primary programming language of a repository
// from marker files at its root.
package detect

import (
	"os"
	"path/filepath"
)

// Language represents a supported project language.
type Language string

const (
	Go         Language = "go"
	Python     Language = "python"
	TypeScript Language = "typescript"
	Java       Language = "java"
	Unknown    Language = "unknown"
)

// String returns the display name for the language.
func (l Language) String() string {
	switch l {
	case Go:
		return "Go"
	case Python:
		return "Python"
	case TypeScript:
		return "TypeScript"
	case Java:
		return "Java/Kotlin"
	default:
		return "Unknown"
	}
}

// CtagsLanguages returns the value to pass to ctags --languages= for this language.
func (l Language) CtagsLanguages() string {
	switch l {
	case Go:
		return "Go"
	case Python:
		return "Python"
	case TypeScript:
		return "TypeScript,JavaScript"
	case Java:
		return "Java,Kotlin"
	default:
		return ""
	}
}

// markers maps marker filenames (checked at repo root) to their language.
// Checked in order; first match wins.
var markers = []struct {
	file string
	lang Language
}{
	{"go.mod", Go},
	{"pyproject.toml", Python},
	{"requirements.txt", Python},
	{"setup.py", Python},
	{"setup.cfg", Python},
	{"package.json", TypeScript},
	{"build.gradle.kts", Java},
	{"build.gradle", Java},
	{"pom.xml", Java},
}

// Detect returns the primary language detected at root.
// Returns Unknown if no marker file is found.
func Detect(root string) Language {
	for _, m := range markers {
		if fileExists(filepath.Join(root, m.file)) {
			return m.lang
		}
	}
	return Unknown
}

// ReviewFilePatterns returns newline-separated glob patterns for files that
// should be included in an AI code review workflow for this language.
func (l Language) ReviewFilePatterns() string {
	switch l {
	case Go:
		return "*.go\n*.yaml\n*.yml\n*.md\n*.sh"
	case Python:
		return "*.py\n*.yaml\n*.yml\n*.md\n*.sh\n*.toml\n*.cfg\n*.ini"
	case TypeScript:
		return "*.ts\n*.tsx\n*.js\n*.jsx\n*.yaml\n*.yml\n*.md\n*.sh\n*.json"
	case Java:
		return "*.java\n*.kt\n*.yaml\n*.yml\n*.md\n*.sh\n*.xml\n*.gradle\n*.gradle.kts"
	default:
		return "*.go\n*.py\n*.ts\n*.tsx\n*.js\n*.java\n*.kt\n*.yaml\n*.yml\n*.md\n*.sh"
	}
}

// ReviewExcludedPatterns returns newline-separated glob patterns for files
// that should be excluded from an AI code review workflow for this language.
func (l Language) ReviewExcludedPatterns() string {
	switch l {
	case Go:
		return "go.sum\n**/go.sum\n**/vendor/**\n**/testdata/**"
	case Python:
		return "**/__pycache__/**\n**/.venv/**\n**/venv/**\n**/*.pyc\n**/dist/**\n**/build/**"
	case TypeScript:
		return "**/node_modules/**\n**/dist/**\n**/build/**\n**/*.min.js\n**/*.generated.*\npackage-lock.json\nyarn.lock"
	case Java:
		return "**/build/**\n**/target/**\n**/*.class"
	default:
		return "**/vendor/**\n**/node_modules/**\n**/__pycache__/**\n**/dist/**\n**/build/**\n**/target/**"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
