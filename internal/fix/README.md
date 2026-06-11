# fix

Applies the fixes that rules emit. When a rule reports a violation it can attach
a `types.FixInfo` (delete count, edit column, insert text, …); this package turns
those into edited text.

## API

```go
func ApplyFix(line string, fi types.FixInfo, lineEnding string) (string, bool)
func ApplyFixes(input string, errors []types.Error) string
```

- `ApplyFix` applies a single fix to one line. The returned `bool` is `false`
  when the fix deletes the line entirely.
- `ApplyFixes` applies as many fixes as possible across a whole document: it
  resolves overlaps, accounts for front-matter line offsets, and reassembles the
  text, preserving the document's line endings.

These are re-exported as `markdownlint.ApplyFix` / `markdownlint.ApplyFixes` and
are what the CLI's `-fix` flag uses.
