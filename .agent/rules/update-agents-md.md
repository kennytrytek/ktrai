---
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

**Do invoke the `update-agents-md` skill** after any of these:
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
- Module responsibilities (one line per non-obvious entry-point or hotspot file — what it does, not what's in it; omit trivial glue files and test files)
- Hard conventions an agent must follow when writing code (naming, logging style, error handling)
- Commit pre-flight: the ordered sequence of formatter, linter, and test commands an agent must run before committing so that CI passes on the first push. Unit tests are included here; if they are slow, mark them as preferred-but-optional so the agent can skip them on tight iteration loops.

## What does NOT belong
- Function names, signatures, or class members — read source files directly for these
- Call-chain diagrams or data flow pseudocode listing function names
- Make command recipes or setup instructions
- Prose introductions, section numbers, horizontal rules, or explanatory narrative

## Human-only section
Every AGENTS.md ends with a `## Notes` section reserved for human annotation. Agents must:
- Never modify, reformat, truncate, or read this section for guidance.
- Skip past it entirely when updating.
- Emit it as an empty stub when creating a new file.

To execute an update or creation, invoke the `update-agents-md` skill.
