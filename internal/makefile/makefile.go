// Package makefile injects gen-context and prep targets into an existing
// Makefile. All operations are idempotent — targets already present are left
// unchanged.
package makefile

import (
	"fmt"
	"os"
	"strings"

	"github.com/kennytrytek/ktrai/internal/detect"
)

const targetMarker = "gen-context:"

// genContextTarget returns the gen-context Makefile snippet for the given language.
func genContextTarget(lang detect.Language) string {
	srcDir := srcDirFor(lang)
	langFlag := ""
	if l := lang.CtagsLanguages(); l != "" {
		langFlag = fmt.Sprintf(" --languages=%s", l)
	}
	return fmt.Sprintf(`gen-context: ## Regenerate AI context index (requires universal-ctags with JSON support)
	@ctags --list-output-formats 2>/dev/null | grep -q json || { echo "ctags found but lacks JSON output support — install universal-ctags (e.g. brew install universal-ctags)"; exit 1; }
	ctags --output-format=json --fields=+KSZnte%s \
	  -R %s \
	  | ktrai gen-symbols > .agent/context/symbols.md
`, langFlag, srcDir)
}

const prepTarget = `prep: gen-context ## Prepare for a commit (regenerate AI context)
`

func srcDirFor(lang detect.Language) string {
	switch lang {
	case detect.Go:
		return "."
	case detect.Python:
		return "."
	case detect.TypeScript:
		return "src"
	case detect.Java:
		return "src"
	default:
		return "."
	}
}

// Inject reads the Makefile at path and appends gen-context and prep targets
// if they are not already present. Does nothing if path does not exist.
func Inject(path string, lang detect.Language) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading Makefile: %w", err)
	}

	content := string(data)

	if strings.Contains(content, targetMarker) {
		return nil // already present, leave it alone
	}

	// Ensure the file ends with a newline before appending.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + genContextTarget(lang)

	if !strings.Contains(content, "prep:") {
		content += "\n" + prepTarget
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing Makefile: %w", err)
	}
	return nil
}
