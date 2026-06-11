# User Guide

How to use the **markdownlint** command-line tool to check (and fix) your
Markdown files. If you want to call the linter from Go code instead, see the
[Developer Guide](DEVELOPER.md).

## Contents

1. [Install](#install)
2. [Basic usage](#basic-usage)
3. [Reading the output](#reading-the-output)
4. [Exit codes](#exit-codes)
5. [Configuration files](#configuration-files)
6. [Enabling and disabling rules](#enabling-and-disabling-rules)
7. [Rule options & severity](#rule-options--severity)
8. [Built-in styles](#built-in-styles)
9. [Auto-fixing](#auto-fixing)
10. [Inline configuration comments](#inline-configuration-comments)
11. [Front matter](#front-matter)
12. [Using it in CI](#using-it-in-ci)
13. [Flag reference](#flag-reference)
14. [Troubleshooting](#troubleshooting)

---

## Install

With a Go toolchain (1.16+):

```sh
go install github.com/ldmonster/go-markdownlint/cmd/markdownlint@latest
```

This puts a `markdownlint` binary in your `$GOBIN` (usually `~/go/bin` — make
sure it's on your `PATH`). To build from a checkout instead:

```sh
go build -o markdownlint ./cmd/markdownlint
```

---

## Basic usage

```sh
markdownlint [flags] <file-or-glob>...
```

Lint one file:

```sh
markdownlint README.md
```

Lint several, using globs:

```sh
markdownlint README.md CHANGELOG.md
markdownlint 'docs/*.md'
markdownlint 'docs/**/*.md'      # recursive (quote globs so your shell doesn't expand them)
```

- Each argument is a file path **or** a glob pattern (`*`, `?`, `[...]`). A
  pattern that matches nothing is treated as a literal path.
- Duplicate matches are removed and the file list is sorted.
- At least one input file is required.

---

## Reading the output

By default, each finding is printed on one line:

```
README.md: 12: MD013/line-length Line length [Expected: 80; Actual: 95]
README.md: 18: MD009/no-trailing-spaces Trailing spaces [Expected: 0 or 2; Actual: 1]
docs/intro.md: 3: MD041/first-line-heading/first-line-h1 First line in a file should be a top-level heading
```

The fields are:

```
<file>: <line>: <RULE>/<aliases> <description> [<detail>] [Context: "<text>"]
```

- **line** is 1-based and refers to the document body (front matter excluded).
- **RULE** is the canonical id (`MD013`) followed by its aliases (`line-length`).
- **detail** and **context** appear only when the rule provides them.

For machine-readable output, add `--json`:

```sh
markdownlint --json README.md
```

```json
{
  "README.md": [
    {
      "lineNumber": 12,
      "ruleNames": ["MD013", "line-length"],
      "ruleDescription": "Line length",
      "errorDetail": "Expected: 80; Actual: 95",
      "errorRange": [81, 15],
      "fixInfo": null,
      "severity": "error"
    }
  ]
}
```

---

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | No violations found. |
| `1` | One or more violations were found. |
| `2` | An error occurred (bad flag value, unreadable config, no input files, …). |

This makes the tool easy to gate a build on: a non-zero exit fails the step.

---

## Configuration files

Pass a configuration file with `--config`:

```sh
markdownlint --config .markdownlint.jsonc 'docs/**/*.md'
```

> The CLI does **not** auto-discover a config file — you must point at one with
> `--config`. With no `--config` and no `--style`, every rule is enabled.

Config files may be **JSON, JSONC** (JSON with comments and trailing commas),
**YAML**, or **TOML**. A typical `.markdownlint.jsonc`:

```jsonc
{
  // Start from "all rules enabled", then adjust.
  "default": true,

  // Allow longer lines, and don't measure lines in tables.
  "MD013": { "line_length": 100, "tables": false },

  // Allow inline HTML.
  "MD033": false,

  // Rules can be referenced by their alias too.
  "heading-style": { "style": "atx" }
}
```

The same config in YAML:

```yaml
default: true
MD013:
  line_length: 100
  tables: false
MD033: false
heading-style:
  style: atx
```

### Extending another config

Use `extends` to inherit from a base file (resolved relative to the current
file; the current file's keys win on conflicts):

```jsonc
{
  "extends": "../.markdownlint-base.json",
  "MD013": false
}
```

---

## Enabling and disabling rules

The special `default` key sets the baseline; then list rules to override it.

| Goal | Config |
| --- | --- |
| Enable everything | `"default": true` |
| Disable everything, enable a few | `"default": false`, then `"MD013": true` |
| Turn one rule off | `"MD033": false` |
| Configure a rule (also enables it) | `"MD013": { "line_length": 100 }` |
| Turn a whole **tag** group off | `"whitespace": false` |

Rules can be named by id (`MD013`) or alias (`line-length`). A **tag** key (e.g.
`whitespace`, `headings`, `code`) toggles every rule carrying that tag at once.

---

## Rule options & severity

Each rule that has options takes them as an object. For example, MD013
(line-length):

```jsonc
{
  "MD013": {
    "line_length": 100,
    "code_block_line_length": 120,
    "tables": false,
    "headings": true
  }
}
```

Add `"severity": "warning"` to downgrade a rule's findings (they still print,
and JSON output shows `"severity": "warning"`):

```jsonc
{ "MD013": { "line_length": 100, "severity": "warning" } }
```

For the full list of rules and their options, see the upstream
[markdownlint rules documentation](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md);
this tool targets that rule set (version 0.40.0). To list the rules your binary
knows about and which are auto-fixable, the library's `Rules()` API exposes the
metadata (see the [Developer Guide](DEVELOPER.md#rule-metadata)).

---

## Built-in styles

Instead of writing a config, you can pick a bundled preset with `--style`:

```sh
markdownlint --style relaxed 'docs/**/*.md'
```

| `--style` | Description |
| --- | --- |
| `all` | Every rule enabled. |
| `relaxed` | A lenient everyday preset. |
| `prettier` | Disables rules that conflict with the Prettier formatter. |
| `cirosantilli` | The "cirosantilli" community style. |

> `--style` takes precedence over `--config`: if you pass both, the style is
> used and the config file is ignored.

---

## Auto-fixing

Many findings can be fixed automatically (trailing spaces, missing final
newline, hard tabs, spacing issues, and more). Add `--fix` to rewrite files in
place:

```sh
markdownlint --fix 'docs/**/*.md'
```

- Only fixable findings are changed; everything else is reported as usual.
- Files are rewritten only when a fix actually applies.
- Run it on a clean working tree (or under version control) so you can review
  the diff afterward.

> Note: `--fix` still reports — and exits `1` for — the violations it found in
> this run, even ones it just fixed. Re-run without `--fix` to confirm the file
> is now clean.

---

## Inline configuration comments

You can adjust linting from within a document using HTML comments, the same as
upstream markdownlint:

```markdown
<!-- markdownlint-disable MD013 -->
This very long line is exempt from the line-length rule...
<!-- markdownlint-enable MD013 -->

<!-- markdownlint-disable-next-line MD033 -->
<div>inline HTML allowed only on the next line</div>

<!-- markdownlint-configure-file { "MD013": { "line_length": 120 } } -->
```

Supported directives: `disable`, `enable`, `disable-line`, `disable-next-line`,
`capture`, `restore`, `enable-file`, `disable-file`, and `configure-file`.
A directive with no rule names applies to all rules.

`configure-file` blocks may be written in JSON, JSONC, YAML, or TOML.

---

## Front matter

Leading front matter at the top of a file — YAML (`---` … `---`), TOML
(`+++` … `+++`), or JSON (`{` … `}`) — is recognized and excluded from linting,
so rules and the reported line numbers apply to the document body, not the
metadata block.

---

## Using it in CI

Because the tool exits non-zero on violations, wiring it into CI is a one-liner.

GitHub Actions:

```yaml
- name: Lint Markdown
  run: |
    go install github.com/ldmonster/go-markdownlint/cmd/markdownlint@latest
    markdownlint --config .markdownlint.jsonc 'docs/**/*.md' '**/*.md'
```

A pre-commit-style check:

```sh
#!/bin/sh
markdownlint --config .markdownlint.jsonc $(git diff --cached --name-only --diff-filter=ACM '*.md') || {
  echo "Markdown lint failed. Fix the issues above (or run with --fix)."
  exit 1
}
```

---

## Flag reference

| Flag | Default | Description |
| --- | --- | --- |
| `--config <file>` | _none_ | Configuration file (JSON / JSONC / YAML / TOML). |
| `--style <name>` | _none_ | Built-in preset: `all`, `relaxed`, `prettier`, `cirosantilli`. Overrides `--config`. |
| `--fix` | off | Apply fixes in place where rules support them. |
| `--json` | off | Emit results as JSON instead of the human-readable format. |

Flags accept either one or two dashes (`-fix` or `--fix`) and either
`--config x` or `--config=x`.

---

## Troubleshooting

**"no input files"** — you didn't pass any path/glob, or your shell expanded a
glob that matched nothing. Quote recursive globs (`'docs/**/*.md'`) so the tool
expands them.

**"unknown style ..."** — `--style` must be one of `all`, `relaxed`,
`prettier`, `cirosantilli`.

**My `--config` seems ignored** — check you didn't also pass `--style`, which
takes precedence.

**A config key isn't taking effect** — make sure you used a valid rule id
(`MD013`), alias (`line-length`), or tag (`whitespace`); unknown keys are
treated as custom-rule/tag entries and silently ignored by the built-in rules.

**Exit code 1 after `--fix`** — expected; the run still reports what it found.
Re-run without `--fix` to verify the files are now clean.
