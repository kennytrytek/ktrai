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

// UpdateAgentsMdRule is the always-available rule that tells agents when to
// invoke the update-agents-md skill and what AGENTS.md should contain.
const UpdateAgentsMdRule = `---
description: >
  Invoke the update-agents-md skill after any change with lasting effects on
  how developers or agents work in this repo: new or removed modules, changed
  formatter / linter / test commands, new external dependencies or services,
  toolchain migrations, architectural restructuring, or CI workflow changes.
  Do NOT invoke for routine bug fixes, small refactors, or localized changes
  that leave the repo's tools and structure intact.
alwaysApply: false
---

# When to update AGENTS.md

**Do invoke the ` + "`update-agents-md`" + ` skill** after any of these:
- A new module, package, or service is added or removed
- The formatter, linter, test runner, or build command changes
- A toolchain or runtime migration (e.g. Python virtualenv strategy, Node version, Go toolchain)
- A new external dependency or system integration is introduced
- An architectural restructuring that changes how the codebase is organized
- A CI/CD workflow change that alters what must pass before merging

**Do not invoke** after:
- Routine bug fixes or small localized refactors
- Changes that touch only one module without altering its purpose or the repo's conventions
- Documentation-only edits that do not reflect a structural change

## What belongs in AGENTS.md
Only high-level context an agent cannot easily infer by reading individual source files:
- One-sentence purpose: what the tool does and which systems it touches
- **Structure** — a directory-level code map (≤ 20 rows, one per package/subsystem/service). The unit is a stable architectural concern, not an individual file. See the skill for depth-calibration, tier rules, and special-directory handling.
- **Commands** table (Small tier and above): quick-reference for common dev tasks extracted from the Makefile or package.json scripts
- **Architecture** stub (Large / complex tier only): high-level subsystem description for repos with 500+ source files or monorepos
- **Patterns** stub (Medium tier and above): non-standard idioms agents must follow (error handling, concurrency, code generation, etc.)
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
- Emit it as an empty stub when creating a new file.

To execute an update or creation, invoke the ` + "`update-agents-md`" + ` skill.
`

