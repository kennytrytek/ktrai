---
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

1. **Changed files** — `git diff --name-only <base>...<head>`. Note count and paths.
2. **AGENTS.md** — Read it in full. This is the primary review baseline. Apply all conventions, module roles, and commit pre-flight rules documented there without re-deriving them.
3. **Existing review threads** — Fetch all before forming any findings. See [Deduplication](#deduplication).
4. **Change type** — Classify:
   - `adr` — Architecture Decision Record
   - `docs` — markdown content, templates, guides
   - `code` — logic, build scripts, config
   - `mixed` — both content and code
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

```markdown
## CI Review

**Tier**: [Low / Medium / High]
**Change type**: [code / docs / adr / mixed]
**Verdict**: [APPROVE / REQUEST_CHANGES / COMMENT]

### Acceptance criteria
- [criterion]: ✅ fully met / ⚠️ partially met / ❌ not addressed

### Findings

#### Blocking
- `path/to/file:N` — [specific, actionable description]

#### Advisory
- `path/to/file:N` — [specific description]

#### Nit
- `path/to/file:N` — Nit: [description]
```

For Low-tier PRs with no findings, output `✅ LGTM — no issues found.` and set verdict `APPROVE`.

Do not pad the report with:
- Comments on PR description format, branch naming, or other meta-process rules
- Repetition of what the code obviously does
- Generic advice that does not apply to this specific change

---

## Step 5 — Post the review

Use GitHub MCP (`user-github`):
1. `pull_request_review_write` with `method: "create"` to open a pending review.
2. `add_comment_to_pending_review` for each **new** finding — one per file+line. Omit anything already covered by an existing thread (see [Deduplication](#deduplication)).
3. `pull_request_review_write` with `method: "submit_pending"` to publish.
4. After submitting, react `+1` on any existing comments you agreed with but did not duplicate.

**Review body:**
- `REQUEST_CHANGES` or `COMMENT`: leave body **blank**. Inline comments are the entire review.
- `APPROVE`: body is `⛄️` and nothing else, unless a cross-cutting point cannot be anchored to any file or line — in that case `⛄️` followed by one sentence at most.

**Verdict:**
- `APPROVE` — no blocking issues. Nits get inline `Nit:` prefix.
- `REQUEST_CHANGES` — one or more blocking issues, or an acceptance criterion not met.
- `COMMENT` — draft PR or a PR where approval/rejection is not appropriate.

---

## Deduplication

Before writing any finding:
1. Fetch all existing review threads (human and bot).
2. For each thread, note: file, line, and substance of the feedback.
3. **Skip** any issue already raised — do not repeat open feedback.
4. **React `+1`** on existing comments you agree with and have nothing additive to say:
```bash
# Inline review comment
gh api repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions -f content='+1'

# General PR conversation comment
gh api repos/{owner}/{repo}/issues/comments/{comment_id}/reactions -f content='+1'
```

---

## ADR review rules

Apply when change type is `adr`, in addition to standard quality checks.

### Format
- Confirm the ADR follows the repo's established format (commonly: Title / Date / Status / Context / Decision / Considered alternatives / Consequences). Check existing ADRs for the expected structure.
- Status must be one of: `Proposed`, `Accepted`, `Deprecated`, or `Superseded by [ADR N]`.
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

## B. File patterns

Review these file types: `*.go`, `*.md`, `*.yml`, `*.yaml`

Skip without comment:
- `go.sum` — dependency lock file
- `.agent/context/symbols.md` — generated by `make gen` / `ktrai gen`; never edited directly
- `bin/` — build artifacts

## C. Test completeness

No test files exist yet; CI runs `go test ./...` and will catch failures. Any PR that adds or changes logic in `internal/` **must** include a corresponding `_test.go` file. There are no shared test helpers or fixtures — standard `testing.T` only.

Table-driven tests are the expected pattern for functions with multiple input cases (see how `scanModules`, `isTestFile`, and `parse` would be tested).

## D. WriteIfNotExists vs WriteFile

Every scaffold call in `cmd/align.go` must use the correct write function:
- `scaffold.WriteIfNotExists` — scaffolded content the user may have edited; must never overwrite
- `scaffold.WriteFile` — always-overwrite (for generated/derived files)

Using `WriteFile` where `WriteIfNotExists` is correct silently destroys user edits on re-run.

## E. Idempotency contract for scaffold functions

Every function in `internal/scaffold/scaffold.go` must be safe to call multiple times on the same path. Check that new functions: (1) inspect current state with `os.Lstat` before acting, and (2) treat an already-correct state as a no-op rather than an error.

`os.Lstat` must be used (not `os.Stat`) wherever symlink detection matters — AGENTS.md documents this convention; flag any violation.

## F. Error propagation in scaffold vs. scan functions

Two different error philosophies coexist:
- **Scaffold functions** (`scaffold.go`, `align.go`): must propagate all errors via `fmt.Errorf("context: %w", err)` — no silent swallowing.
- **Scan/walk functions** (`agentsmd.go:scanModules`): intentionally swallow `filepath.Walk` errors (best-effort scan). This is correct here, but the same pattern in a scaffold function would be a bug.

Flag any `_ = err` or dropped error in scaffold code. Accept it in scan code only when the function's purpose is explicitly best-effort.

## G. Adding new scaffolded content

The canonical pattern for new scaffolded files:
1. Add a string constant to `internal/scaffold/rules.go`
2. Wire it via `scaffold.WriteIfNotExists(filepath.Join(...), scaffold.TheConstant)` in `cmd/align.go`

Any PR that introduces new files written during `ktrai align` must follow both steps. A constant without a write call, or a write call using an inline string literal, are both incomplete.
