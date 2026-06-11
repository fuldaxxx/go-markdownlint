// Copyright 2026 go-markdownlint Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rules

import (
	"regexp"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() {
	register(&md049)
	register(&md050)
}

var md049IntrawordRe = regexp.MustCompile(`^\w$`)

// md049EmphasisOrStrongStyleFor returns the string representation of an
// emphasis or strong markup character.
func md049EmphasisOrStrongStyleFor(markup string) string {
	if len(markup) > 0 && markup[0] == '*' {
		return "asterisk"
	}

	return "underscore"
}

// md049Impl is the shared implementation for MD049 and MD050.
func md049Impl(
	p *rule.RuleParams,
	onError types.OnError,
	typ, typeSequence mm.TokenType,
	asterisk, underline, style string,
) {
	lines := p.Lines
	emphasisTokens := mdhelpers.FilterByPredicate(
		p.MicromarkTokens(),
		func(token *mm.Token) bool { return token.Type == typ },
		func(token *mm.Token) []*mm.Token {
			if token.Type == mm.TypeHTMLFlow {
				return []*mm.Token{}
			}

			return token.Children
		},
	)

	// intrawordAt reports whether the rune at the given position is intraword;
	// out-of-range indices yield false.
	intrawordAt := func(lineIndex, runeIndex int) bool {
		if lineIndex < 0 || lineIndex >= len(lines) {
			return false
		}

		r := []rune(lines[lineIndex])
		if runeIndex < 0 || runeIndex >= len(r) {
			return false
		}

		return md049IntrawordRe.MatchString(string(r[runeIndex]))
	}

	for _, token := range emphasisTokens {
		sequences := mdhelpers.GetDescendantsByType(
			[]*mm.Token{token},
			[][]mm.TokenType{{typeSequence}},
		)
		if len(sequences) == 0 {
			continue
		}

		startSequence := sequences[0]

		endSequence := sequences[len(sequences)-1]
		if startSequence != nil && endSequence != nil {
			markupStyle := md049EmphasisOrStrongStyleFor(startSequence.Text)
			if style == "consistent" {
				style = markupStyle
			}

			if style != markupStyle {
				underscoreIntraword := (style == "underscore") &&
					(intrawordAt(startSequence.StartLine-1, startSequence.StartColumn-2) ||
						intrawordAt(endSequence.EndLine-1, endSequence.EndColumn-1))
				if !underscoreIntraword {
					for _, sequence := range []*mm.Token{startSequence, endSequence} {
						insertText := underline
						if style == "asterisk" {
							insertText = asterisk
						}

						helpers.AddError(
							onError,
							sequence.StartLine,
							"Expected: "+style+"; Actual: "+markupStyle,
							"",
							rng(sequence.StartColumn, runeLen(sequence.Text)),
							fixReplace(sequence.StartColumn, runeLen(sequence.Text), insertText),
						)
					}
				}
			}
		}
	}
}

var md049 = rule.Rule{
	Names:       []string{"MD049", "emphasis-style"},
	Description: "Emphasis style",
	Tags:        []string{"emphasis"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD049

		style := types.StringOr(c.Style, "consistent")
		if style == "" {
			style = "consistent"
		}

		md049Impl(p, onError, mm.TypeEmphasis, mm.TypeEmphasisSequence, "*", "_", style)
	},
}

var md050 = rule.Rule{
	Names:       []string{"MD050", "strong-style"},
	Description: "Strong style",
	Tags:        []string{"emphasis"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD050

		style := types.StringOr(c.Style, "consistent")
		if style == "" {
			style = "consistent"
		}

		md049Impl(p, onError, mm.TypeStrong, mm.TypeStrongSequence, "**", "__", style)
	},
}
