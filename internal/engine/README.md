# engine

Lint orchestration. `engine.Run` is the workhorse behind the public `Lint` API:
it resolves configuration, decides which rules are active on each line, parses
each document, and runs the enabled rules to produce results.

## API

```go
func Run(ctx context.Context, opts Options) (types.Results, error)
```

`Options` carries the config, custom rules, input files and/or strings, front
matter handling, concurrency, and an optional `ReadFile` hook (mirroring the
root package's `Options`).

## What it does

1. **Resolve config** — the typed `Configuration` is flattened to its untyped map
   form (`ToMap`) so rule names, aliases, and tag keys resolve uniformly; after
   merging `extends`/inline config and computing per-rule enable/severity, one
   fully-resolved typed `Configuration` is rebuilt and handed to every rule.
2. **Front matter** — strip front matter (unless disabled) so line numbers and
   rule input match the document body.
3. **Per-line enabled rules** — process inline configuration comments
   (`markdownlint-disable`, `-enable`, `-capture`, `-disable-line`, …) to build,
   for each line, the set of rules that apply (`getEnabledRulesPerLineNumber`).
4. **Parse** — tokenize via [`micromark`](../micromark) when any active rule needs
   it, and build a [`cache`](../cache) for the document.
5. **Run rules** — invoke each rule's function with [`RuleParams`](../rule),
   gathering `types.Error` findings; rule panics become errors (or are swallowed
   when `HandleRuleFailures` is set).
6. **Collect** — sort and return `types.Results` keyed by input name.

Files can be linted concurrently (`Options.Concurrency`).

## Notable internals

- `config_resolve.go` — effective-config and truthiness helpers.
- `inline_config.go` — inline configuration comment handling.
- `lint.go` — single-document lint loop and result assembly.
