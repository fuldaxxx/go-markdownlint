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
	"strings"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md032) }

var md032 = rule.Rule{
	Names:       []string{"MD032", "blanks-around-lists"},
	Description: "Lists should be surrounded by blank lines",
	Tags:        []string{"bullet", "ul", "ol", "blank_lines"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		isList := func(token *mm.Token) bool {
			return token.Type == mm.TypeListOrdered || token.Type == mm.TypeListUnordered
		}
		lines := p.Lines
		blockQuotePrefixes := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeBlockQuotePrefix, mm.TypeLinePrefix},
			false,
		)

		// isBlankLine that treats out-of-range lines as blank.
		blankAt := func(index int) bool {
			if index < 0 || index >= len(lines) {
				return true
			}

			return helpers.IsBlankLine(lines[index])
		}

		// For every top-level list...
		topLevelLists := mdhelpers.FilterByPredicate(
			p.MicromarkTokens(),
			isList,
			func(token *mm.Token) []*mm.Token {
				if isList(token) || token.Type == mm.TypeHTMLFlow {
					return []*mm.Token{}
				}

				return token.Children
			},
		)
		for _, list := range topLevelLists {
			// Look for a blank line above the list
			firstLineNumber := list.StartLine
			if !blankAt(firstLineNumber - 2) {
				helpers.AddErrorContext(
					onError,
					firstLineNumber,
					strings.TrimSpace(lines[firstLineNumber-1]),
					false,
					false,
					nil,
					&types.FixInfo{
						InsertText: mdhelpers.GetBlockQuotePrefixText(
							blockQuotePrefixes,
							firstLineNumber,
							1,
						),
					},
				)
			}

			// Find the "visual" end of the list
			flattenedChildren := mdhelpers.FilterByPredicate(
				list.Children,
				func(token *mm.Token) bool {
					return !mdhelpers.NonContentTokens[token.Type]
				},
				func(token *mm.Token) []*mm.Token {
					if mdhelpers.NonContentTokens[token.Type] {
						return []*mm.Token{}
					}

					return token.Children
				},
			)

			endLine := list.EndLine
			if len(flattenedChildren) > 0 {
				endLine = flattenedChildren[len(flattenedChildren)-1].EndLine
			}

			// Look for a blank line below the list
			lastLineNumber := endLine
			if !blankAt(lastLineNumber) {
				helpers.AddErrorContext(
					onError,
					lastLineNumber,
					strings.TrimSpace(lines[lastLineNumber-1]),
					false,
					false,
					nil,
					&types.FixInfo{
						LineNumber: lastLineNumber + 1,
						InsertText: mdhelpers.GetBlockQuotePrefixText(
							blockQuotePrefixes,
							lastLineNumber,
							1,
						),
					},
				)
			}
		}
	},
}
