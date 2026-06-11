# styles

Bundled markdownlint configuration presets, embedded at build time and exposed as
parsed `Configuration` values. These match the style files distributed with
upstream markdownlint.

## Presets

| Function | Source | Description |
| --- | --- | --- |
| `All()` | `all.json` | Enables every rule. |
| `Relaxed()` | `relaxed.json` | A lenient everyday preset. |
| `Prettier()` | `prettier.json` | Disables rules that conflict with the Prettier formatter. |
| `Cirosantilli()` | `cirosantilli.json` | The "cirosantilli" community style. |

Each returns a [`types.Configuration`](../internal/types) ready to pass to the
linter:

```go
errs, _ := markdownlint.LintString(ctx, "doc.md", content, styles.Relaxed())
```

The CLI exposes these via `-style all|relaxed|prettier|cirosantilli`.

## Files

The `*.json` files are embedded with `//go:embed` and parsed once on first use.
Tests verify each preset is non-empty and matches its source JSON.
