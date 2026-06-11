# types

Shared value types used across the module. Kept in a dependency-free leaf package
so every other package (and the public API) can reference them without import
cycles. Most are re-exported from the root [`markdownlint`](../../) package.

## Types

| Type | Description |
| --- | --- |
| `Configuration` | The typed config object (see [Configuration](#configuration)). |
| `MD0NNConfig` | One per built-in rule (e.g. `MD013Config{LineLength *int, …}`); typed, optional (pointer) option fields tagged with their config keys. |
| `RuleConfig` | Generic entry (`Enabled`, `Severity`, `Options`) for rules without typed options and for `Custom` keys. |
| `ConfigParser` | `func(content string) (Configuration, error)` — a config parser. |
| `Error` | A single lint finding (line, rule names, description, detail, range, fix, severity). |
| `Results` | `map[string][]Error` — findings keyed by input name. |
| `ErrorInfo` | The value a rule passes to its `OnError` callback. |
| `OnError` | `func(ErrorInfo)` — the rule error-reporting callback. |
| `FixInfo` | Describes an automatic fix (edit column, delete count, insert text, …). |
| `Severity` | Error severity (`error` / `warning`). |
| `ParserType` | Which parser a rule consumes (`Micromark`, `MarkdownIt`, `None`). |

Most are plain data; the conversion logic for `Configuration` lives in
`config.go`.

## Configuration

`Configuration` (`config.go`) is a struct, not a map:

```go
type Configuration struct {
    Default *bool                   // "default" — toggle all rules (nil = unset)
    Extends []string                // "extends" — parent config paths
    MD001   MD001Config             // ...one typed field per built-in rule
    MD013   MD013Config             //    e.g. LineLength *int, CodeBlocks *bool
    // ...
    Custom  map[string]RuleConfig   // custom-rule names + tag keys
}
```

Because custom-rule names and tag keys (e.g. `"whitespace": false`) can't map to
a fixed field, they live in `Custom` — the only dynamic part.

Conversion helpers (driven by the `mdlint:"MD013,line-length"` field tags, which
list each rule's canonical name and aliases):

- `ConfigFromMap(map[string]interface{}) Configuration` — build from the untyped
  shape produced by the parsers; rule names/aliases route to typed fields,
  unknown keys go to `Custom`.
- `(Configuration).ToMap() map[string]interface{}` — render back to the untyped
  form (used by the [`engine`](../engine) so alias/tag/severity resolution keeps
  working on maps).
- `BoolOr`/`IntOr`/`StringOr(ptr, default)` — read an optional pointer field,
  falling back to the rule's default; used throughout the [`rules`](../rules).
