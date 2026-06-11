# markdownlint (CLI)

`package main` — the reference command-line interface for go-markdownlint.

## Usage

```sh
markdownlint [flags] <file-or-glob>...
```

Each argument is a file path or a glob; all matched Markdown files are linted.

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-config` | _none_ | Configuration file to load (JSON / JSONC / YAML / TOML). |
| `-style` | _none_ | Built-in style preset: `all`, `relaxed`, `prettier`, or `cirosantilli`. |
| `-fix` | `false` | Apply fixes in place where rules support them. |
| `-json` | `false` | Emit results as JSON instead of the human-readable format. |

## Behavior

- Findings are printed in the default `file:line RULE description` format, or as
  JSON when `-json` is set.
- With `-fix`, fixable violations are rewritten in place and the file is updated.
- The process exits non-zero when violations remain (so it can gate CI).
- Unknown `-style` values and missing input files are reported as errors.

## Examples

```sh
# Lint a tree of docs with the relaxed preset
markdownlint -style relaxed 'docs/**/*.md'

# Lint with a project config and auto-fix
markdownlint -config .markdownlint.jsonc -fix README.md

# Machine-readable output
markdownlint -json CHANGELOG.md
```

For end-to-end usage (configuration files, styles, auto-fixing, inline config,
CI), see the [User Guide](../../docs/USER_GUIDE.md).

The CLI is a thin wrapper over the public [`markdownlint`](../../) library API
(`Lint`, `ApplyFixes`, `ResultsString`, and the [`styles`](../../styles) presets).
