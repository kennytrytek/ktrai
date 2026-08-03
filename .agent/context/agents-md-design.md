# AGENTS.md generation — design decisions and research findings

This document captures design decisions for how ktrai should generate `AGENTS.md` files. It
summarises research performed across ~70 real repositories so future agents do not need to repeat
the analysis.

## The core problem

The current `scanModules` implementation produces one table row per source file.  This breaks down
at scale:

| Files | Result |
|---|---|
| < 15 | Fine — file-per-row is readable |
| 15 – 500 | Verbose and fragile — files change, package contracts don't |
| 500+ | Completely unusable (bigsky: 5,286 py files → 5,286 rows) |

The deepest failure: even correctly excluding `node_modules` and `vendor`, a single large
service (`cerberus/` in skaardb) contains 7,914 non-test Go files.  A file listing tells an
agent nothing about _where to look_.

## The right unit of analysis: conceptual concerns, not files

An AGENTS.md Modules/Structure section should answer:
> "Given a task, which directory should I look in first?"

A row should represent a **stable architectural concern** — a package, subsystem, or service — not
an individual file.  Package contracts are stable; file listings change on every PR.

### Depth-calibration heuristic

Walk the directory tree and count source-containing directories.  Pick the coarsest depth that
keeps the table ≤ 20 rows:

1. Count top-level (`depth=1`) directories that contain source files.
2. If ≤ 20, use depth-1 entries.
3. If > 20, use the parent directory with a parenthetical summary of notable sub-packages.
4. Directories with zero source files but meaningful non-source content (`docs/`, `helm/`,
   `idl/`, `static/`, `templates/`, `content/`) always get one row regardless of file count.

### The "dominant package" rule

When one directory contains > 40% of the repo's source files (e.g. `cerberus/internal/common/`
in skaardb with 3,987 / ~8,000 files), its role description should include a parenthetical
enumeration of its key sub-packages:

```
| `cerberus/internal/common/` | Domain logic — documents, tables, linking, permissions, revisions, … |
```

## Project archetypes and their ideal Structure sections

Research was performed across these repos (in `~/git/`):

