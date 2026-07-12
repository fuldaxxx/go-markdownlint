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
	"github.com/ldmonster/go-markdownlint/internal/helpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md028) }

var md028IgnoreTypes = map[mm.TokenType]bool{
	mm.TypeLineEnding:     true,
	mm.TypeListItemIndent: true,
	mm.TypeLinePrefix:     true,
}

var md028 = rule.Rule{
	Names:       []string{"MD028", "no-blanks-blockquote"},
	Description: "Blank line inside blockquote",
	Tags:        []string{"blockquote", "whitespace"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		topTokens := p.MicromarkTokens()
		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeBlockQuote}, false) {
			var errorLineNumbers []int

			siblings := topTokens
			if token.Parent != nil {
				siblings = token.Parent.Children
			}

			idx := -1

			for i, s := range siblings {
				if s == token {
					idx = i
					break
				}
			}

			for i := idx + 1; i < len(siblings); i++ {
				sibling := siblings[i]
				startLine := sibling.StartLine

				typ := sibling.Type
				if typ == mm.TypeLineEndingBlank {
					// Possible blank between blockquotes
					errorLineNumbers = append(errorLineNumbers, startLine)
				} else if md028IgnoreTypes[typ] {
					// Ignore invisible formatting
				} else if typ == mm.TypeBlockQuote {
					// Blockquote followed by blockquote
					for _, lineNumber := range errorLineNumbers {
						helpers.AddError(onError, lineNumber, "", "", nil, nil)
					}

					break
				} else {
					// Blockquote not followed by blockquote
					break
				}
			}
		}
	},
}
