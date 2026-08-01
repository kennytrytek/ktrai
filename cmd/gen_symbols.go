package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kennytrytek/ktrai/internal/ctags"
	"github.com/spf13/cobra"
)

var genSymbolsCmd = &cobra.Command{
	Use:   "gen-symbols",
	Short: "Convert ctags JSON (stdin) to a Markdown symbol index (stdout)",
	Long: `Read universal-ctags --output-format=json output from stdin and write
a structured Markdown symbol index to stdout.

Typical usage in a Makefile:
  ctags --output-format=json --fields=+KSZnte -R ./src \
    | ktrai gen-symbols > .agent/context/symbols.md`,
	RunE: runGenSymbols,
}

// RegisterGenSymbols wires gen-symbols into the root command.
// Call from main or a dev build tag to enable it.
func RegisterGenSymbols() {
	rootCmd.AddCommand(genSymbolsCmd)
}

func runGenSymbols(_ *cobra.Command, _ []string) error {
	byFile, paths, err := ctags.Parse(os.Stdin)
	if err != nil {
		return fmt.Errorf("parsing ctags input: %w", err)
	}

	fmt.Printf("# Symbol Index\n\n")
	fmt.Printf("_Generated: %s_\n\n", time.Now().Format("2006-01-02"))

	for _, path := range paths {
		syms := byFile[path]
		if len(syms) == 0 {
			continue
		}

		fmt.Printf("## `%s`\n\n", path)

		// Separate top-level symbols from scoped members.
		var topLevel []ctags.Symbol
		scoped := make(map[string][]ctags.Symbol)

		for _, s := range syms {
			if s.Scope == "" {
				topLevel = append(topLevel, s)
			} else {
				scoped[s.Scope] = append(scoped[s.Scope], s)
			}
		}

		sort.Slice(topLevel, func(i, j int) bool {
			return topLevel[i].Line < topLevel[j].Line
		})

		for _, sym := range topLevel {
			line := formatTopLevel(sym)
			fmt.Println(line)

			if members, ok := scoped[sym.Name]; ok {
				sort.Slice(members, func(i, j int) bool {
					return members[i].Line < members[j].Line
				})
				for _, m := range members {
					fmt.Println(formatMember(m))
				}
				fmt.Println()
			} else {
				fmt.Println()
			}
		}
	}

	return nil
}

func formatTopLevel(s ctags.Symbol) string {
	var sb strings.Builder
	sb.WriteString("### ")
	if s.Access != "" && s.Access != "public" {
		sb.WriteString(fmt.Sprintf("_%s_ ", s.Access))
	}
	sb.WriteString(fmt.Sprintf("`%s` (%s)", s.Name, s.Kind))
	if s.Signature != "" {
		sb.WriteString(fmt.Sprintf(" `%s`", s.Signature))
	}
	if s.Inherits != "" {
		sb.WriteString(fmt.Sprintf(" — inherits `%s`", s.Inherits))
	}
	return sb.String()
}

func formatMember(s ctags.Symbol) string {
	var sb strings.Builder
	sb.WriteString("- ")
	if s.Access != "" && s.Access != "public" {
		sb.WriteString(fmt.Sprintf("_%s_ ", s.Access))
	}
	sb.WriteString(fmt.Sprintf("`%s` (%s)", s.Name, s.Kind))
	if s.Signature != "" {
		sb.WriteString(fmt.Sprintf(" `%s`", s.Signature))
	}
	return sb.String()
}
