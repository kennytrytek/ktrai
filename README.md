# ktrai

`ktrai` installs a tool-agnostic AI agent context layer into any software repository. It creates a canonical `.agent/` directory that Cursor, Claude, and other AI coding tools share — no more duplicated rule files, no more per-tool config sprawl.

## What ktrai creates

```
your-repo/
├── .agent/
│   ├── context/
│   │   ├── AGENTS.md            ← repo overview, conventions, module map (edit this)
│   │   └── symbols.md           ← ctags-generated symbol index
│   ├── rules/
│   │   └── update-agents-md.md  ← instructs agents when and how to keep AGENTS.md current
│   └── skills/                  ← on-demand agent procedure library (add SKILL.md files here)
│       └── <skill-name>/
│           └── SKILL.md
│
├── AGENTS.md                    → symlink → .agent/context/AGENTS.md
├── CLAUDE.md                    → symlink → .agent/context/AGENTS.md
├── .skills                      → symlink → .agent/skills
├── .cursor/
│   ├── rules                    → symlink → ../.agent/rules
│   └── skills                   → symlink → ../.agent/skills
└── .claude/
    ├── rules                    → symlink → ../.agent/rules
    └── skills                   → symlink → ../.agent/skills
```

All symlinks use relative paths for portability across machines. Every operation is idempotent — re-running `ktrai align` on an already-aligned repo is safe.

## Install

```sh
brew tap kennytrytek/tap
brew install ktrai
```

## Usage

### `ktrai align [directory]`

Set up or migrate a repository into the `.agent/` layout. If `directory` is omitted, the current working directory is used.

```sh
ktrai align          # align the current directory
ktrai align ~/work/my-repo
```

What `align` does:

1. Creates `.agent/context/`, `.agent/rules/`, and `.agent/skills/` if absent.
2. Migrates any real (non-symlink) `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/`, and `.claude/rules/` into `.agent/`, replacing them with symlinks.
3. Migrates skill subdirectories from `.skills/`, `.claude/skills/`, and `.cursor/skills/` into `.agent/skills/`, replacing each source with a symlink. Duplicate skill names across sources cause an error — rename or merge the conflicting skills and re-run.
4. Removes configuration files for unsupported AI tools (e.g. `.github/copilot-instructions.md`).
5. Creates a skeleton `AGENTS.md` under `.agent/context/` if none exists.
6. Drops the `update-agents-md` rule into `.agent/rules/` if absent.
7. Ensures `AGENTS.md`, `CLAUDE.md`, `.cursor/rules`, `.cursor/skills`, `.claude/rules`, `.claude/skills`, and `.skills` are symlinks pointing into `.agent/`.

After running `align`:
- Open `.agent/context/AGENTS.md` and fill it in — the skeleton has TODOs for the project purpose, module descriptions, conventions, and commit pre-flight commands.
- Add shared skills to `.agent/skills/` — each skill is a subdirectory containing a `SKILL.md`.

See the `update-agents-md` rule in `.agent/rules/` for detailed instructions on keeping `AGENTS.md` current.

### Skills vs rules vs commands

| Concept | Location | Description |
|---|---|---|
| **Rules** | `.agent/rules/` | Always-on context injected into the agent automatically (conventions, style, commit steps) |
| **Skills** | `.agent/skills/` | On-demand procedures the agent follows when explicitly invoked (e.g. `ci-review`, `deploy-service`) |
| **Commands** | `.cursor/commands/` | Cursor-only slash commands — not managed by ktrai; no Claude Code equivalent |


## Makefile targets

| Target | Description |
|---|---|
| `make build` | Compile `./bin/ktrai` with the current git version baked in |
| `make run` | Build and run `ktrai align` against this repo |

## Development

Requires **Go 1.24+**.

```sh
git clone git@github.com:kennytrytek/ktrai.git
cd ktrai
make build          # compile to ./bin/ktrai
go vet ./...        # must pass before committing
go build ./...      # must compile cleanly
```

CI (GitHub Actions) runs `go build ./...`, `go vet ./...`, and `go test ./...` on every push and pull request to `main`.

## Release

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. Cross-platform binaries (macOS and Linux, amd64 and arm64) are built and attached to the GitHub release; the Homebrew formula in `kennytrytek/homebrew-tap` is updated automatically.

Required repository secrets:

- `GITHUB_TOKEN` — provided automatically by Actions
- `HOMEBREW_TAP_TOKEN` — a GitHub PAT with `repo` scope on `kennytrytek/homebrew-tap`

To cut a release:

```sh
git tag v0.2.0
git push origin v0.2.0
```

## License

MIT — see [LICENSE](LICENSE).
