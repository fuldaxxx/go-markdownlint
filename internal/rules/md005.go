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
	"strings"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md005) }

var md005 = rule.Rule{
	Names:       []string{"MD005", "list-indent"},
	Description: "Inconsistent indentation for list items at the same level",
	Tags:        []string{"bullet", "ul", "indentation"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		for _, list := range p.FilterByTypesCached([]mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered}, false) {
			expectedIndent := list.StartColumn - 1
			expectedEnd := 0
			endMatching := false

			for _, listItemPrefix := range list.Children {
				if listItemPrefix.Type != mm.TypeListItemPrefix {
					continue
				}

				lineNumber := listItemPrefix.StartLine
				actualIndent := listItemPrefix.StartColumn - 1

				rangeVal := rng(1, listItemPrefix.EndColumn-1)
				if list.Type == mm.TypeListUnordered {
					helpers.AddErrorDetailIf(
						onError,
						lineNumber,
						expectedIndent,
						actualIndent,
						"",
						"",
						rangeVal,
						nil,
						// No fixInfo; MD007 handles this scenario better
					)
				} else {
					markerLength := runeLen(strings.TrimSpace(listItemPrefix.Text))

					actualEnd := listItemPrefix.StartColumn + markerLength - 1
					if expectedEnd == 0 {
						expectedEnd = actualEnd
					}

					if (expectedIndent != actualIndent) || endMatching {
						if expectedEnd == actualEnd {
							endMatching = true
						} else {
							var (
								detail           string
								expected, actual int
							)

							if endMatching {
								detail = "Expected: (" + strconv.Itoa(
									expectedEnd,
								) + "); Actual: (" + strconv.Itoa(
									actualEnd,
								) + ")"
								expected = expectedEnd - markerLength
								actual = actualEnd - markerLength
							} else {
								detail = "Expected: " + strconv.Itoa(
									expectedIndent,
								) + "; Actual: " + strconv.Itoa(
									actualIndent,
								)
								expected = expectedIndent
								actual = actualIndent
							}

							deleteCount := actual - expected
							if deleteCount < 0 {
								deleteCount = 0
							}

							insertCount := expected - actual
							if insertCount < 0 {
								insertCount = 0
							}

							editColumn := actual
							if expected < editColumn {
								editColumn = expected
							}

							helpers.AddError(
								onError,
								lineNumber,
								detail,
								"",
								rangeVal,
								fixReplace(
									editColumn+1,
									deleteCount,
									strings.Repeat(" ", insertCount),
								),
							)
						}
					}
				}
			}
		}
	},
}
