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

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md034) }

var md034 = rule.Rule{
	Names:       []string{"MD034", "no-bare-urls"},
	Description: "Bare URL used",
	Tags:        []string{"links", "url"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		allowed := func(token *mm.Token) bool {
			if token.Type == mm.TypeLiteralAutolink && !token.InHTMLFlow() {
				// Detect and ignore micromark/micromark#164.
				var siblings []*mm.Token
				if token.Parent != nil {
					siblings = token.Parent.Children
				}

				index := -1

				for i, s := range siblings {
					if s == token {
						index = i
						break
					}
				}

				at := func(i int) *mm.Token {
					// Negative indices are supported (counting from the end).
					n := len(siblings)
					if i < 0 {
						i += n
					}

					if i < 0 || i >= n {
						return nil
					}

					return siblings[i]
				}
				prev := at(index - 1)
				next := at(index + 1)

				return prev == nil ||
					next == nil ||
					prev.Type != mm.TypeData ||
					next.Type != mm.TypeData || !strings.HasSuffix(prev.Text, "<") || !strings.HasPrefix(next.Text, ">")
			}

			return false
		}

		transform := func(token *mm.Token) []*mm.Token {
			children := token.Children

			var result []*mm.Token

			for i := 0; i < len(children); i++ {
				current := children[i]

				openTagInfo := mdhelpers.GetHTMLTagInfo(current)
				if openTagInfo != nil && !openTagInfo.Close {
					count := 1

					for j := i + 1; j < len(children); j++ {
						candidate := children[j]

						closeTagInfo := mdhelpers.GetHTMLTagInfo(candidate)
						if closeTagInfo != nil && openTagInfo.Name == closeTagInfo.Name {
							if closeTagInfo.Close {
								count--
								if count == 0 {
									i = j
									break
								}
							} else {
								count++
							}
						}
					}
				} else {
					result = append(result, current)
				}
			}

			return result
		}
		for _, token := range mdhelpers.FilterByPredicate(p.MicromarkTokens(), allowed, transform) {
			column := token.StartColumn
			length := token.EndColumn - token.StartColumn
			fixInfo := fixReplace(column, length, "<"+token.Text+">")
			helpers.AddErrorContext(
				onError,
				token.StartLine,
				token.Text,
				false,
				false,
				rng(column, length),
				fixInfo,
			)
		}
	},
}
