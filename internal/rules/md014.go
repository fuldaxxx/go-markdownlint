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

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md014) }

var md014DollarCommandRe = regexp.MustCompile(`^(\s*)(\$\s+)`)

var md014 = rule.Rule{
	Names:       []string{"MD014", "commands-show-output"},
	Description: "Dollar signs used before commands without showing output",
	Tags:        []string{"code"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		type dollarMatch struct {
			g1Len       int
			g2Len       int
			startColumn int
			startLine   int
			text        string
		}

		for _, codeBlock := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented}, false) {
			var codeFlowValues []*mm.Token

			for _, child := range codeBlock.Children {
				if child.Type == mm.TypeCodeFlowValue {
					codeFlowValues = append(codeFlowValues, child)
				}
			}

			var dollarMatches []dollarMatch

			for _, cfv := range codeFlowValues {
				m := md014DollarCommandRe.FindStringSubmatch(cfv.Text)
				if m == nil {
					continue
				}

				dollarMatches = append(dollarMatches, dollarMatch{
					g1Len:       runeLen(m[1]),
					g2Len:       runeLen(m[2]),
					startColumn: cfv.StartColumn,
					startLine:   cfv.StartLine,
					text:        cfv.Text,
				})
			}

			if len(dollarMatches) == len(codeFlowValues) {
				for _, dm := range dollarMatches {
					column := dm.startColumn + dm.g1Len
					length := dm.g2Len
					helpers.AddErrorContext(
						onError,
						dm.startLine,
						dm.text,
						false,
						false,
						rng(column, length),
						fixDelete(column, length),
					)
				}
			}
		}
	},
}
