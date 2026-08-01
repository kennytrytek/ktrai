package scaffold

// CodemapRule is the manually-invoked rule that directs agents to read AGENTS.md
// before exploring source files.
const CodemapRule = `---
description: Read the AI context index before any substantially complex implementation task.
alwaysApply: false
---

Before exploring source files, read ` + "`AGENTS.md`" + `.
Only read individual source files when that is insufficient for the task.
`

// UpdateAgentsMdRule is the manually-invoked rule codifying how to create or update AGENTS.md.
const UpdateAgentsMdRule = `---
description: Use when creating or updating AGENTS.md. Keeps the file trim and agent-optimized.
alwaysApply: false
---

# Creating or Updating AGENTS.md

## What belongs in AGENTS.md
Only high-level context an agent cannot easily infer by reading individual source files:
- One-sentence purpose: what the tool does and which systems it touches
- Module responsibilities (one line per file — what it does, not what's in it)
- Hard conventions an agent must follow when writing code (naming, logging style, error handling)
- Commit pre-flight: the ordered sequence of formatter, linter, and test commands an agent must run before committing so that CI passes on the first push. Unit tests are included here; if they are slow, mark them as preferred-but-optional so the agent can skip them on tight iteration loops.

## What does NOT belong
- Function names, signatures, or class members — read source files directly for these
- Call-chain diagrams or data flow pseudocode listing function names
- Make command recipes or setup instructions
- Prose introductions, section numbers, horizontal rules, or explanatory narrative

## Human-only section
Every AGENTS.md ends with a ` + "`## Notes`" + ` section reserved for human annotation. Agents must:
- Never modify, reformat, truncate, or read this section for guidance.
- Skip past it entirely when updating.
- Emit it as an empty stub when creating a new file (see template below).

## Scenario A — Updating an existing AGENTS.md

**When to run:** only after complex updates, major new features, or significant refactors. Very small or localized code changes do not warrant an update.

1. Read the current AGENTS.md in full — stop at ` + "`## Notes`" + ` and do not read past it.
2. Identify only what changed: new module, removed module, changed convention, new external system, or significantly revised logic.
3. Update only the affected table row or bullet; preserve everything else that remains accurate.
4. Do not touch the ` + "`## Notes`" + ` section.
5. Do not expand AGENTS.md beyond its current scope. If in doubt, leave it out.

## Scenario B — Creating a new AGENTS.md (none exists)

Sample git history across three time slices to identify hotspots and foundational areas before writing.

**Step 1 — gather history:**
` + "```" + `
# Recent (last 12 months)
git log --since="1 year ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Mid-range (1–3 years ago)
git log --before="1 year ago" --since="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Foundational (inception to 3 years ago)
git log --before="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20
` + "```" + `

**Step 2 — interpret:**
- Files appearing frequently in recent history → active hotspots; prioritize accurate module descriptions for these.
- Files appearing in old history and still present today → foundational; note them as such if their role is non-obvious.
- Changes within the last year are most relevant to the current project state; older commits may indicate more foundational areas of focus.

**Step 3 — write using this template:**
` + "```" + `markdown
# <repo-name> — agent context

<repo-name> — <one sentence: what the tool does and which systems it touches>.

## Modules
| File | Role |
|---|---|
| ` + "`path/to/file.go`" + ` | One-line description of what this module does. |

## Conventions
- <naming / style / error-handling rules>

## Commit pre-flight
Run in order before every commit:
1. ` + "`<formatter command>`" + ` — auto-fixes style
2. ` + "`<linter command>`" + ` — must pass with zero warnings
3. ` + "`<test command>`" + ` — preferred; skip only when iteration speed matters (mark slow tests as optional)

## Notes
<!-- reserved for human annotation — agents must not modify this section -->
` + "```" + `
`

// SymbolsPlaceholder is written when ctags is not available at init time.
const SymbolsPlaceholder = `# Symbol Index

_Not yet generated._

Run the following to generate this file:

` + "```" + `sh
make gen
` + "```" + `

Requires [universal-ctags](https://github.com/universal-ctags/ctags) with JSON support:

` + "```" + `sh
brew install universal-ctags
` + "```" + `
`