// UpdateAgentsMdSkill contains the imperative steps for creating or updating
// AGENTS.md. It is written as a skill so it can be invoked both automatically
// (when the rule fires) and manually by a human.
const UpdateAgentsMdSkill = `---
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

1. Run ` + "`git log --oneline -20`" + ` to understand what changed since the last update.
2. Read the current AGENTS.md in full — stop at ` + "`## Notes`" + ` and do not read past it.
3. For each changed module, read the source file and any associated docs (README sections, ADRs, changelogs) to verify the existing description is still accurate.
4. Update only the affected table row or bullet; preserve everything else that remains accurate.
5. Do not touch the ` + "`## Notes`" + ` section.
6. Do not expand AGENTS.md beyond its current scope. If in doubt, leave it out.

## Scenario B — Creating or rebuilding AGENTS.md

**Step 1 — gather git history** across three time slices to identify hotspots:
` + "```" + `sh
# Recent (last 12 months)
git log --since="1 year ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Mid-range (1–3 years ago)
git log --before="1 year ago" --since="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20

# Foundational (inception to 3 years ago)
git log --before="3 years ago" --name-only --pretty=format: | sort | uniq -c | sort -rn | head -20
` + "```" + `

Interpret results:
- Files appearing frequently in recent history → active hotspots; read and describe these first.
- Files present since early commits and still present today → foundational; describe their role if non-obvious.

**Step 2 — read all documentation.**
Read top-level and subdirectory READMEs, any ADRs (` + "`docs/adr/`" + `, ` + "`doc/decisions/`" + `, or similar), changelogs, and any developer guides (` + "`CONTRIBUTING.md`" + `, ` + "`DEVELOPMENT.md`" + `, ` + "`docs/`" + ` tree). These often reveal purpose, architecture decisions, and conventions that are not visible from source code alone.

**Step 3 — read the hotspot source files.**
Open each file identified in Step 1. Read enough to write a one-sentence role description. Do not skip this step — documentation may describe intent, but source is the ground truth for what a module actually does.

**Step 4 — discover conventions and commit pre-flight commands.**
Check these locations in order; use whatever you find:
- Formatter / linter config: ` + "`.golangci.yml`" + `, ` + "`.eslintrc*`" + `, ` + "`pyproject.toml`" + ` ` + "`[tool.*]`" + ` sections, ` + "`.rubocop.yml`" + `, etc.
- Test / build commands: ` + "`Makefile`" + ` targets, ` + "`package.json`" + ` scripts, ` + "`.github/workflows/`" + `, ` + "`Jenkinsfile`" + `, ` + "`tox.ini`" + `.
- Existing code patterns: scan a handful of hotspot files for error-handling style, naming conventions, and logging approach.
Only include a convention if you can verify it from the repo — never invent placeholders.

**Step 5 — determine project tier.**

Count non-test, non-vendor source files to pick the right tier:

| Tier | Source file count | Sections to include |
|---|---|---|
| Nano | < 15 | Structure + Conventions |
| Small | 15–100 | Structure + Commands + Conventions |
| Medium | 100–500 | Structure + Commands + Patterns (stub) + Conventions |
| Large / complex | 500+ or monorepo | Structure + Commands + Architecture (stub) + Patterns (stub) + Conventions |

**Step 6 — build the Structure table.**

The unit of a Structure row is a **directory** (package, subsystem, service, or content section) — never an individual file. Rows represent stable architectural concerns that survive PRs; individual file listings go stale on every commit.

Rules:
- **Cap at 20 rows.** If the first depth produces > 20 source-containing directories, go shallower: use the parent directory and enumerate notable children in parentheses in the Role column.
- **Dominant package rule.** When one directory holds > 40% of the repo's source files, enumerate its key sub-packages in the Role: ` + "`Domain logic — documents, tables, linking, permissions, …`" + `
- **Always one row** regardless of file count: ` + "`docs/`" + `, ` + "`documentation/`" + `, ` + "`helm/`" + `, ` + "`content/`" + `, ` + "`static/`" + `, ` + "`assets/`" + `, ` + "`templates/`" + `, ` + "`idl/`" + `, ` + "`proto/`" + `, ` + "`scripts/`" + `
- **Never list as rows:** ` + "`vendor/`" + `, ` + "`node_modules/`" + `, ` + "`.venv/`" + `, ` + "`__pycache__/`" + `, ` + "`testdata/`" + `, ` + "`testutils/`" + `
- For config/infra-only repos (zero source files): use configuration concerns (Helm, CloudFormation, Terraform, monitoring rules) as the structural units.

**Step 7 — write the document.**

Include only the sections required by the tier from Step 5. Remove the ` + "`← omit`" + ` lines before writing.
` + "```" + `markdown
# <repo-name> — agent context

<repo-name> — <one sentence: what it does and which systems it touches>.

## Structure
| Path | Role |
|---|---|
| ` + "`dir/`" + ` | One-line description of what this concern does. |

## Commands              ← omit for Nano
| Task | Command |
|---|---|
| Build | ` + "`<command>`" + ` |
| Test | ` + "`<command>`" + ` |
| Lint | ` + "`<command>`" + ` |

## Architecture          ← include for Large / complex only; omit for Nano / Small / Medium
TODO: describe major subsystems and how they interact.

| Component | Role |
|---|---|

## Patterns              ← include for Medium and above; omit for Nano / Small
TODO: document non-standard idioms agents must follow (error handling, concurrency, mock generation, etc.).

## Conventions
- <verified naming / style / error-handling rule>

## Commit pre-flight
Run in order before every commit:
1. ` + "`<formatter command>`" + ` — auto-fixes style
2. ` + "`<linter command>`" + ` — must pass with zero warnings
3. ` + "`<test command>`" + ` — preferred; skip only when iteration speed matters (mark slow tests as optional)

## Notes
<!-- reserved for human annotation — agents must not modify this section -->
` + "```" + `
`

