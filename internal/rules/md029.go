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
	"strconv"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md029) }

var md029ListStyleExamples = map[string]string{
	"one":     "1/1/1",
	"ordered": "1/2/3",
	"zero":    "0/0/0",
}

var md029 = rule.Rule{
	Names:       []string{"MD029", "ol-prefix"},
	Description: "Ordered list item prefix",
	Tags:        []string{"ol"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		getOrderedListItemValue := func(listItemPrefix *mm.Token) (int, int) {
			listItemValue := mdhelpers.GetDescendantsByType(
				[]*mm.Token{listItemPrefix},
				[][]mm.TokenType{{mm.TypeListItemValue}},
			)[0]
			value, _ := strconv.Atoi(listItemValue.Text)

			return listItemValue.StartColumn, value
		}

		c := p.Config.MD029
		style := types.StringOr(c.Style, "")

		for _, listOrdered := range p.FilterByTypesCached([]mm.TokenType{mm.TypeListOrdered}, false) {
			listItemPrefixes := mdhelpers.GetDescendantsByType(
				[]*mm.Token{listOrdered},
				[][]mm.TokenType{{mm.TypeListItemPrefix}},
			)
			expected := 1
			incrementing := false
			// Check for incrementing number pattern 1/2/3 or 0/1/2
			if len(listItemPrefixes) >= 2 {
				_, firstValue := getOrderedListItemValue(listItemPrefixes[0])

				_, secondValue := getOrderedListItemValue(listItemPrefixes[1])
				if secondValue != 1 || firstValue == 0 {
					incrementing = true

					if firstValue == 0 {
						expected = 0
					}
				}
			}
			// Determine effective style
			var listStyle string
			if style == "one" || style == "ordered" || style == "zero" {
				listStyle = style
			} else if incrementing {
				listStyle = "ordered"
			} else {
				listStyle = "one"
			}

			switch listStyle {
			case "zero":
				expected = 0
			case "one":
				expected = 1
			}
			// Validate each list item marker
			for _, listItemPrefix := range listItemPrefixes {
				column, actual := getOrderedListItemValue(listItemPrefix)
				fixInfo := fixReplace(column, runeLen(strconv.Itoa(actual)), strconv.Itoa(expected))
				helpers.AddErrorDetailIf(
					onError,
					listItemPrefix.StartLine,
					expected,
					actual,
					"Style: "+md029ListStyleExamples[listStyle],
					"",
					rng(
						listItemPrefix.StartColumn,
						listItemPrefix.EndColumn-listItemPrefix.StartColumn,
					),
					fixInfo,
				)

				if listStyle == "ordered" {
					expected++
				}
			}
		}
	},
}
