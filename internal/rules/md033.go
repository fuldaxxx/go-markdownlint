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

func init() { register(&md033) }

var md033 = rule.Rule{
	Names:       []string{"MD033", "no-inline-html"},
	Description: "Inline HTML",
	Tags:        []string{"html"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD033
		toLowerCaseStringArray := func(arr []string) []string {
			out := make([]string, 0, len(arr))
			for _, e := range arr {
				out = append(out, strings.ToLower(e))
			}

			return out
		}
		contains := func(slice []string, v string) bool {
			for _, s := range slice {
				if s == v {
					return true
				}
			}

			return false
		}
		allowedElements := toLowerCaseStringArray(c.AllowedElements)
		// If not defined, use allowedElements for backward compatibility
		var tableAllowedElements []string
		if c.TableAllowedElements != nil {
			tableAllowedElements = toLowerCaseStringArray(*c.TableAllowedElements)
		} else {
			tableAllowedElements = toLowerCaseStringArray(c.AllowedElements)
		}

		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeHTMLText}, true) {
			htmlTagInfo := mdhelpers.GetHTMLTagInfo(token)
			if htmlTagInfo != nil && !htmlTagInfo.Close {
				elementName := strings.ToLower(htmlTagInfo.Name)

				inTable := mdhelpers.GetParentOfType(token, []mm.TokenType{mm.TypeTable}) != nil
				if (inTable || !contains(allowedElements, elementName)) &&
					(!inTable || !contains(tableAllowedElements, elementName)) {
					length := runeLen(helpers.NextLinesRe.ReplaceAllString(token.Text, ""))
					helpers.AddError(
						onError,
						token.StartLine,
						"Element: "+htmlTagInfo.Name,
						"",
						rng(token.StartColumn, length),
						nil,
					)
				}
			}
		}
	},
}
