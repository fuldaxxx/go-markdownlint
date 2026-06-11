# internal

Implementation packages for go-markdownlint. Everything here is private to the
module (not importable by external code); the public API lives in the root
[`markdownlint`](../) package.

## Packages

| Package | Responsibility |
| --- | --- |
| [`engine`](engine) | Top-level orchestration: resolve config, compute per-line enabled rules, parse, run rules (with concurrency), collect results. |
| [`rules`](rules) | The 53 built-in `MD###` rule implementations and their registry. |
| [`rule`](rule) | The `Rule` descriptor type and `RuleParams` — the API a rule sees. |
| [`micromark`](micromark) | Markdown tokenizer producing a micromark-compatible token tree. |
| [`mdhelpers`](mdhelpers) | Query helpers over the token tree (filtering, descendants, heading info). |
| [`helpers`](helpers) | Error reporting plus string, line-ending, and reference-link utilities. |
| [`fix`](fix) | Applies the fixes that rules emit. |
| [`cache`](cache) | Per-document memoization of token filtering and reference data. |
| [`types`](types) | Shared value types (`Configuration`, `Error`, `FixInfo`, …). |
| [`hujson`](hujson) | Minimal, dependency-free JSONC → JSON standardizer. |

## Rough data flow

```
engine.Run
  ├─ configparse / config_resolve   → one resolved typed Configuration
  ├─ micromark.Parse                → token tree (+ flat list)
  ├─ cache.New                      → per-document memoization
  └─ for each enabled rule:
        rule.Fn(RuleParams, OnError) → types.Error findings
        (reads p.Config.MD###; uses mdhelpers + helpers)
```

Fixes reported by rules are later applied by [`fix`](fix).
