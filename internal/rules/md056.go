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
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md056) }

var md056 = rule.Rule{
	Names:       []string{"MD056", "table-column-count"},
	Description: "Table column count",
	Tags:        []string{"table"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		makeRange := func(start, end int) *[2]int { return rng(start, end-start+1) }

		// The Go parser emits header cells as tableData (not tableHeader) and
		// delimiter cells as tableDelimiter; cover both cell kinds.
		md056IsCell := func(t *mm.Token) bool {
			return t.Type == mm.TypeTableData ||
				t.Type == mm.TypeTableDelimiter ||
				t.Type == mm.TypeTableHeader
		}

		rows := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeTableDelimiterRow, mm.TypeTableRow},
			false,
		)
		expectedCount := 0

		var currentTable *mm.Token

		for _, row := range rows {
			table := mdhelpers.GetParentOfType(row, []mm.TokenType{mm.TypeTable})
			if currentTable != table {
				expectedCount = 0
				currentTable = table
			}

			var cells []*mm.Token

			for _, child := range row.Children {
				if md056IsCell(child) {
					cells = append(cells, child)
				}
			}

			actualCount := len(cells)
			if expectedCount == 0 {
				expectedCount = actualCount
			}

			detail := ""

			var rangeVal *[2]int

			if actualCount < expectedCount {
				detail = "Too few cells, row will be missing data"
				rangeVal = rng(row.EndColumn-1, 1)
			} else if expectedCount < actualCount {
				detail = "Too many cells, extra data will be missing"
				rangeVal = makeRange(cells[expectedCount].StartColumn, row.EndColumn-1)
			}

			helpers.AddErrorDetailIf(
				onError,
				row.EndLine,
				expectedCount,
				actualCount,
				detail,
				"",
				rangeVal,
				nil,
			)
		}
	},
}
