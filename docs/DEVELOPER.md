# Developer Guide

A comprehensive guide to using **go-markdownlint** as a Go library — a port of
[markdownlint](https://github.com/DavidAnson/markdownlint) for linting (and
fixing) Markdown / CommonMark.

- Module: `github.com/ldmonster/go-markdownlint`
- Minimum Go: **1.16**
- Rule set: 53 built-in `MD###` rules (compatible with markdownlint `v0.40.0`)

## Contents

1. [Install](#install)
2. [Quick start](#quick-start)
3. [The linting API](#the-linting-api)
4. [Options reference](#options-reference)
5. [Configuration](#configuration)
6. [Reading config files & `extends`](#reading-config-files--extends)
7. [Bundled style presets](#bundled-style-presets)
8. [Working with results](#working-with-results)
9. [Applying fixes](#applying-fixes)
10. [Inline configuration comments](#inline-configuration-comments)
11. [Front matter](#front-matter)
12. [Concurrency & cancellation](#concurrency--cancellation)
13. [Rule metadata](#rule-metadata)
14. [Custom rules](#custom-rules)
15. [End-to-end example](#end-to-end-example)
16. [API reference](#api-reference)
17. [Compatibility notes](#compatibility-notes)

---

## Install

```sh
go get github.com/ldmonster/go-markdownlint
```

```go
import markdownlint "github.com/ldmonster/go-markdownlint"
```

The root package re-exports everything you need (`Configuration`, `Error`,
`Rule`, …), so a single import is usually enough. The `configparse` and `styles`
sub-packages are imported only when you need config parsers or built-in presets.

---

## Quick start

Lint a single in-memory document:

```go
package main

import (
	"context"
	"fmt"

	markdownlint "github.com/ldmonster/go-markdownlint"
)

func main() {
	doc := "#Heading\n\nSome  text.\n"

	errs, err := markdownlint.LintString(
		context.Background(),
		"doc.md",
		doc,
		markdownlint.ConfigFromMap(map[string]interface{}{"default": true}),
	)
	if err != nil {
		panic(err)
	}

	for _, e := range errs {
		fmt.Printf("doc.md:%d %s %s\n", e.LineNumber, e.RuleNames[0], e.RuleDescription)
	}
}
```

---

## The linting API

There are three entry points; all take a `context.Context` and a
[`Configuration`](#configuration).

### `LintString` — one in-memory document

```go
func LintString(ctx context.Context, name, content string, cfg Configuration) ([]Error, error)
```

Returns the findings for that one document, sorted by rule then line.

### `LintFiles` — files on disk

```go
func LintFiles(ctx context.Context, files []string, cfg Configuration) (Results, error)
```

Reads and lints each path. `Results` is `map[string][]Error` keyed by file name.

### `Lint` — full control via `Options`

```go
func Lint(ctx context.Context, opts Options) (Results, error)
```

`Lint` is the most flexible form. It can lint files **and** in-memory strings in
one call, register custom rules, set concurrency, override file reading, and
more. `LintString`/`LintFiles` are thin convenience wrappers over it.

```go
res, err := markdownlint.Lint(ctx, markdownlint.Options{
	Strings: map[string]string{
		"a.md": "# Title\n",
		"b.md": "## Subhead\n",
	},
	Files:  []string{"README.md", "CHANGELOG.md"},
	Config: markdownlint.ConfigFromMap(map[string]interface{}{"default": true}),
})
```

---

## Options reference

```go
type Options struct {
	Config             Configuration                       // the configuration (see below)
	ConfigParsers      []ConfigParser                      // parsers for CONFIGURE-FILE inline config
	CustomRules        []*Rule                             // additional rules to run
	Files              []string                            // file paths to lint
	Strings            map[string]string                   // in-memory docs (name -> content)
	FrontMatter        *regexp.Regexp                      // front matter matcher (nil = default)
	DisableFrontMatter bool                                // do not strip front matter
	HandleRuleFailures bool                                // turn rule panics into findings, not errors
	NoInlineConfig     bool                                // ignore markdownlint-* comments
	Concurrency        int                                 // parallel documents (0 = a sensible default)
	ReadFile           func(path string) (string, error)   // override file reading (nil = os.ReadFile)
}
```

| Field | Purpose |
| --- | --- |
| `Config` | The configuration to apply. Zero value (`Configuration{}`) means "all rules enabled". |
| `ConfigParsers` | Parser set used for `markdownlint-configure-file` inline blocks. Defaults to JSON only; pass `configparse.Common()` to accept JSON/JSONC/YAML/TOML. |
| `CustomRules` | Your own `*Rule` values, run alongside the built-ins. See [Custom rules](#custom-rules). |
| `Files` / `Strings` | Inputs. You can supply either or both; results are keyed by file path or string name. |
| `FrontMatter` | Regexp that matches leading front matter to strip. `nil` uses the default matcher. |
| `DisableFrontMatter` | Set to keep front matter in the linted content. |
| `HandleRuleFailures` | When `true`, a panicking rule produces a finding instead of failing the whole run. |
| `NoInlineConfig` | When `true`, `markdownlint-disable`/`-enable`/etc. comments are ignored. |
| `Concurrency` | Number of documents linted in parallel. `0` picks a default. |
| `ReadFile` | Inject a custom reader (e.g. an in-memory FS or a VFS). Defaults to `os.ReadFile`. |

---

## Configuration

`Configuration` is a **typed struct**, not a map. It has explicit fields for the
well-known keys and one typed config struct per built-in rule:

```go
type Configuration struct {
	Default *bool                   // "default": toggle all rules (nil = unset)
	Extends []string                // "extends": parent config paths
	MD013   MD013Config             // ...one field per built-in rule
	// ... MD001 ... MD060
	Custom  map[string]RuleConfig   // custom-rule names and tag keys
}

type MD013Config struct {
	Enabled             *bool   // nil = inherit "default"
	Severity            string  // "" or "warning"
	LineLength          *int    // option (nil = rule default, 80)
	CodeBlocks          *bool
	Tables              *bool
	// ...
}
```

There are two ways to build one.

### From the untyped map shape — `ConfigFromMap`

This mirrors a markdownlint config file (rule names, aliases, `"default"`,
`"extends"`, options objects), and is usually the most convenient:

```go
cfg := markdownlint.ConfigFromMap(map[string]interface{}{
	"default": true,                                            // enable all rules
	"MD013":   map[string]interface{}{"line_length": 100},      // configure a rule (also enables it)
	"MD033":   false,                                           // disable a rule
	"line-length": map[string]interface{}{"tables": false},    // aliases work too
})
```

Rules can be referenced by canonical name (`MD013`) or alias (`line-length`).
Keys that aren't built-in rules — **custom-rule names** and **tag keys** — land
in `Custom`.

### With struct fields directly

Useful when you want compile-time-checked config. The per-rule config types
(`MD013Config`, …) and the generic `RuleConfig` are re-exported from the root
package:

```go
// Optional fields are pointers, so a tiny helper keeps things readable.
ptr := func(b bool) *bool { return &b }

ll := 100
cfg := markdownlint.Configuration{
	Default: ptr(true),
	MD013:   markdownlint.MD013Config{LineLength: &ll, Tables: ptr(false)},
}
```

Pointer fields distinguish "unset → use the rule's default" from an explicit
value (e.g. an explicit `line_length: 0`). For most code, `ConfigFromMap` is
simpler; reach for struct fields when you want the compiler to check key names.

### Enabling / disabling rules

| Intent | Map form | Effect |
| --- | --- | --- |
| Enable everything | `"default": true` | All rules on. |
| Disable everything, then opt in | `"default": false, "MD013": true` | Only MD013 runs. |
| Configure a rule | `"MD013": {"line_length": 100}` | Enables MD013 with options. |
| Disable one rule | `"MD013": false` | MD013 off. |
| Set severity | `"MD013": {"severity": "warning"}` | Findings carry `SeverityWarning`. |
| Disable a whole tag group | `"whitespace": false` | All rules tagged `whitespace` off. |

### Reading typed config in your own code

```go
cfg := markdownlint.ConfigFromMap(map[string]interface{}{"MD013": map[string]interface{}{"line_length": 120}})

if cfg.MD013.LineLength != nil {
	fmt.Println("line length:", *cfg.MD013.LineLength) // 120
}
// Round-trip back to the untyped form if needed:
m := cfg.ToMap()
```

---

## Reading config files & `extends`

`ReadConfig` loads a config file from disk, resolving `extends` references
relative to the file's directory (parent-then-child merge; the child wins):

```go
cfg, err := markdownlint.ReadConfig(".markdownlint.jsonc", nil)
if err != nil {
	// missing file, parse error, or circular extends
}
res, _ := markdownlint.LintFiles(ctx, files, cfg)
```

Passing `nil` for the parser list uses the common set (JSON, JSONC, YAML, TOML).

To parse config **content** yourself, use the `configparse` package:

```go
import "github.com/ldmonster/go-markdownlint/configparse"

cfg, err := configparse.Parse("config.jsonc", content, configparse.Common())
```

| Function | Format |
| --- | --- |
| `configparse.JSON` | strict JSON |
| `configparse.JSONC` | JSON with comments / trailing commas |
| `configparse.YAML` | YAML |
| `configparse.TOML` | TOML |
| `configparse.Default()` | `[JSON]` |
| `configparse.Common()` | `[JSON, JSONC, YAML, TOML]` |
| `configparse.Parse(name, content, parsers)` | try each parser; first success wins |

---

## Bundled style presets

The `styles` package provides ready-made configurations matching the upstream
style files:

```go
import "github.com/ldmonster/go-markdownlint/styles"

errs, _ := markdownlint.LintString(ctx, "doc.md", content, styles.Relaxed())
```

| Function | Preset |
| --- | --- |
| `styles.All()` | every rule enabled |
| `styles.Relaxed()` | lenient everyday preset |
| `styles.Prettier()` | disables rules that conflict with Prettier |
| `styles.Cirosantilli()` | the "cirosantilli" community style |

Each returns a `Configuration` you can pass directly, or merge with your own (a
preset is just a value — build on it with `ConfigFromMap` overrides via
`extends`, or by reading its `ToMap()` and layering keys).

---

## Working with results

A finding is an `Error`:

```go
type Error struct {
	LineNumber      int      // 1-based, in the document body (front matter excluded)
	RuleNames       []string // e.g. ["MD013", "line-length"]
	RuleDescription string
	RuleInformation string   // URL with more info, if any
	ErrorDetail     string   // e.g. "Expected: 80; Actual: 120"
	ErrorContext    string   // surrounding text, if any
	ErrorRange      *[2]int  // [column, length], 1-based; nil if absent
	FixInfo         *FixInfo // non-nil when the finding is auto-fixable
	Severity        Severity // SeverityError or SeverityWarning
}
```

Render the default human-readable report:

```go
res, _ := markdownlint.LintFiles(ctx, files, cfg)
fmt.Print(markdownlint.ResultsString(res))
// README.md: 12: MD013/line-length Line length [Expected: 80; Actual: 95]
```

Emit JSON (every field has a `json` tag):

```go
import "encoding/json"

b, _ := json.MarshalIndent(res, "", "  ")
fmt.Println(string(b))
```

Decide an exit code from whether anything fired:

```go
total := 0
for _, errs := range res {
	total += len(errs)
}
if total > 0 {
	os.Exit(1)
}
```

---

## Applying fixes

Findings whose `FixInfo` is non-nil can be auto-fixed. The easiest path is
`ApplyFixes`, which applies as many fixes as possible to a whole document:

```go
content := "# Title \n"                                  // trailing space (MD009)
errs, _ := markdownlint.LintString(ctx, "doc.md", content,
	markdownlint.ConfigFromMap(map[string]interface{}{"default": true}))

fixed := markdownlint.ApplyFixes(content, errs)          // "# Title\n"
```

`ApplyFixes` resolves overlapping fixes, accounts for front-matter offsets, and
preserves the document's line endings.

For line-level control there is `ApplyFix`:

```go
func ApplyFix(line string, fi FixInfo, lineEnding string) (newLine string, keep bool)
```

`keep` is `false` when the fix deletes the line entirely.

```go
type FixInfo struct {
	LineNumber  int    // 1-based; 0 = use the error's line
	EditColumn  int    // 1-based; 0 normalizes to 1
	DeleteCount int    // -1 deletes the whole line
	InsertText  string
}
```

A typical "lint, fix, rewrite" flow:

```go
errs, _ := markdownlint.LintString(ctx, path, content, cfg)
if fixed := markdownlint.ApplyFixes(content, errs); fixed != content {
	_ = os.WriteFile(path, []byte(fixed), 0o644)
}
```

---

## Inline configuration comments

Documents can adjust linting with HTML comments, exactly as in upstream
markdownlint:

```markdown
<!-- markdownlint-disable MD013 -->
A very long line that would normally trigger MD013...
<!-- markdownlint-enable MD013 -->

<!-- markdownlint-disable-next-line MD033 -->
<div>inline html allowed only on the next line</div>

<!-- markdownlint-configure-file { "MD013": { "line_length": 120 } } -->
```

Supported directives include `disable`, `enable`, `disable-line`,
`disable-next-line`, `capture`, `restore`, `enable-file`, `disable-file`, and
`configure-file`.

- `configure-file` blocks are parsed with `Options.ConfigParsers` (JSON only by
  default — set `configparse.Common()` to allow JSONC/YAML/TOML).
- Set `Options.NoInlineConfig = true` (or pass it through the CLI) to ignore all
  of these comments.

---

## Front matter

Leading YAML/TOML front matter is detected and stripped before linting, so rules
and reported line numbers refer to the document body. Reported `LineNumber`s are
relative to the body, not the raw file.

- Override the matcher with `Options.FrontMatter` (a `*regexp.Regexp`).
- Set `Options.DisableFrontMatter = true` to lint the front matter as content.

---

## Concurrency & cancellation

`Lint` processes documents concurrently. Control the degree with
`Options.Concurrency` (0 picks a default based on the input size). The work
honors the provided `context.Context`, so you can bound a run with a timeout or
cancel it:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

res, err := markdownlint.Lint(ctx, opts)
if errors.Is(err, context.DeadlineExceeded) {
	// linting was cancelled
}
```

---

## Rule metadata

Enumerate the built-in rules (for `--list`-style output, docs, or validation):

```go
for _, r := range markdownlint.Rules() {
	fmt.Printf("%-7s %-30s fixable=%v tags=%v\n",
		r.Names[0], r.Description, r.Fixable, r.Tags)
}

fmt.Println("library version:", markdownlint.Version())
```

```go
type RuleInfo struct {
	Names       []string
	Description string
	Tags        []string
	Fixable     bool
}
```

---

## Custom rules

Add your own rules through `Options.CustomRules`. A rule is a `*Rule`:

```go
type Rule struct {
	Names        []string  // canonical name first, then aliases
	Description  string
	Tags         []string
	Parser       ParserType // ParserNone, ParserMicromark, or ParserMarkdownIt
	Information  *url.URL
	Asynchronous bool
	Fn           func(p *RuleParams, onError OnError)
}
```

Inside `Fn` you inspect the document and call `onError` for each violation:

```go
type ErrorInfo struct {
	LineNumber  int      // 1-based, in the document body
	Detail      string
	Context     string
	Information *url.URL
	Range       *[2]int  // [column, length], 1-based
	FixInfo     *FixInfo
}
```

`RuleParams` gives a rule its view of the document:

```go
type RuleParams struct {
	Name             string
	Version          string
	Lines            []string        // the document body, split into lines
	FrontMatterLines []string
	Config           Configuration   // the fully-resolved config (shared by all rules)
	// plus token accessors (see note below)
}
```

A custom rule reads its own options from `p.Config.Custom[<its name>]`:

```go
var noTodo = &markdownlint.Rule{
	Names:       []string{"no-todo"},
	Description:  "TODO markers are not allowed",
	Tags:        []string{"style"},
	Parser:      markdownlint.ParserNone,
	Fn: func(p *markdownlint.RuleParams, onError markdownlint.OnError) {
		for i, line := range p.Lines {
			if strings.Contains(line, "TODO") {
				col := strings.Index(line, "TODO") + 1
				onError(markdownlint.ErrorInfo{
					LineNumber: i + 1,
					Detail:     "remove the TODO",
					Range:      &[2]int{col, len("TODO")},
				})
			}
		}
	},
}

res, _ := markdownlint.Lint(ctx, markdownlint.Options{
	Strings:     map[string]string{"a.md": "x\nTODO: finish\n"},
	Config:      markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "no-todo": true}),
	CustomRules: []*markdownlint.Rule{noTodo},
})
```

> **Token-tree note.** `RuleParams` also exposes `MicromarkTokens()`,
> `FilterByTypesCached(...)`, and `ReferenceLinkImageData()`, but these return
> types from the module-internal `micromark` package, which external code cannot
> import. External custom rules therefore work line-by-line via `p.Lines` (use
> `Parser: ParserNone`). Token-tree rules are practical only for code living
> inside this module. The built-in rules use the token tree heavily and are good
> references.

Set `Options.HandleRuleFailures = true` if you want a panicking custom rule to
surface as a finding rather than fail the whole `Lint` call.

---

## End-to-end example

A minimal lint-and-fix tool:

```go
package main

import (
	"context"
	"fmt"
	"os"

	markdownlint "github.com/ldmonster/go-markdownlint"
	"github.com/ldmonster/go-markdownlint/configparse"
)

func main() {
	ctx := context.Background()

	cfg, err := markdownlint.ReadConfig(".markdownlint.jsonc", configparse.Common())
	if err != nil {
		// fall back to "all rules enabled"
		cfg = markdownlint.ConfigFromMap(map[string]interface{}{"default": true})
	}

	files := os.Args[1:]
	res, err := markdownlint.LintFiles(ctx, files, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lint error:", err)
		os.Exit(2)
	}

	violations := 0
	for _, file := range files {
		errs := res[file]
		violations += len(errs)

		// Auto-fix what we can and rewrite the file.
		content, _ := os.ReadFile(file)
		if fixed := markdownlint.ApplyFixes(string(content), errs); fixed != string(content) {
			_ = os.WriteFile(file, []byte(fixed), 0o644)
		}
	}

	fmt.Print(markdownlint.ResultsString(res))
	if violations > 0 {
		os.Exit(1)
	}
}
```

There is also a ready-made CLI at
[`cmd/markdownlint`](../cmd/markdownlint) you can study or use directly.

---

## API reference

Root package `markdownlint` (all re-exported for one-import use):

| Symbol | Kind | Summary |
| --- | --- | --- |
| `Lint(ctx, Options)` | func | Lint files and/or strings with full options. |
| `LintString(ctx, name, content, cfg)` | func | Lint one in-memory document. |
| `LintFiles(ctx, files, cfg)` | func | Lint files on disk. |
| `ReadConfig(file, parsers)` | func | Load a config file, resolving `extends`. |
| `ConfigFromMap(m)` | func | Build a `Configuration` from the untyped map shape. |
| `ApplyFix` / `ApplyFixes` | func | Apply fixes to a line / a whole document. |
| `Rules()` | func | Metadata for all built-in rules. |
| `ResultsString(res)` | func | Default human-readable report. |
| `Version()` | func | Library version. |
| `Options` | type | Input/behavior options for `Lint`. |
| `Configuration` | type | The typed config object (+ `ConfigFromMap`, `ToMap`). |
| `Error`, `Results`, `FixInfo`, `ErrorInfo` | type | Result and fix data. |
| `Rule`, `RuleParams`, `OnError` | type | Custom-rule API. |
| `Severity` (`SeverityError`/`SeverityWarning`) | type/const | Finding severity. |
| `ParserType` (`ParserMicromark`/`ParserMarkdownIt`/`ParserNone`) | type/const | Which parser a rule consumes. |

Sub-packages:

- [`configparse`](../configparse) — JSON/JSONC/YAML/TOML parsers and the `Parse` driver.
- [`styles`](../styles) — the bundled `All`/`Relaxed`/`Prettier`/`Cirosantilli` presets.

For internal package layout (engine, rules, tokenizer, …), see the
[`internal`](../internal) READMEs.

---

## Compatibility notes

- **Go 1.16+.** The module avoids newer language/stdlib features and vendors a
  minimal JSONC standardizer so it stays compatible with old toolchains.
- **Rule parity.** Rules target markdownlint `v0.40.0` behavior. A small number
  of rules have documented behavior differences (see `*_test.go` notes); the
  differential tooling that tracked these has been removed.
- **Configuration shape.** `Configuration` is a struct. Code written against the
  old `map[string]interface{}` form should call `ConfigFromMap(...)` to build a
  value, and `cfg.ToMap()` to get the map form back.
