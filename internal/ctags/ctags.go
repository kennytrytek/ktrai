// Package ctags parses universal-ctags JSON output into Symbol structs.
package ctags

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"
	"strings"
)

// Symbol represents a single tag entry from ctags JSON output.
type Symbol struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Line      int    `json:"line"`
	Signature string `json:"signature"`
	Scope     string `json:"scope"`
	ScopeKind string `json:"scopeKind"`
	Inherits  string `json:"inherits"`
	Access    string `json:"access"`
	Type      string `json:"_type"`
}

// skipKinds are low-signal tag kinds that add noise without helping agents.
var skipKinds = map[string]bool{
	"variable":  true,
	"parameter": true,
	"unknown":   true,
	"local":     true,
	"label":     true,
	"constant":  true,
}

// Parse reads ctags JSON lines from r and returns symbols grouped by file path.
// Non-tag entries (e.g. metadata lines) are silently skipped.
func Parse(r io.Reader) (map[string][]Symbol, []string, error) {
	byFile := make(map[string][]Symbol)
	scanner := bufio.NewScanner(r)
	// ctags lines can be long (large signatures)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var sym Symbol
		if err := json.Unmarshal([]byte(line), &sym); err != nil {
			continue
		}
		if sym.Type != "tag" {
			continue
		}
		if skipKinds[sym.Kind] {
			continue
		}
		if sym.Path == "" || sym.Name == "" {
			continue
		}
		byFile[sym.Path] = append(byFile[sym.Path], sym)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	paths := make([]string, 0, len(byFile))
	for p := range byFile {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	return byFile, paths, nil
}
