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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
