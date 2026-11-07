# Elvish (tab-dance fork)

This repository is the `tab-dance/elvish` fork of the Elvish shell. It keeps the
upstream Elvish language and shell, and adds a few features aimed at external
completion tools such as [carapace](https://github.com/rsteube/carapace).

## Enhancements over upstream

-   `$edit:completion:command-completer`: override command-name completion.

-   `$edit:completion:variable-completer`: override variable-name completion.

-   `runtime:context` and `eval.Evaler.Context()`: live shell state
    introspection (aliases, builtins, functions, variables).

## Building

```sh
go build -o ./elvish ./cmd/elvish
# or
go install ./cmd/elvish
```

See [docs/building.md](docs/building.md) for details and alternative entrypoints
(`cmd/nodaemon/elvish`, `cmd/withpprof/elvish`).

## Testing and checks

```sh
make test
make most-checks
```

See [docs/testing.md](docs/testing.md) and
[docs/workflows.md](docs/workflows.md).

## Documentation

User documentation is hosted at <https://elv.sh>. Development documentation is
in [docs/](docs/).

## License

Elvish is released under the BSD 2-clause license, except for files noted in the
original upstream
[README](https://github.com/elves/elvish/blob/master/README.md). See
[LICENSE](LICENSE).
