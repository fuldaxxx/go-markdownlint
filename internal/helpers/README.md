# helpers

General-purpose utilities shared across rules and the engine: error reporting,
string/value formatting, line-ending handling, regex helpers, and reference-link
data extraction.

## Error reporting

Convenience constructors that build `types.ErrorInfo` and hand it to a rule's
`OnError` callback:

- `AddError` — basic line-number violation.
- `AddErrorWithInfo` — with a fix and/or range.
- `AddErrorDetailIf` — adds an `Expected: … Actual: …` detail when values differ.
- `AddErrorContext` — with surrounding context (optionally ellipsified).

## Strings & values

- `Stringify(v)` — stringify a config value for concatenation (ints without a
  decimal point, bools as `true`/`false`).
- `Ellipsify(text, start, end)` — truncate with `…` markers.
- `IsString`, `IsEmptyString`, `IsBlankLine` — small predicates.
- `EscapeForRegExp(s)` — escape a string for use in a regular expression.
- `GetHTMLAttributeRe(name)` — regexp matching a given HTML attribute.
- `ClearHTMLCommentText(text)` — blank out HTML comment bodies.

## Lines & references

- `GetPreferredLineEnding(input, defaultEOL)` — detect the dominant EOL.
- `FrontMatterHasTitle(...)` — whether front matter declares a title.
- `GetReferenceLinkImageData(flat)` — collect link/image reference definitions
  and uses (`ReferenceLinkImageData`, `DefInfo`, `DupDef`); memoized per document
  by [`cache`](../cache).