// UpdateCiReviewRule is the rule that tells agents when to invoke the
// update-ci-review skill and what triggers a refresh of the review criteria.
const UpdateCiReviewRule = `---
description: >
  Invoke the update-ci-review skill after any change with lasting effects on
  the codebase structure, conventions, or test patterns. Run AFTER AGENTS.md
  has been updated — update-ci-review reads AGENTS.md as its primary source.
  Do NOT invoke for routine bug fixes, small refactors, or localized changes.
alwaysApply: false
---

# When to update ci-review

**Do invoke the ` + "`update-ci-review`" + ` skill** after any of these:
- A new module, package, or service is added or removed
- A new framework, library, or external dependency is introduced
- Test utilities, fixtures, or coverage expectations change significantly
- An architectural restructuring that changes file/directory patterns
- Deprecated patterns or APIs are identified that the review should catch

**Do not invoke** after:
- Routine bug fixes or small localized refactors
- Changes that do not affect the codebase's structure, patterns, or conventions
- Documentation-only edits

**Prerequisites**: Invoke ` + "`update-agents-md`" + ` first if AGENTS.md is also stale —
` + "`update-ci-review`" + ` reads AGENTS.md as its primary source and will produce better
output if AGENTS.md is current.

To execute an update or creation, invoke the ` + "`update-ci-review`" + ` skill.
`

