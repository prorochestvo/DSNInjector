# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

This is a pure-Go library (single package `dsninjector` at the module root) with no
binaries and no CGO. All commands run with `CGO_ENABLED=0`.

```bash
CGO_ENABLED=0 go build ./...                 # compile the package
CGO_ENABLED=0 go test -race ./...            # run the full test suite
CGO_ENABLED=0 go vet ./...                   # static checks
gofmt -l .                                   # list files needing formatting (empty == clean)
go fmt ./...                                 # format in place
go mod tidy                                  # reconcile go.mod / go.sum
```

Targeted test runs (the package lives at the module root, so the path is `.`):
```bash
# Single top-level test
CGO_ENABLED=0 go test -race -run TestParse .

# Single subtest
CGO_ENABLED=0 go test -race -run 'TestParse/successful' .

# Verbose output (see every subtest pass/fail)
CGO_ENABLED=0 go test -race -v .

# Benchmarks
CGO_ENABLED=0 go test -bench=. -benchmem -run=^$ .

# Coverage
CGO_ENABLED=0 go test -race -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

## Architecture Overview

`dsninjector` is a small, dependency-free library for parsing and emitting Data Source
Names (DSNs) / connection strings. `Parse` turns a URL-shaped DSN
(`driver://login:password@host:port/database?opt=val`) into a `DataSourceMapper` — a
`map[string]string` that exposes typed accessors (`Driver`, `Host`, `Port`, `Login`,
`Password`, `Database`) plus arbitrary query options via `Option`/`OptionsNames`.
`Marshal` does the reverse, rebuilding the DSN string from a `DataSource`. Convenience
helpers read DSNs straight from the environment (`Unmarshal`, `UnmarshalOrEmpty`) and
load `.env`-style files into the process environment (`InitEnvFrom`). There is no
service, no I/O beyond optional env-file reading, and no long-running process.

### File Responsibilities

| File | Role |
|------|------|
| `parser.go` | Free functions: `Parse` (regex-driven DSN → `DataSourceMapper`), `Marshal` (`DataSource` → DSN string), `Unmarshal`/`UnmarshalOrEmpty` (read+parse a DSN from an env var), `InitEnvFrom` (load `.env`-style files via the unexported `readEnvVarName` scanner). |
| `source.go` | The data model: the `DataSource` interface and its `DataSourceMapper` (`map[string]string`) implementation — typed getters/setters, `Addr`, `AuthBasicBase64`, the `Option`/`OptionsNames` option bag, and the `keyDriverName`/`keyHostName`/… field-key constants. |

### Key Patterns

- **Map-backed value object** — `DataSourceMapper` is a `map[string]string` behind the
  `DataSource` interface. The six well-known fields are stored under the reserved
  `key*Name` constants (`driver`, `hostname`, `port`, `login`, `password`, `database`);
  every other key is treated as a free-form connection option. `OptionsNames` returns
  exactly the keys that are *not* reserved, which is how `Marshal` reconstructs the
  `?a=b&c=d` query tail.
- **Regex-driven parsing** — `Parse` uses a single named-group regular expression to
  split the DSN, then post-processes `instance` into host/port and `credentials` into
  login/password (a credentials string with no `:` is treated as a password-only value).
  Query parameters are decoded with `net/url` and flattened into the same map.
- **Typed accessors over a string map** — getters coerce on read (`Port` does
  `strconv.ParseInt` and swallows the error, returning `0` on a bad value); `Addr` and
  `AuthBasicBase64` compose stored fields on demand rather than caching.
- **Env / file convenience layer** — `Unmarshal` reads a single env var (with optional
  default) and delegates to `Parse`; `UnmarshalOrEmpty` never errors, falling back to an
  empty `DataSourceMapper`. `InitEnvFrom` reads `KEY=VALUE` lines (default file `.env`),
  upper-cases keys, and calls `os.Setenv`; missing files are skipped silently.

### Key Dependencies

