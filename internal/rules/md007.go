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

func init() { register(&md007) }

var md007UnorderedListTypes = []mm.TokenType{
	mm.TypeBlockQuotePrefix, mm.TypeListItemPrefix, mm.TypeListUnordered,
}

var md007UnorderedParentTypes = []mm.TokenType{
	mm.TypeBlockQuote, mm.TypeListOrdered, mm.TypeListUnordered,
}

var md007 = rule.Rule{
	Names:       []string{"MD007", "ul-indent"},
	Description: "Unordered list indentation",
	Tags:        []string{"bullet", "ul", "indentation"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD007
		indent := types.IntOr(c.Indent, 2)
		startIndented := types.BoolOr(c.StartIndented, false)
		startIndent := types.IntOr(c.StartIndent, indent)
		unorderedListNesting := map[*mm.Token]int{}

		var lastBlockQuotePrefix *mm.Token

		tokens := p.FilterByTypesCached(md007UnorderedListTypes, false)
		for _, token := range tokens {
			switch token.Type {
			case mm.TypeBlockQuotePrefix:
				lastBlockQuotePrefix = token
			case mm.TypeListUnordered:
				nesting := 0

				current := token
				for {
					current = mdhelpers.GetParentOfType(current, md007UnorderedParentTypes)
					if current == nil {
						break
					}

					if current.Type == mm.TypeListUnordered {
						nesting++
						continue
					} else if current.Type == mm.TypeListOrdered {
						nesting = -1
					}

					break
				}

				if nesting >= 0 {
					unorderedListNesting[token] = nesting
				}
			default:
				// listItemPrefix
				nesting, ok := unorderedListNesting[token.Parent]
				if ok {
					baseIndent := 0
					if mdhelpers.GetParentOfType(
						token,
						[]mm.TokenType{mm.TypeGfmFootnoteDefinition},
					) != nil {
						baseIndent = 4
					}

					expectedIndent := baseIndent + nesting*indent
					if startIndented {
						expectedIndent += startIndent
					}

					blockQuoteAdjustment := 0
					if lastBlockQuotePrefix != nil &&
						lastBlockQuotePrefix.EndLine == token.StartLine {
						blockQuoteAdjustment = lastBlockQuotePrefix.EndColumn - 1
					}

					actualIndent := token.StartColumn - 1 - blockQuoteAdjustment
					rangeVal := rng(1, token.EndColumn-1)

					deleteCount := actualIndent - expectedIndent
					if deleteCount < 0 {
						deleteCount = 0
					}

					insertCount := expectedIndent - actualIndent
					if insertCount < 0 {
						insertCount = 0
					}

					fixInfo := fixReplace(
						token.StartColumn-actualIndent,
						deleteCount,
						strings.Repeat(" ", insertCount),
					)
					helpers.AddErrorDetailIf(
						onError,
						token.StartLine,
						expectedIndent,
						actualIndent,
						"",
						"",
						rangeVal,
						fixInfo,
					)
				}
			}
		}
	},
}
