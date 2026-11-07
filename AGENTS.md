# AGENTS.md

This file documents non-obvious knowledge needed to work effectively in the
Elvish repository. Elvish is a Go implementation of the Elvish shell and
language. This repo is the `tab-dance/elvish` fork, which adds a few
completion/runtime features on top of upstream Elvish (see
[Fork-specific features](#fork-specific-features)).

## Project overview

-   Module path: `src.elv.sh` (import paths are `src.elv.sh/pkg/...`).

-   Go version: `go 1.22` (CI also tests `1.22.x` and `1.23.x`).

-   Entrypoints live in `cmd/`:

    -   `cmd/elvish` - default shell (interactive + scripting).

    -   `cmd/nodaemon/elvish` - build without storage daemon support.

    -   `cmd/withpprof/elvish` - build with profiling flags.

    -   `cmd/elvmdfmt` - Markdown formatter used by `make fmt`.

-   Most code lives under `pkg/`. `website/` is a separate Go module
    (`src.elv.sh/website`) that replaces `src.elv.sh` with `../`.

-   `e2e/` contains end-to-end transcript tests that build the binary at
    runtime.

## Architecture

High-level layout (see `docs/architecture/doc.go` for the full overview):

-   `pkg/prog` - composable `Program` interface; `prog.Run` parses flags and
    dispatches subprograms.

-   `pkg/shell` - shell subprogram; interactive REPL vs non-interactive script
    modes; initializes storage daemon.

-   `pkg/eval` - interpreter (`eval.Evaler`): parse -> compile operation tree ->
    run. Also implements the `builtin` module functions (`builtin_fn_*.go`),
    values, frames, namespaces, and `Evaler.Context()`.

-   `pkg/parse` - handwritten recursive-descent parser (no separate tokenizer).

-   `pkg/edit` - interactive line editor and completion engine
    (`edit/complete`).

-   `pkg/cli`, `pkg/ui` - lower-level terminal UI and styled text.

-   `pkg/daemon` - storage daemon subprogram and client.

-   `pkg/mods` - standard library modules (`runtime`, `str`, `path`, etc.).

-   `pkg/persistent` - Clojure-like persistent lists/maps.

-   `pkg/store`, `pkg/env`, `pkg/sys`, `pkg/fsutil`, etc. - supporting packages.

Data/control flow for a command: `shell.Program.Run` creates an `eval.Evaler`,
then either runs a script file/code (`script.go`) or starts the interactive REPL
(`interact.go`), which uses `edit.Editor` (built on `cli.App`) to read input and
`eval.Evaler.Eval` to execute it. Completion requests go through
`edit/complete.Complete`.

## Essential commands

Build:

```sh
make get                       # builds to $GOBIN/elvish (override via ELVISH_MAKE_BIN)
go install ./cmd/elvish        # install from local tree
go build -o ./elvish ./cmd/elvish
```

Test:

```sh
make test                      # unit tests (race where supported) + website tests
go test ./...                  # root module only
cd website && go test ./...    # website module
```

Set `ELVISH_TEST_TIME_SCALE=20` (CI uses 10/20) on slow/VM hosts; some tests
depend on time thresholds.

Checks (CI runs these on ubuntu/macos/windows):

```sh
make most-checks               # fmt-go, fmt-md, disallowed, vet, staticcheck, codespell
make all-checks                # most-checks + check-gen (requires clean tree)
```

External tools required by checks:

```sh
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/stringer@latest
go install honnef.co/go/tools/cmd/staticcheck@v0.6.0
pip install codespell==2.3.0
```

Formatting:

```sh
make fmt                       # goimports + gofmt -s + elvmdfmt -w -width 80
go run src.elv.sh/cmd/elvmdfmt -width 80 -w $file   # single markdown file
```

Code generation:

```sh
go generate ./...              # regenerate stringer outputs (needs stringer)
```

Coverage:

```sh
make cover                     # unit + e2e coverage, opens HTML report
```

Website:

```sh
cd website && make gen         # generate site into _dst
cd website && make check-rellinks
```

Pre-push hook (runs tests + all checks + cross-compile):

```sh
ln -s ../../tools/pre-push .git/hooks/pre-push
```

## Important environment/invocation gotchas

-   `GOWORK=off` must be set when running website tests/checks (pre-push does
    this). A `go.work` file may exist for gopls in VS Code and changes `go test`
    behavior.

-   On Windows, turn off `core.autocrlf` before tests/checks (CI does this);
    autocrlf breaks fuzz testdata and `check-fmt-go`.

-   Race detector is only supported on a limited set of `GOOS/GOARCH` (see
    `tools/run-race.elv`). `make test` uses that script to decide whether to
    pass `-race`.

-   `make all-checks` / `tools/check-gen.sh` require a clean working tree (no
    unstaged/untracked files) because they run `go generate` and diff.

-   Elvish transcript tests use `ELVISH_TRANSCRIPT_RUN=file.elvts:lineno` to run
    and update a single interaction (see `pkg/eval/evaltest`).

## Code organization and conventions

-   Import paths use the module name `src.elv.sh`, not the GitHub path.

-   File names use snake_case; generated files are `zstring.go` (stringer) and
    `ztypes_windows.go` (cgo godefs). Do not hand-edit generated files;
    regenerate.

-   Doc comments: exported symbols start with the symbol name; unexported
    symbols do not repeat the name (see `docs/documenting.md`).

-   Markdown is formatted to 80 columns by `cmd/elvmdfmt`; CI enforces this.

-   Go is formatted with `goimports` then `gofmt -s`.

-   Disallowed content (enforced by `tools/check-disallowed.sh`):

    -   no import of `"net/rpc"` (Elvish ships a trimmed copy in `pkg/rpc`);

    -   no `unix.Umask/Getrlimit/Setrlimit` calls outside `pkg/mods/unix`,
        `pkg/daemon`, and `pkg/testutil`;

    -   no `./` prefix in link destinations to `language.html` in `.elv` files.

-   Dependency injection for untestable deps: store the real dependency in a
    package-level var (e.g. `var osSleep = os.Sleep`) and override it in tests
    with `testutil.Set`. For external test packages, export a pointer in a
    `*_test.go` file named like `testexport_test.go`.

-   Tests are often transcript-based (`.elvts` / `.elv`) parsed by
    `evaltest.TestTranscriptsInFS`; embed them with `//go:embed`.

## Testing patterns

-   Prefer testing the real thing; mock only for syscall/timing cases (above).

-   Transcript tests are the standard for module/interpreter behavior; use
    `//in-temp-dir`, `//eval`, `//only-on`, etc. directives.

-   External test packages (`_test`) use `testexport_test.go` pointers for
    injection.

-   Run with race detection where the platform supports it
    (`tools/run-race.elv`).

-   Some tests depend on `ELVISH_TEST_TIME_SCALE`.

## Fork-specific features

This fork (branch `tabdance`, default remote) adds:

-   **Command completer**: `$edit:completion:command-completer` (a function) is
    called with the seed (partial command name) and should output candidates;
    when unset or it returns nothing, built-in command completion is used. See
    `adaptCommandGenerator` in `pkg/edit/completion.go` and
    `complete.generateCommandCandidates`.

-   **Variable completer**: `$edit:completion:variable-completer` is called with
    seed and namespace; overrides built-in variable enumeration. See
    `adaptVariableGenerator`.

-   **Runtime context**: `eval.Evaler.Context()` returns a `ShellContext`
    (aliases, builtins, functions, jobs, variables); exposed as
    `runtime:context` for external tools like carapace. See
    `pkg/eval/context.go` and `pkg/mods/runtime`.

-   Smart-completion prefix handling was adjusted for quotation changes
    (`pkg/edit`).

-   CI/release uses GitHub Actions + GoReleaser (`.goreleaser.yml`); upstream
    Cirrus-based infra was removed from CI.

## Release/packaging notes

-   Main package for packaging is `cmd/elvish`.

-   Version info can be injected via linker flags into `pkg/buildinfo` (see
    `docs/packaging.md`).

-   Cross-compilation smoke test is in `tools/pre-push`.

-   Do not push, open PRs, or create releases unless explicitly commanded.

## Where to look first

-   Commands/flags: `pkg/shell/shell.go`, `pkg/prog`.

-   Language semantics: `pkg/eval/compile_*.go`, `pkg/eval/builtin_*.go`.

-   Parser: `pkg/parse`.

-   Editor/completion: `pkg/edit`, `pkg/edit/complete`.

-   Modules: `pkg/mods`.

-   Storage daemon: `pkg/daemon`.
