# ktrai

`ktrai` installs a tool-agnostic AI agent context layer into any Go, Python, TypeScript, or Java/Kotlin repository. It creates a canonical `.agent/` directory that Cursor, Claude, and other AI coding tools share — no more duplicated rule files, no more per-tool config sprawl.

## What ktrai creates

```
your-repo/
├── .agent/
│   ├── context/
│   │   ├── AGENTS.md        ← repo overview, conventions, module map
│   │   └── symbols.md       ← ctags-generated symbol index (if universal-ctags is installed)
│   └── rules/
│       ├── codebase-map.md  ← instructs agents to read the symbol index
│       └── update-agents-md.md ← instructs agents to keep AGENTS.md current
│
├── AGENTS.md                → symlink → .agent/context/AGENTS.md
├── CLAUDE.md                → symlink → .agent/context/AGENTS.md
├── .cursor/
│   └── rules                → symlink → ../.agent/rules
└── .claude/
    └── rules                → symlink → ../.agent/rules
```

All symlinks use relative paths for portability across machines. Every operation is idempotent — re-running `ktrai init` on an already-initialized repo is safe.

## Install

```sh
brew tap kennytrytek/tap
brew install ktrai
```

`universal-ctags` is an optional but recommended dependency. Without it, `ktrai gen` will not run and a placeholder `symbols.md` is written instead.

```sh
brew install universal-ctags
```

## Usage

### `ktrai init`

Scaffold the full `.agent/` context layer in the current directory.

```sh
ktrai init
ktrai init --language go      # override language detection
```

Detection reads `go.mod`, `pyproject.toml` / `requirements.txt`, `package.json`, and `build.gradle` / `settings.gradle` in that order.

### `ktrai restructure`

Migrate existing `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/`, and `.claude/rules/` into `.agent/`, then run `init` for anything still missing.

```sh
ktrai restructure
ktrai restructure --language typescript
```

### `ktrai gen`

Read `universal-ctags --output-format=json` from stdin and write a Markdown symbol table to stdout.

```sh
ctags --output-format=json --fields='*' -R . | ktrai gen > .agent/context/symbols.md
```

This is also injected as a `gen` Make target when `ktrai init` detects a Makefile.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--language` | auto-detected | Force a specific language (`go`, `python`, `typescript`, `java`) |
| `--version` | — | Print version and exit |

## Directory layout (detail)

```
.agent/
├── context/
│   ├── AGENTS.md          Repo overview for AI agents (editable — this is your source of truth)
│   └── symbols.md         Auto-generated symbol index; regenerate with `make gen`
└── rules/
    ├── codebase-map.md    Rule: consult symbols.md before editing unfamiliar code
    └── update-agents-md.md Rule: keep AGENTS.md accurate after significant changes
```

## Makefile integration

When `ktrai init` finds a `Makefile` in the repo root it injects two targets (idempotently):

```make
gen:
	ctags --output-format=json --fields='*' -R . | ktrai gen > .agent/context/symbols.md

prep: gen
```

## Development

```sh
git clone git@github.com:kennytrytek/ktrai.git
cd ktrai
go build ./...
go test ./...
go vet ./...
```

Requires **Go 1.24+**.

## Release

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions.

1. Ensure the following secrets are set on the `kennytrytek/ktrai` GitHub repo:
   - `GITHUB_TOKEN` — provided automatically by Actions
   - `HOMEBREW_TAP_TOKEN` — a GitHub PAT with `repo` scope on `kennytrytek/homebrew-tap`

2. Tag and push:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds cross-platform binaries, creates a GitHub release, and updates the Homebrew formula in `kennytrytek/homebrew-tap` automatically.

## License

MIT — see [LICENSE](LICENSE).
