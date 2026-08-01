---
name: update-ci-review
description: >
  Create or refresh the ci-review skill for this repository. Reads AGENTS.md as
  the primary source of repo context, then scans for the four review-specific
  details AGENTS.md does not cover. Invoke when the update-ci-review rule fires,
  or manually when you want to refresh the review criteria.
---

# Creating or Updating ci-review

`update-ci-review` builds on top of `update-agents-md` — it does not re-scan the
full codebase. AGENTS.md already captures module roles, conventions, and commit
pre-flight. This skill only discovers the review-specific delta that AGENTS.md
does not provide.

## Choosing a scenario

Read the existing `.agent/skills/ci-review/SKILL.md` (if any) before deciding:

- **Scenario A** — file exists and sections B–N contain real, repo-specific rules (not TODO placeholders). A few stale entries do not disqualify it.
- **Scenario B** — no file exists, or sections B–N are stubs or placeholders. Treat as Scenario B and populate from source.

## Step 1 — Read AGENTS.md

Read `.agent/context/AGENTS.md` in full. This is the primary source of repo context:
- Module responsibilities tell you which directories contain code worth reviewing closely.
- Conventions tell you what patterns are already required — do not add review rules for things AGENTS.md already enforces.
- Commit pre-flight tells you what the formatter and linter catch — do not duplicate those either.

## Step 2 — Scan for the four review-specific gaps

AGENTS.md does not capture these; discover them now.

### 2a — File inclusion/exclusion globs

Check the repo tree for paths that should be excluded from review:
- Generated files: `**/generated/**`, `**/*.generated.*`, proto/gRPC output, etc.
- Lock files: `pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`, `go.sum`
- Vendored code: `**/vendor/**`, `**/node_modules/**`
- Build artifacts: `**/dist/**`, `**/__pycache__/**`

Determine `file-patterns` (what to include) from the language extensions present in the repo. Tailor to what is actually present — do not use generic defaults blindly.

### 2b — Test utilities and completeness expectations

Read up to 5 test files from modules identified as active in AGENTS.md. Look for:
- Shared fixtures, factories, or test helpers the team uses
- Whether integration tests exist and when they are required
- Test naming conventions or table-driven patterns
- Coverage thresholds in CI config (`.github/workflows/`, `Makefile`, `codecov.yml`, etc.)

### 2c — Deprecated patterns

Scan recently changed files for signals of in-progress migrations:
```sh
git log --since="6 months ago" --name-only --pretty=format: | sort -u | head -30
```
Look for:
- Comments containing `Deprecated`, `TODO: migrate`, `TODO: remove`, `FIXME`
- Annotations like `@Deprecated` (Java/Kotlin), `# noqa` patterns, `//nolint` directives
- Import aliases suggesting a replacement is in progress

For each deprecated pattern found, record the old form and the preferred replacement.

### 2d — "Beyond the linter" catches

Read 3–5 recently changed source files (from the list in 2c). Look for recurring structural patterns the formatter/linter does not enforce:
- DRY violations common to this codebase (copy-pasted blocks, duplicated logic)
- Missing abstractions the team clearly intends (a utility that should be used but is re-implemented inline)
- Resource or goroutine leak patterns specific to this stack
- Framework-specific patterns (state management rules, component patterns, missing resolvers for non-scalar fields)

Only document patterns you actually observe — never fabricate rules.

## Step 3 — Write ci-review/SKILL.md

### Scenario A — update existing

Read the existing file. Keep Section A (AGENTS.md baseline), the workflow steps, the deduplication section, and the ADR review rules unchanged. Replace only sections B–N with what you found in Step 2.

### Scenario B — create from scratch

Write the full file using the `CiReviewSkill` template embedded in ktrai. Then populate sections B–N with what you found in Step 2.

**Section content rules:**
- 30–60 lines of dense, repo-specific rules across all added sections
- No generic advice that applies to every codebase
- No rules already enforced by the linter or formatter (AGENTS.md commit pre-flight tells you what those are)
- No rules already covered in AGENTS.md conventions
- If a gap (2a–2d) yielded nothing actionable, add a comment stub noting what the team should fill in — do not fabricate content
