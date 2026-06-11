# rule

Defines the `Rule` descriptor and the `RuleParams` object passed to every rule
function. This is the contract between the [`engine`](../engine) and the
individual [`rules`](../rules).

## Types

### `Rule`

A rule descriptor: its `Names` (canonical name plus aliases), `Description`,
`Tags`, the `Parser` it consumes, optional `Information`, and `Fn` — the function
that inspects a document and reports violations.

### `RuleParams`

The per-document view a rule receives. It exposes the document name, version,
lines, front matter, the fully-resolved [`Configuration`](../types) (`Config`),
and cache-backed token accessors so rules read the document consistently and
cheaply. A rule reads its own typed options from `Config` — e.g.
`p.Config.MD013.LineLength` — rather than an untyped options map:

| Method | Returns |
| --- | --- |
| `MicromarkTokens()` | The document's top-level micromark tokens. |
| `FilterByTypesCached(types, htmlFlow)` | Tokens of the given types, memoized per document. |
| `ReferenceLinkImageData()` | Cached reference-link/image data for the document. |

```go
func NewRuleParams(name, version string, lines, frontMatter []string,
	cfg types.Configuration, tokens []*mm.Token, c *cache.Cache) *RuleParams
```

`NewRuleParams` is called by the engine with the resolved configuration for the
document (the same `Config` is shared by every rule; each reads its own field).
Token filtering is delegated to [`cache`](../cache) so repeated queries across
rules are computed once.
