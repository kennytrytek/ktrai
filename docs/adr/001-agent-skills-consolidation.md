# ADR 001: Agent Skills Consolidation

**Date:** 2026-08-01
**Status:** Accepted

## Context

`ktrai align` already consolidates always-on agent instructions (rules) under `.agent/rules/`, symlinked from `.cursor/rules/` and `.claude/rules/`. A parallel problem exists for on-demand agent procedures — skills — which have grown in three separate locations:

| Location | Description |
|---|---|
| `.claude/skills/` | Skills for Claude Code; the most common pattern across Workiva repos |
| `.cursor/skills/` | Skills for Cursor; sometimes individual per-skill symlinks pointing elsewhere |
| `.skills/` | A tool-agnostic root-level convention used in some repos as the canonical source |

This scattered layout means the same skill is often duplicated across two or three locations, edits made in one place are not reflected in others, and new repos have no clear home for skills.

### Concepts that matter here

**Rules** are passive, always-on context injected into the agent's system prompt automatically. They encode conventions, naming standards, error-handling style, and commit pre-flight steps — things the agent must always be aware of.

**Skills** are active, on-demand procedures invoked explicitly by the user or agent. They are imperative step-by-step guides that tell the agent *how* to perform a specific task (e.g., `ci-review`, `deploy-service`). Both Cursor and Claude Code understand the `SKILL.md` format, making skills a cross-tool concern suitable for consolidation.

**Commands** (`.cursor/commands/`) are Cursor-specific slash commands — UI affordances invoked via `/command-name` in the Cursor chat input. Claude Code has no equivalent concept, so commands cannot be meaningfully shared across tools and are excluded from this consolidation.

### Supported tools

Workiva supports two AI coding tools: **Cursor** and **Claude Code**. No other tool-specific agent configuration (e.g., GitHub Copilot's `.github/copilot-instructions.md`) is sanctioned. Configurations for unsupported tools are removed by `ktrai align` rather than migrated.

## Decision

Extend `ktrai align` to:

1. **Consolidate skills** from `.skills/`, `.claude/skills/`, and `.cursor/skills/` into `.agent/skills/` as the single canonical location.
2. **Replace** each source skills directory with a symlink pointing to `.agent/skills/`, so both Cursor and Claude Code continue to find skills at their expected paths.
3. **Remove** configuration files for unsupported AI tools (currently: `.github/copilot-instructions.md`).

### Target layout after `align`

```
.agent/
├── context/
│   ├── AGENTS.md
│   └── symbols.md
├── rules/
│   └── update-agents-md.md
└── skills/
    └── <skill-name>/
        └── SKILL.md

.cursor/rules    → symlink → .agent/rules
.cursor/skills   → symlink → .agent/skills
.claude/rules    → symlink → .agent/rules
.claude/skills   → symlink → .agent/skills
.skills          → symlink → .agent/skills
```

### Migration order and collision policy

Skills are migrated in priority order: `.skills/` first (tool-agnostic canonical), then `.claude/skills/`, then `.cursor/skills/`. If the same skill name already exists in `.agent/skills/` when a subsequent source is processed, `align` **errors** rather than silently overwriting — the two variants may have genuinely different content targeting different tools (as observed with `setup-wk-mcp` in `wk-plugin-mcp`). The user must rename or merge the conflicting skills before re-running.

### What is not touched

- `.cursor/commands/` — Cursor-only slash commands; no Claude Code equivalent
- `.claude/hooks/`, `.claude/settings.json`, `.claude/pr-review.md` — Claude Code runtime configuration, not agent knowledge
- `.cursor/environment.json`, `.cursor/mcp.json` — Cursor workspace configuration
- Sub-directory `AGENTS.md` files (e.g., `terraform/datadog/AGENTS.md`) — out of `align`'s root scope

## Consequences

- All skills live in one place; edits propagate to both tools automatically via symlinks.
- Removing support for unsupported tools is now an automated step rather than a manual cleanup.
- The collision error provides a safe guard when per-tool skill variants exist; it forces an explicit human decision rather than silently losing content.
- The extensible unsupported-configs list in `cmd/align.go` makes it straightforward to add future entries as the tooling landscape changes.
