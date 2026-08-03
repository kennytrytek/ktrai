# ktrai — agent context

ktrai — a CLI that scaffolds a `.agent/` directory layout in any repository, wiring AI agent context, rules, and skills for Cursor and Claude.

AGENTS.md generation design (scaling strategy, project archetypes, section scaffolding tiers):
`.agent/context/agents-md-design.md`

## Structure
| Path | Role |
|---|---|
| `cmd/` | CLI subcommands: `align` (scaffold + migrate repos); `gen` (ctags JSON → Markdown symbol index, not registered in the default binary — call `RegisterGenSymbols()` explicitly) |
| `internal/agentsmd/` | Parse, draft, and render `.agent/context/AGENTS.md` |
| `internal/scaffold/` | Idempotent filesystem primitives: dirs, symlinks, file migration, skills migration |
| `internal/detect/` | Infer the primary programming language from root marker files |
| `internal/ctags/` | Parse universal-ctags JSON output into filtered `Symbol` structs |
| `internal/makefile/` | Inject `gen` and `prep` targets into an existing Makefile |
| `main.go` | Binary entry point |

## Hot Spots
<!-- No merge commits yet (repo < 1 month old, inception 2026-07-25). Data below reflects all commits; re-run the skill once merges exist. -->
Frequently changed files (all commits) — high-churn paths; understand these deeply before editing:
| Commits | File |
|---|---|
| 6 | `internal/scaffold/rules.go` |
| 4 | `cmd/root.go` |
| 4 | `cmd/align.go` |
| 3 | `internal/scaffold/scaffold.go` |
| 3 | `cmd/gen_symbols.go` |
| 2 | `internal/scaffold/workflows.go` |
| 2 | `internal/makefile/makefile.go` |
| 2 | `internal/detect/detect.go` |
| 2 | `internal/agentsmd/agentsmd.go` |

## Coupling Clusters
<!-- Same caveat: all-commit data, no merges yet. Source files only; .agent/, Makefile, README excluded. -->
Files that frequently change together — touching one likely requires touching the other:
| Co-commits | Pair |
|---|---|
| 3 | `cmd/root.go` ↔ `internal/scaffold/rules.go` |
| 3 | `cmd/root.go` ↔ `internal/scaffold/scaffold.go` |
| 3 | `cmd/align.go` ↔ `internal/scaffold/rules.go` |
| 2 | `internal/scaffold/rules.go` ↔ `internal/scaffold/scaffold.go` |

## Stabilized Core
<!-- Repo is < 1 month old (inception 2026-07-25) — no files qualify yet. Refresh when the repo has 3+ months of history. -->

## Conventions
- Wrap errors with `fmt.Errorf("context: %w", err)` — never discard or log-and-return.
- All filesystem operations must be idempotent: check existence before creating or modifying.
- Use `os.Lstat` (not `os.Stat`) whenever symlink detection matters.

## Commit pre-flight
Run in order before every commit:
1. `go build ./...` — must compile cleanly
2. `go vet ./...` — must pass with zero warnings
3. `go test ./...` — must pass

## Notes
<!-- reserved for human annotation — agents must not modify this section -->
