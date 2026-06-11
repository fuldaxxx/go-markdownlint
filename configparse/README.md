# configparse

Configuration parsers for go-markdownlint and the `Parse` driver that tries them
in order. Mirrors markdownlint's configuration loading: a config file may be
written in any supported syntax, and the first parser that succeeds wins.

## Parsers

| Function | Format | Notes |
| --- | --- | --- |
| `JSON(content)` | strict JSON | — |
| `JSONC(content)` | JSON with comments / trailing commas | standardized via [`internal/hujson`](../internal/hujson). |
| `YAML(content)` | YAML | nested `map[interface{}]interface{}` shapes are normalized to string keys. |
| `TOML(content)` | TOML | via `BurntSushi/toml`. |

Each parser has the signature `func(content string) (Configuration, error)`
(`Parser`). Internally each decodes to an untyped `map[string]interface{}` and
builds the typed [`Configuration`](../internal/types) via `types.ConfigFromMap`;
a non-object result coerces to an empty `Configuration`.

## Driver

```go
cfg, err := configparse.Parse("config.jsonc", content, configparse.Common())
```

- `Parse(name, content, parsers)` tries each parser in turn and returns the first
  successful, non-empty result; if all fail it returns an aggregated error.
- `Default()` returns `[JSON]` — the JSON-only parser list.
- `Common()` returns `[JSON, JSONC, YAML, TOML]` — all parsers.

`Configuration` and `Parser` are re-exported from [`internal/types`](../internal/types)
for convenience.