// CiReviewSkill is the skill that agents invoke to perform an automated,
// non-interactive code review. It is scaffolded as a starter template;
// sections B–N are populated by the update-ci-review skill.
const CiReviewSkill = `---
name: ci-review
description: Non-interactive CI code review for pull requests. Runs automatically against this repo's standards and outputs a structured Markdown report. No user interaction.
allowed-tools: Bash(git:*), Read, Grep, Glob
---

# CI Code Review

Fully autonomous. No user interaction at any step — gather, review, and post in a single unattended pass.

## Effort tiers

Classify the PR's complexity early. The tier affects verdict behavior, not whether to pause for input.

| Tier | Signal |
|------|--------|
| **Low** | ≤ ~5 files changed, narrow focused change (typo fix, single-field addition, config tweak, dependency bump, small refactor with no logic change) |
| **Medium** | Moderate scope, a few interdependent files, some logic changes |
| **High** | Large diff, architectural impact, multiple subsystems, risky refactor, security-sensitive change |

When in doubt, err toward the lower tier.

---

## Step 1 — Gather context

1. **Changed files** — ` + "`git diff --name-only <base>...<head>`" + `. Note count and paths.
2. **AGENTS.md** — Read it in full. This is the primary review baseline. Apply all conventions, module roles, and commit pre-flight rules documented there without re-deriving them.
3. **Existing review threads** — Fetch all before forming any findings. See [Deduplication](#deduplication).
4. **Change type** — Classify:
   - ` + "`adr`" + ` — Architecture Decision Record
   - ` + "`docs`" + ` — markdown content, templates, guides
   - ` + "`code`" + ` — logic, build scripts, config
   - ` + "`mixed`" + ` — both content and code
5. **Assign effort tier** using the table above.

---

## Step 2 — Read the changed files

For each changed file:
- Read the full file, not just the diff hunks — surrounding context matters.
- Verify every claim against actual file content before writing it. Do not rely on diff output alone or make assumptions about what code does without reading it.
- When a changed file is a test file, read the corresponding source file too.

---

## Step 3 — Review

### Primary: acceptance criteria

If the PR description contains acceptance criteria or a linked ticket:
- For each criterion, identify which changed files/lines address it.
- Determine: **fully met** / **partially met** / **not addressed**.
- If partially met or missing, state specifically what is lacking.

### Secondary: code quality

**Apply AGENTS.md conventions as the baseline.** Do not re-state anything AGENTS.md already documents — apply it silently.

Then check the repo-specific criteria in Sections A–N below, followed by these universal checks:
- Confusing or unnecessarily complex logic
- Missing tests for logic changes
- Error handling gaps
- Security or performance concerns where relevant (flag; do not fix)
- Missing or outdated documentation for user-facing changes

**For ADR changes:** apply the [ADR review rules](#adr-review-rules) below.

---

## Step 4 — Output the report

` + "```" + `markdown
## CI Review

**Tier**: [Low / Medium / High]
**Change type**: [code / docs / adr / mixed]
**Verdict**: [APPROVE / REQUEST_CHANGES / COMMENT]

### Acceptance criteria
- [criterion]: ✅ fully met / ⚠️ partially met / ❌ not addressed

### Findings

#### Blocking
- ` + "`path/to/file:N`" + ` — [specific, actionable description]

#### Advisory
- ` + "`path/to/file:N`" + ` — [specific description]

#### Nit
- ` + "`path/to/file:N`" + ` — Nit: [description]
` + "```" + `

For Low-tier PRs with no findings, output ` + "`✅ LGTM — no issues found.`" + ` and set verdict ` + "`APPROVE`" + `.

Do not pad the report with:
- Comments on PR description format, branch naming, or other meta-process rules
- Repetition of what the code obviously does
- Generic advice that does not apply to this specific change

---

## Step 5 — Post the review

Use GitHub MCP (` + "`user-github`" + `):
1. ` + "`pull_request_review_write`" + ` with ` + "`method: \"create\"`" + ` to open a pending review.
2. ` + "`add_comment_to_pending_review`" + ` for each **new** finding — one per file+line. Omit anything already covered by an existing thread (see [Deduplication](#deduplication)).
3. ` + "`pull_request_review_write`" + ` with ` + "`method: \"submit_pending\"`" + ` to publish.
4. After submitting, react ` + "`+1`" + ` on any existing comments you agreed with but did not duplicate.

**Review body:**
- ` + "`REQUEST_CHANGES`" + ` or ` + "`COMMENT`" + `: leave body **blank**. Inline comments are the entire review.
- ` + "`APPROVE`" + `: body is empty, unless a cross-cutting point cannot be anchored to any file or line — in that case one sentence at most.

**Verdict:**
- ` + "`APPROVE`" + ` — no blocking issues. Nits get inline ` + "`Nit:`" + ` prefix.
- ` + "`REQUEST_CHANGES`" + ` — one or more blocking issues, or an acceptance criterion not met.
- ` + "`COMMENT`" + ` — draft PR or a PR where approval/rejection is not appropriate.

---

## Deduplication

Before writing any finding:
1. Fetch all existing review threads (human and bot).
2. For each thread, note: file, line, and substance of the feedback.
3. **Skip** any issue already raised — do not repeat open feedback.
4. **React ` + "`+1`" + `** on existing comments you agree with and have nothing additive to say:
` + "```" + `bash
# Inline review comment
gh api repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions -f content='+1'

# General PR conversation comment
gh api repos/{owner}/{repo}/issues/comments/{comment_id}/reactions -f content='+1'
` + "```" + `

---

## ADR review rules

Apply when change type is ` + "`adr`" + `, in addition to standard quality checks.

### Format
- Confirm the ADR follows the repo's established format (commonly: Title / Date / Status / Context / Decision / Considered alternatives / Consequences). Check existing ADRs for the expected structure.
- Status must be one of: ` + "`Proposed`" + `, ` + "`Accepted`" + `, ` + "`Deprecated`" + `, or ` + "`Superseded by [ADR N]`" + `.
- Numbering must be consistent with existing ADRs.

### Problem statement (Context)
- Is the problem clearly stated and scoped? A reader unfamiliar with the codebase should understand what constraint is being addressed.
- Are the key constraints that ruled out doing nothing stated explicitly?
- Watch for hidden assumptions — things the author knows but did not write down.

### The decision
- **Logical consistency**: do the stated constraints lead to the stated decision? Does the decision fully solve the problem in Context, or only partially?
- **Invariants**: if invariants are introduced, check whether anything else in the ADR contradicts them.
- **Runtime assumptions**: if the decision depends on ordering guarantees or component availability, are those assumptions stated and is the behavior when they are violated documented?

### Considered alternatives
- At least one rejected alternative must be present. A decision with no alternatives considered is incomplete.
- Each rejection must state *why* the alternative was ruled out — not just that it was.
- Check for overlooked alternatives: lazy loading, feature-flag approach, doing nothing with documented trade-offs.

### Consequences
- Migration impact: are existing consumers called out? Is a migration path referenced?
- Phased behavior: if a temporary exception exists, it must be marked intentional and time-bounded with a follow-up ticket referenced.
- Operational consequences: version coupling, deployment ordering, ownership changes.
- Cross-references to related ADRs, tickets, or parallel work.

### Language and terminology
- Key terms should be defined or linked. Check that the ADR uses the repo's preferred terms.
- Watch for ambiguous terms used without qualification.

### Tone
- Phrase findings as **considerations**, not corrections. Use "it is worth considering whether..." rather than "this is wrong because...".
- Distinguish between "the decision is questionable" and "the decision is good but this section needs clarification."

---

## A. Repo standards (AGENTS.md)

AGENTS.md is already in context from Step 1. Apply all conventions, module roles, and commit pre-flight rules documented there as the review baseline without re-stating them here.

## B–N. Repo-specific review criteria

_Not yet populated. Run the_ ` + "`update-ci-review`" + ` _skill to generate repo-specific sections._
`