Runtime: **standard library only** — `bufio`, `encoding/base64`, `errors`, `fmt`, `io`,
`net/url`, `os`, `path`, `regexp`, `strconv`, `strings`. No third-party runtime imports.

Test-only:
- `github.com/stretchr/testify` — assertions in `*_test.go` (pulls in `go-spew`,
  `go-difflib`, `yaml.v3` as indirect deps).

Go version: `1.23.3` (per `go.mod`).

## Code Organization Principles

These rules govern *where code lives*. Apply them by default; treat a violation as
something to flag, not silently accept.

### Package placement follows consumption, not aspiration

- Code shared by **multiple** binaries/entry points belongs in the shared tree
  (`internal/`).
- Code with **exactly one** consumer belongs **next to that consumer**
  (`cmd/<binary>/`), not in the shared tree.
- Prefer the private location (`internal/`) over a public one (`pkg/`) unless there
  is a **real external (out-of-module) consumer**. Don't promise a public API
  surface the project doesn't actually provide.
- **Why:** the shared tree is for the genuinely shared layers of one app. Putting
  single-consumer code (or a separate app) there bloats it and implies a contract
  that doesn't exist; a `pkg/` package nobody outside the module imports is dead
  weight. Before placing or keeping a package in the shared or public tree, check
  who actually imports it — one consumer means co-locate, no external module means
  keep it private. Never keep something in the shared tree just because it's
  "reusable in principle"; treat such a move as its own deliberate refactor.

### Deduplication is not a goal in itself

- Distinguish **coincidental similarity** (looks alike today but must be free to
  diverge) from a **genuine cross-cutting invariant**. Coincidental similarity →
  duplicate the few trivial lines and let each site evolve. A true invariant →
  centralize it once, where it belongs.
- Do **not** build a shared `bootstrap` / `startup` / `wiring` layer for multiple
  binaries just because their startup looks similar — inline it per entry point
  (`cmd/<binary>/main.go`) so each stays free to diverge (different DBs,
  dependencies, config).
