---
description: >
  Invoke the update-ci-review skill after any change with lasting effects on
  the codebase structure, conventions, or test patterns. Run AFTER AGENTS.md
  has been updated — update-ci-review reads AGENTS.md as its primary source.
  Do NOT invoke for routine bug fixes, small refactors, or localized changes.
alwaysApply: false
---

# When to update ci-review

**Do invoke the `update-ci-review` skill** after any of these:
- A new module, package, or service is added or removed
- A new framework, library, or external dependency is introduced
- Test utilities, fixtures, or coverage expectations change significantly
- An architectural restructuring that changes file/directory patterns
- Deprecated patterns or APIs are identified that the review should catch

**Do not invoke** after:
- Routine bug fixes or small localized refactors
- Changes that do not affect the codebase's structure, patterns, or conventions
- Documentation-only edits

**Prerequisites**: Invoke `update-agents-md` first if AGENTS.md is also stale —
`update-ci-review` reads AGENTS.md as its primary source and will produce better
output if AGENTS.md is current.

To execute an update or creation, invoke the `update-ci-review` skill.
