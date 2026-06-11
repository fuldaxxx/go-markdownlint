# cache

Per-document memoization of the most expensive read operations: token filtering
by type and reference-link/image data extraction.

## API

```go
func New(flat []*mm.Token) *Cache
```

A `Cache` is **request-scoped** — one per document, used by a single goroutine —
so it needs no locking. The engine creates one per document and hands it to every
rule (via [`RuleParams`](../rule)), so repeated queries across the document's
rules are computed once and reused.

It memoizes:

- `FilterByTypes` lookups, keyed by the requested token types and the HTML-flow
  flag.
- `GetReferenceLinkImageData`, computed once per document.

Backed by the flattened token list from [`micromark`](../micromark) and the
helpers in [`mdhelpers`](../mdhelpers) / [`helpers`](../helpers).