| Archetype | Examples | Src files (non-vendor) | Strategy |
|---|---|---|---|
| Nano CLI/library | `ktrai`, `wo11y-go`, `platops-ai-discovery` | < 15 | Enumerate packages (directories) |
| Small service | `opentelemetry-collector`, `wdesk_sdk` | 15 – 50 | Enumerate packages |
| Medium app | `pointing-poker`, `infrastructure-provisioning-system`, `dd-documentation` | 50 – 300 | Directory-level grouping |
| Large single-service | `rmconsole-gae` | 300 – 1 000 | Subsystem grouping |
| Huge monolith | `bigsky` (5 286 py), `skaardb` (12 718 go) | 1 000+ | Coarse architectural domains only |
| Multi-service monorepo | `mcp_tools` (servers/* each 10 – 50 files) | varies | One row per service/server |
| Config / infra-only | `k8s-argo-rollouts`, `k8s-karpenter` | ~0 | Describe config concerns (Helm, CF, …) |
| Docs-heavy | `dd-documentation` (19 k content files, 73 src files) | large content, tiny src | Describe content topology sections |

### Worked examples

**Nano — ktrai**
```markdown
| Path | Role |
|---|---|
| `cmd/` | CLI subcommands (align, gen-symbols) |
| `internal/agentsmd/` | Parse and render AGENTS.md |
| `internal/scaffold/` | Idempotent filesystem scaffolding |
| `internal/detect/` | Language and project type detection |
| `internal/ctags/` | ctags-based symbol index generation |
| `internal/makefile/` | Makefile target injection |
| `main.go` | Binary entry point |
```

**Huge Go monolith — skaardb (12 718 Go files)**
```markdown
| Path | Role |
|---|---|
| `cerberus/internal/frontend/` | HTTP/gRPC API layer — REST handlers, Connect dispatch, middleware |
| `cerberus/internal/backend/` | Gesture handlers, manager workers, backend messaging |
| `cerberus/internal/common/` | Domain logic — 100+ packages (documents, tables, linking, permissions, …) |
| `cerberus/internal/warp/` | Distributed coordination and consistency |
| `cerberus/internal/openapi/` | OpenAPI spec serving and validation |
| `cerberus/support/` | Operator tooling — data viewer, migration runner, export CLI |
| `transpiled/` | Cross-language shared library: formatting, model types, wurls, parsing |
| `generators/` | Code generation tools |
| `loadtest/` | Load testing harness |
| `idl/` | Frugal IDL and Protobuf schemas |
| `helm/` | Kubernetes Helm charts (backend, batch, cron, content-api, …) |
| `tool/` | Developer tooling (mockgen, protoc plugin, linter) |
```

**Multi-service monorepo — mcp_tools**
```markdown
| Path | Role |
|---|---|
| `servers/atlassian/` | MCP server: Jira and Confluence tools |
| `servers/datadog/` | MCP server: Datadog metrics, spans, monitors, dashboards |
| `servers/skynet/` | MCP server: Skynet deployment tools |
| `servers/workiva-platform/` | MCP server: Workiva platform tools |
| `servers/eng-ops-tools/` | MCP server: engineering operations utilities |
| `clients/workiva/` | Shared Workiva API client library |
```

**Config-only — k8s-argo-rollouts (0 source files)**
```markdown
| Path | Role |
|---|---|
| `helm/` | Helm chart for Argo Rollouts controller and dashboard |
| `values/` | Per-environment value overrides |
| `templates/` | Kubernetes manifest templates |
| `patches/` | Kustomize patches on upstream |
| `tests/` | Helm chart render tests |
```

**Docs-heavy — dd-documentation (19 k content files, 73 src files)**
```markdown
| Path | Role |
|---|---|
| `content/en/` | English documentation source — Hugo Markdown |
| `content/{fr,es,ja,ko}/` | Localized content mirrors |
| `layouts/` | Hugo templates, shortcodes, and partials |
| `assets/` | CSS, JS, images processed by the build pipeline |
| `data/` | Structured data consumed by templates |
| `config/` | Hugo site and environment configuration |
| `scripts/` | Build, lint, translation, and CI helper scripts |
```

## Sections design: more than one section for larger projects

The best hand-crafted AGENTS.md files observed (pointing-poker, opentelemetry-collector, skaardb)
**do not have a "Modules" section**.  Instead they use domain-specific sections that each answer
a different question.

| Question agents need answered | Section name | Auto-generatable by ktrai? |
|---|---|---|
| Where does code live? | **Structure** | Yes — directory walk |
| What commands do I run? | **Commands** | Yes — from Makefile |
| What non-standard patterns must I follow? | **Patterns** | No — stub only |
| What must I never change? | **Constraints** | No — stub only |
| What do the abbreviations mean? | **Glossary** | No — stub only |
| What are the coding conventions? | **Conventions** | Partial — stub with defaults |

### Tier thresholds for auto-scaffolded sections

| Tier | Source file count | Sections scaffolded |
|---|---|---|
| Nano | < 15 | Structure + Conventions |
| Small | 15 – 100 | Structure + Commands + Conventions |
| Medium | 100 – 500 | Structure + Commands + Patterns (stub) + Conventions |
| Large / complex | 500+ or monorepo | Structure + Commands + Architecture (stub) + Patterns (stub) + Conventions |

`Architecture` is a stub table (e.g. a two-column tier table) that ktrai leaves for a human or
`ktrai align` to fill in — it signals that this section belongs here without inventing content.

`Patterns` stub text: `TODO: document non-standard idioms an agent must follow (error handling,
goroutine wrapping, mock generation, etc.).`

## Key findings about existing good AGENTS.md files

- `pointing-poker/AGENTS.md` — best example of constraint-driven AGENTS.md: "Absolute
  constraints" table of things that must not change + architecture patterns.  No file listing.
- `opentelemetry-collector/AGENTS.md` — best example of contract-driven AGENTS.md: each
  component type (processor, receiver, exporter) has its own file-layout and pattern subsection.
- `skaardb/AGENTS.md` — best example of glossary-driven AGENTS.md: abbreviation table is the
  most valuable thing in the file for an agent navigating 12 000+ files.

## What `scanModules` should NOT do

- Enumerate individual files for projects with > 15 source files
- Walk into `vendor/`, `node_modules/`, `.venv/`, `testdata/`, `__pycache__/` (already correct)
- Produce > 20–25 rows under any circumstances
- Use the same depth for both a 5-file project and a 5000-file project

## Important directory-name heuristics

These directory names always get special treatment regardless of file count:

| Directory name | Always a single row | Notes |
|---|---|---|
| `docs/`, `documentation/` | Yes | "Architecture decisions and design documents" |
| `content/` | Yes | Describe topology (e.g. `content/en/`) not individual files |
| `helm/` | Yes | Enumerate chart names in parentheses if > 1 chart |
| `static/`, `assets/` | Yes | "Static assets and build pipeline inputs" |
| `templates/` | Yes | "Server-side templates / Kubernetes manifests" |
| `idl/`, `proto/` | Yes | "Interface definitions" |
| `scripts/` | Yes | "Build and CI scripts" |
| `testutils/`, `testdata/` | Skip | Test infrastructure, not conceptual modules |
