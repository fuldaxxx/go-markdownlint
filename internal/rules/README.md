# rules

The built-in rule library: 53 `MD###` rules compatible with markdownlint
`v0.40.0`. Each rule lives in its own file (`md001.go`, `md003.go`, …) and
registers itself into a shared registry at init time.

## Registry API

| Symbol | Purpose |
| --- | --- |
| `BuiltIn() []*rule.Rule` | All registered rules, in a stable order. |
| `Version` | Library version string (`"0.40.0"`). |
| `FixableRuleNames` | Names of rules that can emit fixes. |
| `Homepage` | Upstream project URL. |

## Anatomy of a rule

A rule is a [`rule.Rule`](../rule) value with names/aliases, a description, tags,
the parser it consumes, and a function:

```go
var md001 = rule.Rule{
	Names:       []string{"MD001", "heading-increment"},
	Description: "Heading levels should only increment by one level at a time",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD013 // this rule's typed config
		lineLength := types.IntOr(c.LineLength, 80)
		// inspect p.MicromarkTokens() / p.FilterByTypesCached(...) and
		// call onError(types.ErrorInfo{...}) for each violation
	},
}

func init() { register(&md001) }
```

Rules read the document through [`RuleParams`](../rule) (cache-backed token
accessors), query the tree with [`mdhelpers`](../mdhelpers), and report findings
via the `OnError` callback. Each rule reads its **typed config** from the
resolved [`Configuration`](../types) struct (`p.Config.MD013.LineLength`, …);
optional values use pointer fields, with `types.BoolOr`/`IntOr`/`StringOr`
applying the rule's default when unset. Fixable rules attach a `FixInfo` to
their errors, which [`fix`](../fix) later applies.

## Files

- `registry.go` — registration, `BuiltIn`, `Version`, `FixableRuleNames`.
- `md0NN.go` — one file per rule (reads its `types.MD0NNConfig` field), each with a `md0NN_test.go` beside it.
