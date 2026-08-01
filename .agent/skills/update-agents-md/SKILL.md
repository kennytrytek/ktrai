---
name: update-agents-md
description: >
  Create or update AGENTS.md for this repository. Invoke when the
  update-agents-md rule fires, or manually when you want to refresh the agent
  context document.
---

# Creating or Updating AGENTS.md

## Choosing a scenario

Read the existing AGENTS.md (if any) before deciding:

- **Scenario A** — file exists AND has substantive content: real module descriptions, a genuine purpose sentence, and conventions that reflect the actual codebase. A few stale rows do not disqualify it.
- **Scenario B** — no file exists, OR the file is a stub or severely outdated: the majority of module roles are TODO placeholders, the purpose is a TODO, or the descriptions are so vague or wrong that they would mislead an agent. Treat as Scenario B and rebuild from source. Before writing, note any rows or bullets in the existing file that contain real information — carry those forward verbatim into the new document.

## Scenario A — Updating an existing AGENTS.md

**When to run:** only after complex updates, major new features, or significant refactors. Very small or localized code changes do not warrant an update.

1. Run `git log --oneline -20` to understand what changed since the last update.
2. Read the current AGENTS.md in full — stop at `## Notes` and do not read past it.
3. For each changed module, read the source file and any associated docs (README sections, ADRs, changelogs) to verify the existing description is still accurate.
4. Update only the affected table row or bullet; preserve everything else that remains accurate.
5. Refresh the `## Hot Spots`, `## Coupling Clusters`, and `## Stabilized Core` sections if any are absent, or if this update covers a major refactor or more than ~3 months of changes: re-run the corresponding git commands from Scenario B Step 1 and replace the table rows.
6. Do not touch the `## Notes` section.
7. Do not expand AGENTS.md beyond its current scope. If in doubt, leave it out.

## Scenario B — Creating or rebuilding AGENTS.md

**Step 1 — gather git history.** Run all three analyses. The results will be written directly into the document in Step 5 as hard-to-compute insights agents can reference without re-running git.

**1a — Hot spots** (frequently changed paths):
```sh
# Recent (last 12 months)
git log --since="1 year ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Mid-range (1–3 years ago)
git log --before="1 year ago" --since="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Foundational (inception to 3 years ago)
git log --before="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20
```
Files appearing frequently in recent history → active hotspots; read and describe these first.
Files present since early commits and still present today → foundational.

**1b — Coupling clusters** (files that change together):
```sh
git log --name-only --pretty=format:"---" | awk '
/^---$/{
  if(n>1) for(i=1;i<=n;i++) for(j=i+1;j<=n;j++) print files[i]" "files[j];
  n=0; next
}
NF{ files[++n]=$0 }
' | sort | uniq -c | sort -rn | head -15
```
Filter pairs where both files are source files (exclude lock files, generated files, `.agent/` meta-files). A pair with 2+ co-commits is meaningful on repos up to ~6 months old; use 3+ for older repos.

**1c — Stabilized core** (battle-tested paths with no recent churn):
```sh
# Adjust the cutoff based on project age:
#   < 3 months old  → use --before="2 weeks ago" / --since="2 weeks ago"
#   3–12 months old → use --before="3 months ago" / --since="3 months ago"
#   > 12 months old → use --before="1 year ago"  / --since="1 year ago"
comm -23 \
  <(git log --before="<CUTOFF>" --name-only --pretty=format: | grep -v '^$' | sort -u) \
  <(git log --since="<CUTOFF>"  --name-only --pretty=format: | grep -v '^$' | sort -u)
```
Files in the output have a commit history but have not been touched recently. These are battle-tested; agents should change them with extra care. Omit this section if the repo is too young (no files qualify).

**Tier-aware presentation** — count source files to choose the right level of detail for all three sections:
```sh
# Go
find . -name "*.go" -not -path "*/vendor/*" | wc -l
# JS/TS: *.ts *.tsx *.js  Python: *.py  etc.
```
| Count | Present entries as |
|---|---|
| < 15 (Nano) | Individual file paths |
| 15 – 100 (Small) | Individual file paths |
| 100 – 500 (Medium) | Directory paths (`internal/auth/`) |
| 500+ (Large) | Architectural domains; omit granular tables if signal is too diffuse |

**Step 2 — read all documentation.**
Read top-level and subdirectory READMEs, any ADRs (`docs/adr/`, `doc/decisions/`, or similar), changelogs, and any developer guides (`CONTRIBUTING.md`, `DEVELOPMENT.md`, `docs/` tree). These often reveal purpose, architecture decisions, and conventions that are not visible from source code alone.

**Step 3 — read the hotspot source files.**
Open each file identified in Step 1. Read enough to write a one-sentence role description. Do not skip this step — documentation may describe intent, but source is the ground truth for what a module actually does.

**Step 4 — discover conventions and commit pre-flight commands.**
Check these locations in order; use whatever you find:
- Formatter / linter config: `.golangci.yml`, `.eslintrc*`, `pyproject.toml` `[tool.*]` sections, `.rubocop.yml`, etc.
- Test / build commands: `Makefile` targets, `package.json` scripts, `.github/workflows/`, `Jenkinsfile`, `tox.ini`.
- Existing code patterns: scan a handful of hotspot files for error-handling style, naming conventions, and logging approach.
Only include a convention if you can verify it from the repo — never invent placeholders.

**Step 5 — write using this template:**
```markdown
# <repo-name> — agent context

<repo-name> — <one sentence: what the tool does and which systems it touches>.

## Modules
| File | Role |
|---|---|
| `path/to/module` | One-line description of what this module does. |

## Hot Spots
Frequently changed files (last 12 months) — high-churn paths; understand these deeply before editing:
| Commits | File |
|---|---|
| N | `path/to/file` |

## Coupling Clusters
Files that frequently change together — touching one likely requires touching the other:
| Co-commits | Pair |
|---|---|
| N | `fileA` ↔ `fileB` |

## Stabilized Core
Paths with substantial history but no recent changes — battle-tested; change with extra care:
- `path/to/stable/file`

## Conventions
- <verified naming / style / error-handling rule>

## Commit pre-flight
Run in order before every commit:
1. `<formatter command>` — auto-fixes style
2. `<linter command>` — must pass with zero warnings
3. `<test command>` — preferred; skip only when iteration speed matters (mark slow tests as optional)

## Notes
<!-- reserved for human annotation — agents must not modify this section -->
```