- Before extracting a helper, check whether the only thing being shared is already
  captured elsewhere (e.g. already a one-line call) — if so, don't wrap it. An
  abstraction can re-introduce the very complexity it pretends to hide (e.g. a
  returning constructor needs error-cleanup that an inline fatal-and-exit path
  simply doesn't).
- **Why:** premature extraction imposes a contract where code should diverge. Dedup
  earns its place only when it names a non-obvious invariant, removes a real
  divergence risk, or cuts genuine cognitive load — not because two snippets look
  alike.

### Business logic is organized by concern, not by launcher

- Business-logic packages are judged by being **simple and isolated**, regardless of
  which binary runs them or how they are launched ("how it starts is not the
  package's concern"). Keep a flat, per-concern split.
- Do **not** reorganize business logic by runtime-vs-operator, by deployment, or by
  consuming binary.
- **Why:** grouping by launcher couples organization to deployment, which changes;
  cohesion by concern is stabler. Isolation + simplicity is the real quality bar.

## File Declaration Order

Order the top-level declarations in each `*.go` file so the important, public surface
is at the top and private internals are hidden at the bottom. A reader should see
everything important first; scanning the file should not require digging.

For a file built around one object:

1. Exported `const` and `var`, plus the `New<Object>` constructor(s). These come first
   because they are what you need to create and use the object — the first thing a
   reader looks for.
2. The object's struct definition.
3. The object's methods (prefer alphabetical order; not mandatory).
4. Unexported `const` and `var`.
5. Auxiliary/helper structs (unexported support types) — placed between the unexported
   vars/consts and the unexported methods.
6. Unexported methods/functions (prefer alphabetical order; not mandatory).

- **Multiple structs in one file:** keep the same layout but put the primary ("main")
  struct first. A combined layout is acceptable but very rare — two large objects in one
  file usually means the file should be split into two.
- **Files with no object** (free functions plus a config/data type): apply the same
  spirit — exported type(s) and function(s) on top, then unexported consts/vars, then
  auxiliary structs, then unexported helper functions.

Treat a file that violates this order as something to fix.

## Error Handling

There is no custom error type. Exported functions return plain `error` values:

- `Parse` returns `fmt.Errorf("invalid connection string format: %s", dns)` when the
  regex does not match, and propagates the underlying `url.ParseQuery`/`regexp.Compile`
  error otherwise.
- `Unmarshal` returns `fmt.Errorf("could not recognize configuration")` for an
  empty/whitespace value and otherwise forwards the error from `Parse` unchanged.
- `InitEnvFrom` and `readEnvVarName` wrap I/O failures with `%w` so callers can
  `errors.Is`/`errors.As` the cause (e.g. `could not extract environment variables from
  %s, reason: %w`), and `readEnvVarName` joins the deferred `Close` error via
  `errors.Join`.
- `UnmarshalOrEmpty` deliberately swallows the error and returns an empty
  `DataSourceMapper` — it is the "never fail" convenience entry point.

When adding code, keep this contract: wrap propagated failures with `%w`, return plain
sentinel-free errors for bad input, and reserve error-swallowing for the explicitly
"OrEmpty"/best-effort helpers.

## Constraints

- **Zero runtime dependencies**: this is a standard-library-only library. Keep it that
  way — `go.mod`'s only non-indirect `require` is `testify`, and it must stay
  **test-only**. Do not introduce a third-party runtime import without an explicit
  decision; nothing CGO-dependent may enter the build.
- **Testing**: Use `github.com/stretchr/testify`; run tests with `-race`; parallel
  subtests preferred where there's no shared mutable state.
- **One `Test*` per method, scenarios as subtests**: each tested method/function gets
  exactly one top-level test function named after it (e.g. `TestEncode` for `Encode`),
  and every scenario for that method lives as a `t.Run("descriptive name", ...)`
  subtest inside it. Do **not** create separate top-level tests like
  `TestEncode_EmptyInput`, `TestEncode_Unicode`, `TestEncode_Error` — these belong
  as subtests of a single `TestEncode`. Methods on a type follow the same rule with
  the standard `TestType_Method` form (e.g. `TestUser_Validate`).
  ```go
  func TestEncode(t *testing.T) {
      t.Parallel()

      t.Run("empty input returns empty string", func(t *testing.T) {
          t.Parallel()
          // ...
      })

      t.Run("unicode is preserved", func(t *testing.T) {
          t.Parallel()
          // ...
      })

      t.Run("returns error on invalid byte", func(t *testing.T) {
          t.Parallel()
          // ...
      })
  }
  ```
- **No CGO**: `CGO_ENABLED=0` must be set for all build and test commands (unless the
  project intentionally requires CGO).
- **Compile-time interface checks**: Any type that implements an interface (notably
  `DataSourceMapper` against `DataSource`), and every mock/stub struct in test files,
  must have a `var _ DataSource = &DataSourceMapper{}` style assertion so the
  implementation is enforced at compile time.
- **No section-divider comments**: Do not use `// --- section ---` or `// ----` style
  separator comments. Let the code structure speak for itself.
- **No skipped errors**: Never use `_` to discard error return values in production or
  test code. Always capture the error and assert/check it. The only exceptions are
  `fmt.Fprint*` writes to loggers, `Rollback()` calls in error-recovery paths, and
  resource `.Close()` in `t.Cleanup` / `defer`.
- **Comments**: all comments are in English and start with a lowercase first word
  (e.g. `// wrap the driver error so callers can match on it`).
- **Godoc on exported identifiers**: Every exported identifier (Type, Func, Method,
  Var, Const) gets a doc comment that starts with the identifier name and ends with
  a period — e.g. `// Encode returns the base64-encoded form of v.` Each package
  has exactly one `// Package <name> ...` declaration; `cmd/*` entry points use
  `// Command <name> ...` instead. Skip the comment entirely if it would only
  restate the signature — no `// Foo is a Foo.` fluff. Document concurrency
  guarantees, the error contract (which functions wrap with `%w` vs swallow errors),
  and any non-obvious coercion behaviour (e.g. `Port` returning `0` on a bad value).
  Preserve existing WHY-comments verbatim; do not overwrite a substantive comment
  with a generic restatement. Unexported symbols only get comments when intent is
  non-obvious — do not bulk-add comments to private helpers.
- **Scratch artifacts stay out of the repo root**: this library ships no binaries, so
  `go build` produces nothing to place. Any throwaway artifacts, fixtures, coverage
  profiles, or intermediate files belong in `./tmp/` (not the repo root, where
  `git add .` would pick them up). Keep `cover.out` and similar out of commits.

## Planning Workflow

All non-trivial work is tracked as a Markdown plan file before implementation begins.

### Directory layout

```
plans/
├── NNN-task-slug.md     # active / in-progress plans (e.g. 001-fix-auth.md)
├── completed/           # plans for fully shipped tasks (e.g. 260422.0001.fix-auth.md)
└── history/             # archived / cancelled plans
```

### File naming

- **Active plans (`plans/`)** — zero-padded sequential index + kebab-case slug:
  `NNN-description.md` (e.g. `001-fix-unauthorized-middleware.md`, `002-add-rate-limiting.md`).
  Pick the next number by checking the highest existing prefix across `plans/`, `plans/completed/`,
  and `plans/history/`.

- **Completed plans (`plans/completed/`)** — date prefix + zero-padded daily index (4 digits) + slug:
  `YYMMDD.NNNN.description.md` (e.g. `260422.0001.fix-unauthorized-middleware.md`).
  `NNNN` resets to `0001` each day and increments for each additional completion on that day.

- **Archived plans (`plans/history/`)** — keep the original `NNN-` filename from `plans/`.

### Lifecycle

1. **Create** — before touching code, produce a plan file in `plans/` using the `NNN-slug.md`
   naming convention described above.
2. **Implement** — work through the tasks defined in the plan. The plan file stays in
   `plans/` while work is in progress.
3. **Complete** — once every acceptance criterion is met and `make test` passes, rename
   and move the file to `plans/completed/` using the date-based convention:
   ```bash
   mv plans/001-fix-auth.md plans/completed/260422.0001.fix-auth.md
   ```
4. **Archive** — if a plan is abandoned or superseded without being fully implemented or
   if we need to save intermediate data or task execution logs, move it to `plans/history/` instead.

### Plan file format

Every plan file follows this structure:

```markdown
# Task Breakdown

## Overview
## Assumptions
## Tasks
### Task N: <Title>
- Description:
- Acceptance Criteria:
- Pitfalls & edge cases:
- Complexity: Easy / Medium / Hard
## Execution Order
## Risks
## Trade-offs
```

### Rules

- **One plan per concern.** Don't bundle unrelated changes in a single plan file.
- **Plan before code.** Claude must create (or confirm an existing) plan file before
  writing or modifying any source files.
- **Keep plans honest.** If implementation diverges from the plan, update the plan file
  before moving it to `completed/`.
- **Slug matches intent.** The description part of the filename should be readable at a glance:
  `002-add-rate-limiting.md`, `003-migrate-sqlite-to-postgres.md`, not `004-task.md`.

## Agent Pipeline

All non-trivial tasks follow a three-stage pipeline using specialized agents. The
review stage fans out to **three `gocode-reviewer` instances running in parallel**,
each with a distinct lens. A separate `gocode-testdoctor` agent is invoked
on-demand whenever tests fail, at any stage.

```
User describes task
    ↓
1. gocode-architect
    → Creates plan file at plans/NNN-slug.md (see Planning Workflow)
    ↓
2. gocode-engineer
    → Implements the tasks defined in the plan
    ↓
3. gocode-reviewer × 3 (run in parallel — single message, three tool calls)
    Lens A: correctness & tests — bugs, races, edge cases, error paths,
            context propagation, resource cleanup, test coverage,
            test structure (one Test* per method with subtests),
            scenario completeness, fixtures
    Lens B: security & operations — input validation, auth boundaries,
            secrets handling, injection (SQL, command, template),
            observability (logs, metrics, traces), log volume,
            operator/runbook UX
    Lens C: performance & architecture — allocations, blocking I/O,
            goroutine/resource leaks, layer boundaries, dependency
            direction, API contracts (breaking changes, exported
            surface stability), interface scope, future-proofing
    ↓
   Orchestrator synthesises all three reports, deduplicates findings,
   resolves conflicts (e.g. one reviewer flags as P0 what another
   accepts as a trade-off), and presents the merged punch list to the user.
    ↓
  ❌ P0/P1 found?  → Back to gocode-engineer with the consolidated findings.
                             After fix, run ONE targeted reviewer pass on the changed
                             lines (not all 3 again) before re-approval.
  ⚠️  Tests failing?        → gocode-testdoctor diagnoses and patches, then rerun the
                             targeted reviewer pass.
  ✅ All three approve?     → Orchestrator moves the plan: mv plans/NNN-slug.md
                             plans/completed/YYMMDD.NNNN.slug.md
```

### Agent responsibilities

| Agent | Owns | Output |
|-------|------|--------|
| `gocode-architect` | Planning, decomposition, trade-offs | New plan file in `plans/` |
| `gocode-engineer` | Implementation, tests for new code | Code + tests in the repo |
| `gocode-reviewer` (×3, parallel) | Lens-specific verdicts, priority-ranked findings, patch sketches | Three independent review reports |
| `gocode-testdoctor` | Triage of failing tests, minimal patches | Code/test fixes, re-run of `make test` |

The orchestrating agent (the main Claude session driving the pipeline) owns
synthesis: merging the three reports, resolving conflicting verdicts, deciding
which findings to act on, and moving the plan to `completed/` once everyone
signs off.

Priority scale used by reviewers: **P0 / P1 / P2 / P3**.

### Rules

- **No skipping stages.** Every task starts with the architect and ends with the three-reviewer fan-out.
- **Plan file first.** The architect MUST produce a plan file before any code is written. If a plan already exists for the task, update it rather than creating a new one.
- **Three reviewers, three lenses, one message.** All three `gocode-reviewer` agents are launched in a single tool-call batch (multiple `Agent` blocks in one message) so they run in parallel. Each prompt names the lens explicitly and tells the agent what to SKIP (the other lenses) to avoid duplicated work.
- **No solo reviewer pass on first review.** Even for small changes the full three-lens fan-out is required, because the lenses catch genuinely different classes of issue (Lens A won't see ops/log-volume problems; Lens C won't see test gaps). Skipping lenses is what the orchestrator does AFTER a P0/P1 fix, not BEFORE the first verdict.
- **Lens prompts are self-contained.** Each reviewer's prompt must include: (1) the lens name, (2) what to focus on, (3) what to SKIP (so it doesn't restate other lenses), (4) the file list, (5) the deliverable shape (P0 / P1 / P2 / P3 with `file:line` + patch sketch), (6) the word cap (typically 600 words).
- **Re-review after fixes is single-pass.** Once an engineer addresses P0/P1 findings, the orchestrator runs ONE reviewer pass scoped to the changed lines, not the full fan-out. Re-running all three each iteration is expensive and rediscovers nothing.
- **Conflict resolution is explicit.** When reviewers disagree (one says P0, another says trade-off), the orchestrator chooses, names the rejected suggestion, and explains the reasoning to the user before moving on. The user has final say.
- **Orchestrator gates completion.** The plan moves to `plans/completed/` only after every reviewer's P0 and P1 findings are addressed (either fixed, or explicitly accepted with rationale). The rename uses the standard `YYMMDD.NNNN.slug.md` format.
- **`make test` must pass** before review begins. If it fails, hand the logs to `gocode-testdoctor` first — reviewers should not waste time on a red tree.
- **Testdoctor is scoped.** It patches tests or the minimal production code needed to make the failure go away. It does not redesign or refactor.
