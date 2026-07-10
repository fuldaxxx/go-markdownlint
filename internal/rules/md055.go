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
	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md055) }

var md055 = rule.Rule{
	Names:       []string{"MD055", "table-pipe-style"},
	Description: "Table pipe style",
	Tags:        []string{"table"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		// makeRange => [start, end - start + 1]
		makeRange := func(start, end int) *[2]int { return rng(start, end-start+1) }

		md055IsWhitespace := func(t *mm.Token) bool {
			return t.Type == mm.TypeLinePrefix || t.Type == mm.TypeWhitespace
		}

		c := p.Config.MD055
		style := types.StringOr(c.Style, "consistent")
		expectedStyle := style
		expectedLeadingPipe := (expectedStyle != "no_leading_or_trailing") &&
			(expectedStyle != "trailing_only")
		expectedTrailingPipe := (expectedStyle != "no_leading_or_trailing") &&
			(expectedStyle != "leading_only")

		rows := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeTableDelimiterRow, mm.TypeTableRow},
			false,
		)
		for _, row := range rows {
			// The Go parser emits cells (tableData/tableDelimiter) and dividers
			// (tableCellDivider) as direct children of the row, rather than
			// nesting the divider inside the cell. A
			// leading/trailing pipe is therefore a tableCellDivider at the
			// start/end of the (non-whitespace) children of the row.
			children := row.Children
			if len(children) == 0 {
				continue
			}

			// First non-whitespace child = the leading token of the first cell.
			var leadingToken *mm.Token

			for _, c := range children {
				if !md055IsWhitespace(c) {
					leadingToken = c
					break
				}
			}
			// Last non-whitespace child = the trailing token of the last cell.
			var trailingToken *mm.Token

			for i := len(children) - 1; i >= 0; i-- {
				if !md055IsWhitespace(children[i]) {
					trailingToken = children[i]
					break
				}
			}

			if leadingToken == nil || trailingToken == nil {
				continue
			}

			// firstCell/lastCell: the first/last non-whitespace child overall.
			firstCell := leadingToken
			lastCell := trailingToken

			actualLeadingPipe := leadingToken.Type == mm.TypeTableCellDivider
			actualTrailingPipe := trailingToken.Type == mm.TypeTableCellDivider

			var actualStyle string

			if actualLeadingPipe {
				if actualTrailingPipe {
					actualStyle = "leading_and_trailing"
				} else {
					actualStyle = "leading_only"
				}
			} else if actualTrailingPipe {
				actualStyle = "trailing_only"
			} else {
				actualStyle = "no_leading_or_trailing"
			}

			if expectedStyle == "consistent" {
				expectedStyle = actualStyle
				expectedLeadingPipe = actualLeadingPipe
				expectedTrailingPipe = actualTrailingPipe
			}

			if actualLeadingPipe != expectedLeadingPipe {
				detail := "Unexpected leading pipe"
				if expectedLeadingPipe {
					detail = "Missing leading pipe"
				}

				helpers.AddErrorDetailIf(
					onError,
					firstCell.StartLine,
					expectedStyle,
					actualStyle,
					detail,
					"",
					makeRange(row.StartColumn, firstCell.StartColumn),
					nil,
				)
			}

			if actualTrailingPipe != expectedTrailingPipe {
				detail := "Unexpected trailing pipe"
				if expectedTrailingPipe {
					detail = "Missing trailing pipe"
				}

				helpers.AddErrorDetailIf(
					onError,
					lastCell.EndLine,
					expectedStyle,
					actualStyle,
					detail,
					"",
					makeRange(lastCell.EndColumn-1, row.EndColumn-1),
					nil,
				)
			}
		}
	},
}
