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

func init() { register(&md058) }

var md058 = rule.Rule{
	Names:       []string{"MD058", "blanks-around-tables"},
	Description: "Tables should be surrounded by blank lines",
	Tags:        []string{"table"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		lines := p.Lines
		blockQuotePrefixes := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeBlockQuotePrefix, mm.TypeLinePrefix},
			false,
		)

		tables := p.FilterByTypesCached([]mm.TokenType{mm.TypeTable}, false)
		for _, table := range tables {
			// Look for a blank line above the table.
			firstLineNumber := table.StartLine
			if firstLineNumber-2 >= 0 && firstLineNumber-2 < len(lines) &&
				!helpers.IsBlankLine(lines[firstLineNumber-2]) {
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

			// Look for a blank line below the table.
			lastLineNumber := table.EndLine
			if lastLineNumber >= 0 && lastLineNumber < len(lines) &&
				!helpers.IsBlankLine(lines[lastLineNumber]) {
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
