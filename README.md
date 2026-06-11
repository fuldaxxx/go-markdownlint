# go-markdownlint

A Go port of [markdownlint](https://github.com/DavidAnson/markdownlint) — a static
analysis tool for Markdown / CommonMark files. It bundles a self-contained
Markdown tokenizer, the built-in `MD###` rule library, an auto-fix engine, and
configuration handling, exposed both as a Go library and a reference CLI.

- **Module:** `github.com/ldmonster/go-markdownlint`
- **Go:** 1.16+
- **Rule set:** 53 built-in rules (compatible with markdownlint `v0.40.0`)
- **Dependencies:** only `BurntSushi/toml` and `gopkg.in/yaml.v3` (JSONC support is vendored in [`internal/hujson`](internal/hujson))

## Install

```sh
# Library
go get github.com/ldmonster/go-markdownlint

# CLI
go install github.com/ldmonster/go-markdownlint/cmd/markdownlint@latest
```

## Library usage

> 📖 For a comprehensive walkthrough — options, configuration, fixes, inline
> config, custom rules, and the full API reference — see the
> **[Developer Guide](docs/DEVELOPER.md)**.

```go
package main

import (
	"context"
	"fmt"

	markdownlint "github.com/ldmonster/go-markdownlint"
)

func main() {
	doc := "#Heading\n\nSome text.\n"
	errs, err := markdownlint.LintString(context.Background(), "doc.md", doc,
		markdownlint.ConfigFromMap(map[string]interface{}{"default": true}))
	if err != nil {
		panic(err)
	}
	for _, e := range errs {
		fmt.Printf("%s:%d %s %s\n", "doc.md", e.LineNumber,
			e.RuleNames[0], e.RuleDescription)
	}
}
```

Key entry points (root package `markdownlint`):

| Function | Purpose |
| --- | --- |
| `Lint(ctx, Options)` | Lint files and/or in-memory strings with full options. |
| `LintString(ctx, name, content, cfg)` | Lint one in-memory document. |
| `LintFiles(ctx, files, cfg)` | Lint a set of files on disk. |
| `ApplyFix` / `ApplyFixes` | Apply rule-provided fixes to text. |
| `Rules()` | Metadata for all built-in rules. |
| `ResultsString(results)` | Render results in the default human format. |
| `ReadConfig(file, parsers)` | Read a config file, resolving `extends`. |
| `ConfigFromMap(m)` | Build a `Configuration` from the untyped map shape. |
| `Version()` | Library version string. |

## CLI usage

```sh
markdownlint [flags] <file-or-glob>...

  -config string   configuration file (JSON/JSONC/YAML/TOML)
  -style string    built-in style: all|relaxed|prettier|cirosantilli
  -fix             apply fixes in place where possible
  -json            emit results as JSON
```

See the **[User Guide](docs/USER_GUIDE.md)** for full CLI usage (config files,
styles, auto-fixing, CI), or [`cmd/markdownlint`](cmd/markdownlint) for a quick
reference.

## Configuration

`Configuration` is a typed struct: a `Default` toggle, an `Extends` list, a typed
config field per built-in rule (e.g. `MD013 MD013Config`), and a `Custom` map for
custom-rule and tag keys. Build one with struct fields, or from the familiar
untyped map shape via `ConfigFromMap`:

```go
cfg := markdownlint.ConfigFromMap(map[string]interface{}{
	"default": true,
	"MD013":   map[string]interface{}{"line_length": 100},
	"MD033":   false,
})
```

Config files may be JSON, JSONC, YAML, or TOML, and support `extends`. Bundled
presets live in [`styles`](styles).

## Repository layout

| Path | Description |
| --- | --- |
| `.` | Public library API (`Lint`, `LintString`, `ReadConfig`, …). |
| [`docs`](docs) | Documentation: [user guide](docs/USER_GUIDE.md) (CLI) and [developer guide](docs/DEVELOPER.md) (library). |
| [`cmd/markdownlint`](cmd/markdownlint) | Reference command-line interface. |
| [`configparse`](configparse) | JSON/JSONC/YAML/TOML configuration parsers. |
| [`styles`](styles) | Bundled configuration presets (`all`, `relaxed`, …). |
| [`internal/engine`](internal/engine) | Lint orchestration (config resolution, scheduling). |
| [`internal/rules`](internal/rules) | The 53 built-in `MD###` rule implementations. |
| [`internal/rule`](internal/rule) | `Rule` descriptor and `RuleParams` rule API. |
| [`internal/micromark`](internal/micromark) | Markdown tokenizer (micromark-compatible token tree). |
| [`internal/mdhelpers`](internal/mdhelpers) | Token-tree query helpers used by rules. |
| [`internal/helpers`](internal/helpers) | Error reporting and string/line utilities. |
| [`internal/fix`](internal/fix) | Fix application engine. |
| [`internal/cache`](internal/cache) | Per-document memoization. |
| [`internal/types`](internal/types) | Shared types (`Configuration`, `Error`, `FixInfo`, …). |
| [`internal/hujson`](internal/hujson) | Minimal JSONC standardizer (vendored). |

## Development

The repo uses [Task](https://taskfile.dev):

```sh
task test     # go test ./... and the race suite
task lint     # gofmt + go vet + golangci-lint
task license:header   # add Apache-2.0 headers to .go files
```

## License

Apache License 2.0. See the per-file headers and `hack/license-header.txt`.
