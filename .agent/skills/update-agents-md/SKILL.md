---
name: update-agents-md
description: >
  Create or update AGENTS.md for this repository. Invoke when the
  update-agents-md rule fires, or manually when you want to refresh the agent
  context document.
---

# Creating or Updating AGENTS.md

## Scenario A — Updating an existing AGENTS.md

**When to run:** only after complex updates, major new features, or significant refactors. Very small or localized code changes do not warrant an update.

1. Run `git log --oneline -20` to understand what changed since the last update.
2. Read the current AGENTS.md in full — stop at `## Notes` and do not read past it.
3. For each changed module, read the source file and any associated docs (README sections, ADRs, changelogs) to verify the existing description is still accurate.
4. Update only the affected table row or bullet; preserve everything else that remains accurate.
5. Do not touch the `## Notes` section.
6. Do not expand AGENTS.md beyond its current scope. If in doubt, leave it out.

## Scenario B — Creating a new AGENTS.md (none exists)

**Step 1 — gather git history** across three time slices to identify hotspots:
```sh
# Recent (last 12 months)
git log --since="1 year ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Mid-range (1–3 years ago)
git log --before="1 year ago" --since="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Foundational (inception to 3 years ago)
git log --before="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20
```

Interpret results:
- Files appearing frequently in recent history → active hotspots; read and describe these first.
- Files present since early commits and still present today → foundational; describe their role if non-obvious.

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
