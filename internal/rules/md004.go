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
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md004) }

var md004 = rule.Rule{
	Names:       []string{"MD004", "ul-style"},
	Description: "Unordered list style",
	Tags:        []string{"bullet", "ul"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		markerToStyle := func(marker string) string {
			if marker == "-" {
				return "dash"
			}

			if marker == "+" {
				return "plus"
			}

			return "asterisk"
		}
		styleToMarker := func(style string) string {
			if style == "dash" {
				return "-"
			}

			if style == "plus" {
				return "+"
			}

			return "*"
		}
		differentItemStyle := func(style string) string {
			if style == "dash" {
				return "plus"
			}

			if style == "plus" {
				return "asterisk"
			}

			return "dash"
		}
		validStyles := map[string]bool{
			"asterisk":   true,
			"consistent": true,
			"dash":       true,
			"plus":       true,
			"sublist":    true,
		}

		c := p.Config.MD004
		style := types.StringOr(c.Style, "consistent")

		expectedStyle := "dash"
		if validStyles[style] {
			expectedStyle = style
		}

		nestingStyles := map[int]string{}

		for _, listUnordered := range p.FilterByTypesCached([]mm.TokenType{mm.TypeListUnordered}, false) {
			nesting := 0

			if style == "sublist" {
				parent := listUnordered
				for {
					parent = mdhelpers.GetParentOfType(
						parent,
						[]mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered},
					)
					if parent == nil {
						break
					}

					nesting++
				}
			}

			listItemMarkers := mdhelpers.GetDescendantsByType(
				[]*mm.Token{listUnordered},
				[][]mm.TokenType{
					{mm.TypeListItemPrefix},
					{mm.TypeListItemMarker},
				},
			)
			for _, listItemMarker := range listItemMarkers {
				itemStyle := markerToStyle(listItemMarker.Text)

				if style == "sublist" {
					if nestingStyles[nesting] == "" {
						if itemStyle == nestingStyles[nesting-1] {
							nestingStyles[nesting] = differentItemStyle(itemStyle)
						} else {
							nestingStyles[nesting] = itemStyle
						}
					}

					expectedStyle = nestingStyles[nesting]
				} else if expectedStyle == "consistent" {
					expectedStyle = itemStyle
				}

				column := listItemMarker.StartColumn
				length := listItemMarker.EndColumn - listItemMarker.StartColumn
				helpers.AddErrorDetailIf(
					onError,
					listItemMarker.StartLine,
					expectedStyle,
					itemStyle,
					"",
					"",
					rng(column, length),
					fixReplace(column, length, styleToMarker(expectedStyle)),
				)
			}
		}
	},
}
