# ktrai — agent context

ktrai — a CLI that scaffolds a `.agent/` directory layout in any repository, wiring AI agent context, rules, and skills for Cursor and Claude.

AGENTS.md generation design (scaling strategy, project archetypes, section scaffolding tiers):
`.agent/context/agents-md-design.md`

## Structure
| Path | Role |
|---|---|
| `cmd/` | CLI subcommands: `align` (scaffold + migrate repos) |
| `internal/agentsmd/` | Parse, draft, and render `.agent/context/AGENTS.md` |
| `internal/scaffold/` | Idempotent filesystem primitives: dirs, symlinks, file migration, skills migration |
| `internal/detect/` | Infer the primary programming language from root marker files |
| `internal/ctags/` | Parse universal-ctags JSON output into filtered `Symbol` structs |
| `internal/makefile/` | Inject `gen` and `prep` targets into an existing Makefile |
| `main.go` | Binary entry point |

## Hot Spots
Frequently changed files (last 12 months) — high-churn paths; understand these deeply before editing:
| Commits | File |
|---|---|
| 4 | `internal/scaffold/rules.go` |
| 4 | `cmd/root.go` |
| 3 | `internal/scaffold/scaffold.go` |
| 3 | `cmd/gen_symbols.go` |
| 3 | `Makefile` |
| 2 | `cmd/align.go` |
| 2 | `cmd/restructure.go` |
| 2 | `internal/makefile/makefile.go` |

## Coupling Clusters
Files that frequently change together — touching one likely requires touching the other:
| Co-commits | Pair |
|---|---|
| 2 | `cmd/root.go` ↔ `internal/scaffold/rules.go` |
| 2 | `cmd/root.go` ↔ `internal/scaffold/scaffold.go` |
| 2 | `cmd/root.go` ↔ `cmd/gen_symbols.go` |
| 2 | `cmd/root.go` ↔ `cmd/align.go` |
| 2 | `cmd/align.go` ↔ `internal/scaffold/rules.go` |
| 2 | `Makefile` ↔ `internal/scaffold/rules.go` |

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
