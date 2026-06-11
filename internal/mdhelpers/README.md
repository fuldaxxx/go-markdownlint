# mdhelpers

Utilities that operate on the [`micromark`](../micromark) token tree: filtering,
descendant/parent lookup, heading helpers, and HTML tag info. These are the
building blocks rules use to navigate a parsed document.

## Selected API

| Function | Purpose |
| --- | --- |
| `FilterByTypes(tokens, types, htmlFlow)` | Tokens of the given types among `tokens` (one level). |
| `FilterFlat(flat, types, htmlFlow)` | Same, over a flattened token list. |
| `FilterByPredicate(...)` | General predicate-based filter/transform. |
| `GetDescendantsByType(parents, typePath)` | Walk a path of token types to collect descendants. |
| `GetParentOfType(token, types)` | Nearest ancestor of one of the given types. |
| `GetHeadingLevel(heading)` | Heading level (1–6). |
| `GetHeadingStyle(heading)` | `atx` / `atx_closed` / `setext`. |
| `GetHeadingText(heading)` | Rendered heading text. |
| `GetHTMLTagInfo(token)` | Parsed HTML tag (`HTMLTagInfo`: name, closing, self-closing). |
| `GetBlockQuotePrefixText(...)` | Blockquote prefix text for a line. |
| `IsHTMLFlowComment(token)` | Whether a token is an HTML-flow comment. |
| `IsDocfxTab(heading)` | DocFX tab-heading detection. |
| `AddRangeToSet(set, start, end)` | Add an inclusive integer range to a set. |

The `htmlFlow` flag controls whether tokens produced by the HTML-flow reparse are
included or excluded.