// UpdateCiReviewSkill contains the imperative steps for creating or refreshing
// the ci-review skill. It is lightweight by design — AGENTS.md does the heavy
// lifting; this skill only fills the review-specific delta AGENTS.md omits.
const UpdateCiReviewSkill = `---
name: update-ci-review
description: >
  Create or refresh the ci-review skill for this repository. Reads AGENTS.md as
  the primary source of repo context, then scans for the four review-specific
  details AGENTS.md does not cover. Invoke when the update-ci-review rule fires,
  or manually when you want to refresh the review criteria.
---

# Creating or Updating ci-review

` + "`update-ci-review`" + ` builds on top of ` + "`update-agents-md`" + ` — it does not re-scan the
full codebase. AGENTS.md already captures module roles, conventions, and commit
pre-flight. This skill only discovers the review-specific delta that AGENTS.md
does not provide.

## Choosing a scenario

Read the existing ` + "`.agent/skills/ci-review/SKILL.md`" + ` (if any) before deciding:

- **Scenario A** — file exists and sections B–N contain real, repo-specific rules (not TODO placeholders). A few stale entries do not disqualify it.
- **Scenario B** — no file exists, or sections B–N are stubs or placeholders. Treat as Scenario B and populate from source.

## Step 1 — Read AGENTS.md

Read ` + "`.agent/context/AGENTS.md`" + ` in full. This is the primary source of repo context:
- Module responsibilities tell you which directories contain code worth reviewing closely.
- Conventions tell you what patterns are already required — do not add review rules for things AGENTS.md already enforces.
- Commit pre-flight tells you what the formatter and linter catch — do not duplicate those either.

## Step 2 — Scan for the four review-specific gaps

AGENTS.md does not capture these; discover them now.

### 2a — File inclusion/exclusion globs

Check the repo tree for paths that should be excluded from review:
- Generated files: ` + "`**/generated/**`" + `, ` + "`**/*.generated.*`" + `, proto/gRPC output, etc.
- Lock files: ` + "`pnpm-lock.yaml`" + `, ` + "`package-lock.json`" + `, ` + "`yarn.lock`" + `, ` + "`go.sum`" + `
- Vendored code: ` + "`**/vendor/**`" + `, ` + "`**/node_modules/**`" + `
- Build artifacts: ` + "`**/dist/**`" + `, ` + "`**/__pycache__/**`" + `

Determine ` + "`file-patterns`" + ` (what to include) from the language extensions present in the repo. Tailor to what is actually present — do not use generic defaults blindly.

### 2b — Test utilities and completeness expectations

Read up to 5 test files from modules identified as active in AGENTS.md. Look for:
- Shared fixtures, factories, or test helpers the team uses
- Whether integration tests exist and when they are required
- Test naming conventions or table-driven patterns
- Coverage thresholds in CI config (` + "`.github/workflows/`" + `, ` + "`Makefile`" + `, ` + "`codecov.yml`" + `, etc.)

### 2c — Deprecated patterns

Scan recently changed files for signals of in-progress migrations:
` + "```" + `sh
git log --since="6 months ago" --name-only --pretty=format: | sort -u | head -30
` + "```" + `
Look for:
- Comments containing ` + "`Deprecated`" + `, ` + "`TODO: migrate`" + `, ` + "`TODO: remove`" + `, ` + "`FIXME`" + `
- Annotations like ` + "`@Deprecated`" + ` (Java/Kotlin), ` + "`# noqa`" + ` patterns, ` + "`//nolint`" + ` directives
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

Write the full file using the ` + "`CiReviewSkill`" + ` template embedded in ktrai. Then populate sections B–N with what you found in Step 2.

**Section content rules:**
- 30–60 lines of dense, repo-specific rules across all added sections
- No generic advice that applies to every codebase
- No rules already enforced by the linter or formatter (AGENTS.md commit pre-flight tells you what those are)
- No rules already covered in AGENTS.md conventions
- If a gap (2a–2d) yielded nothing actionable, add a comment stub noting what the team should fill in — do not fabricate content
`
