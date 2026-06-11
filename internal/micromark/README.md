# micromark

A self-contained Markdown tokenizer that produces a token tree compatible with
the [micromark](https://github.com/micromark/micromark) token model used by
markdownlint. Rules consume this tree instead of raw text.

## API

```go
func Parse(markdown string) *Document
```

- `Token` — a node in the tree: `Type` (a `TokenType` string such as
  `"atxHeading"`), 1-based line/column spans, the covered source `Text`, and
  parent/child links.
- `Document` — the parse result: top-level `Children` plus `Flat`, a depth-first
  list of every token for fast filtering.
- `TokenType` constants — the full micromark token vocabulary (headings, code,
  lists, blockquotes, emphasis, links/images, autolinks, HTML, GFM tables and
  footnotes, math, character escapes, and a few synthetic types).

Column values are normalized to micromark's exclusive end-column convention so
`length == EndColumn - StartColumn`.

## How it parses

`Parse` runs two passes: a first pass to discover link-reference definition
labels, then a second pass that uses that label set to classify bracket forms as
real links/images versus undefined references. Block structure
(`block.go`/`container.go`) is parsed into logical lines, then inline structure
(`inline.go`) — emphasis, code spans, links, autolinks — is resolved.

The string values of `TokenType` constants are kept identical to micromark's so
rules and parser tests can compare against them directly. Token-tree queries live
in [`mdhelpers`](../mdhelpers).
