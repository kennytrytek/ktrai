# Symbol Index

_Generated: 2026-07-31_

## `cmd/align.go`

### `cmd` (package)
- `alignCmd` (var)
- `init` (func) `()`
- `runAlign` (func) `(_ *cobra.Command, args []string)`
- `resolveRoot` (func) `(args []string)`
- `fileExists` (func) `(path string)`
- `wireToolSymlinks` (func) `(root, agentDir, rulesDir, contextDir string)`
- `migrateExisting` (func) `(root, contextDir, rulesDir string)`
- `printAlignNextSteps` (func) `(root string)`
- `defaultAgentsMD` (const)

## `cmd/gen_symbols.go`

### `cmd` (package)
- `genSymbolsCmd` (var)
- `RegisterGenSymbols` (func) `()`
- `runGenSymbols` (func) `(_ *cobra.Command, _ []string)`
- `formatTopLevel` (func) `(s ctags.Symbol)`
- `formatMember` (func) `(s ctags.Symbol)`

## `cmd/root.go`

### `cmd` (package)
- `rootCmd` (var)
- `version` (var)
- `Execute` (func) `()`

## `internal/agentsmd/agentsmd.go`

### `agentsmd` (package)
- `Content` (struct)
- `Module` (struct)
- `Load` (func) `(path string)`
- `Draft` (func) `(root string, lang detect.Language)`
- `Render` (func) `(c *Content)`
- `parse` (func) `(raw, defaultProjectName string)`
- `scanModules` (func) `(root string, lang detect.Language)`
- `isTestFile` (func) `(rel string, lang detect.Language)`

## `internal/ctags/ctags.go`

### `ctags` (package)
- `Symbol` (struct)
- `skipKinds` (var)
- `Parse` (func) `(r io.Reader)`

## `internal/detect/detect.go`

### `detect` (package)
- `Language` (type)
- `Go` (const)
- `Python` (const)
- `TypeScript` (const)
- `Java` (const)
- `Unknown` (const)
- `markers` (var)
- `Detect` (func) `(root string)`
- `fileExists` (func) `(path string)`

## `internal/makefile/makefile.go`

### `makefile` (package)
- `targetMarker` (const)
- `genContextTarget` (func) `(lang detect.Language)`
- `prepTarget` (const)
- `srcDirFor` (func) `(lang detect.Language)`
- `Inject` (func) `(path string, lang detect.Language)`

## `internal/scaffold/rules.go`

### `scaffold` (package)
- `CodemapRule` (const)
- `UpdateAgentsMdRule` (const)
- `SymbolsPlaceholder` (const)

## `internal/scaffold/scaffold.go`

### `scaffold` (package)
- `EnsureDir` (func) `(dir string)`
- `WriteFile` (func) `(path, content string)`
- `WriteIfNotExists` (func) `(path, content string)`
- `EnsureSymlink` (func) `(link, target string)`
- `IsSymlinkInto` (func) `(path, agentDir string)`
- `MoveToAgent` (func) `(src, destDir string)`
- `copyDir` (func) `(src, dst string)`

## `main.go`

### `main` (package)
- `main` (func) `()`

