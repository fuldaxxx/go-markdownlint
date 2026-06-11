# hujson

A minimal, dependency-free standardizer that converts JSON with comments and
trailing commas (JWCC / HuJSON) into standard JSON that `encoding/json` can
parse. It exists so the module can support JSONC config without pulling in an
external dependency that would raise the minimum Go version.

## API

```go
func Standardize(b []byte) ([]byte, error)
```

`Standardize` removes the HuJSON-specific extensions:

- `//` line comments and `/* … */` block comments — replaced with spaces so byte
  offsets and line numbers are preserved.
- trailing commas before a closing `}` or `]`.

It is string-aware: `//`, `/*`, and `,` inside JSON string literals are left
untouched. The output is **not** validated as JSON; callers (here,
[`configparse.JSONC`](../../configparse)) unmarshal the result to detect syntax
errors.

This is a focused reimplementation of the single function the config parsers
need, not a full HuJSON library.
