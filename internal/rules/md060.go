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
	"unicode"

	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md060) }

// md060StringWidth approximates terminal string width: the
// visual width of a string, counting East-Asian wide / fullwidth characters as
// 2 columns and zero-width / combining marks as 0. This is used only to compute
// the "effective" alignment column of table pipe dividers.
func md060StringWidth(s string) int {
	width := 0

	for _, r := range s {
		if r == 0 {
			continue
		}
		// Combining marks and other zero-width characters.
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) {
			continue
		}

		if md060IsWide(r) {
			width += 2
		} else {
			width++
		}
	}

	return width
}

// md060IsWide reports whether r is an East-Asian wide or fullwidth code point.
func md060IsWide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115F) || // Hangul Jamo
		r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0x303E) || // CJK Radicals .. Kangxi
		(r >= 0x3041 && r <= 0x33FF) || // Hiragana .. CJK symbols
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Ext A
		(r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified
		(r >= 0xA000 && r <= 0xA4CF) || // Yi
		(r >= 0xAC00 && r <= 0xD7A3) || // Hangul Syllables
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		(r >= 0xFE10 && r <= 0xFE19) || // Vertical forms
		(r >= 0xFE30 && r <= 0xFE6F) || // CJK Compatibility Forms
		(r >= 0xFF00 && r <= 0xFF60) || // Fullwidth Forms
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x1F300 && r <= 0x1FAFF) || // Emoji
		(r >= 0x20000 && r <= 0x3FFFD) // CJK Ext B+
}

type md060errorInfo struct {
	lineNumber int
	column     int
	detail     string
}

// md060Column holds a divider's actual (1-based rune) column and its effective
// (visual width) column.
type md060Column struct {
	actual    int
	effective int
}

var md060 = rule.Rule{
	Names:       []string{"MD060", "table-column-style"},
	Description: "Table column style",
	Tags:        []string{"table"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD060
		lines := p.Lines

		// addError analog: pushes an error info with range [column, 1].
		addError := func(errs *[]md060errorInfo, lineNumber, column int, detail string) {
			*errs = append(
				*errs,
				md060errorInfo{lineNumber: lineNumber, column: column, detail: detail},
			)
		}

		// getTableDividerColumns: actual = divider rune column; effective = visual
		// width of the line text preceding the divider.
		getTableDividerColumns := func(row *mm.Token) []md060Column {
			cols := make([]md060Column, 0, len(row.Children))

			var line string
			if row.StartLine-1 >= 0 && row.StartLine-1 < len(lines) {
				line = lines[row.StartLine-1]
			}

			lineRunes := []rune(line)

			for _, divider := range mdhelpers.FilterByTypes(row.Children, []mm.TokenType{mm.TypeTableCellDivider}, false) {
				prefixLen := divider.StartColumn - 1

				prefix := ""
				if prefixLen >= 0 && prefixLen <= len(lineRunes) {
					prefix = string(lineRunes[:prefixLen])
				}

				cols = append(cols, md060Column{
					actual:    divider.StartColumn,
					effective: md060StringWidth(prefix),
				})
			}

			return cols
		}

		// checkStyleAligned: pipes in each non-header row must align (by effective
		// column) with a pipe in the header row.
		checkStyleAligned := func(rows []*mm.Token, detail string) []md060errorInfo {
			var errorInfos []md060errorInfo
			if len(rows) == 0 {
				return errorInfos
			}

			headerRow := rows[0]

			headerDividerColumns := getTableDividerColumns(headerRow)
			for _, row := range rows[1:] {
				// Set of header effective columns (deduplicated).
				remaining := make(map[int]struct{}, len(headerDividerColumns))
				for _, c := range headerDividerColumns {
					remaining[c.effective] = struct{}{}
				}

				rowDividerColumns := getTableDividerColumns(row)
				for _, dividerColumn := range rowDividerColumns {
					if len(remaining) > 0 {
						if _, ok := remaining[dividerColumn.effective]; ok {
							delete(remaining, dividerColumn.effective)
						} else {
							addError(&errorInfos, row.StartLine, dividerColumn.actual, detail)
						}
					}
				}
			}

			return errorInfos
		}

		style := types.StringOr(c.Style, "any")
		styleAlignedAllowed := style == "any" || style == "aligned"
		styleCompactAllowed := style == "any" || style == "compact"
		styleTightAllowed := style == "any" || style == "tight"
		alignedDelimiter := types.BoolOr(c.AlignedDelimiter, false)

		tables := p.FilterByTypesCached([]mm.TokenType{mm.TypeTable}, false)
		for _, table := range tables {
			rows := mdhelpers.FilterByTypes(
				table.Children,
				[]mm.TokenType{mm.TypeTableDelimiterRow, mm.TypeTableRow},
				false,
			)

			// Errors for style "aligned".
			var errorsIfAligned []md060errorInfo
			if styleAlignedAllowed {
				errorsIfAligned = append(
					errorsIfAligned,
					checkStyleAligned(
						rows,
						"Table pipe does not align with header for style \"aligned\"",
					)...)
			}

			// Errors for styles "compact" and "tight".
			var (
				errorsIfCompact []md060errorInfo
				errorsIfTight   []md060errorInfo
			)

			if (styleCompactAllowed || styleTightAllowed) &&
				(!styleAlignedAllowed || len(errorsIfAligned) != 0) {
				if alignedDelimiter {
					upTo := rows
					if len(upTo) > 2 {
						upTo = rows[:2]
					}

					errorInfos := checkStyleAligned(
						upTo,
						"Table pipe does not align with header for option \"aligned_delimiter\"",
					)
					errorsIfCompact = append(errorsIfCompact, errorInfos...)
					errorsIfTight = append(errorsIfTight, errorInfos...)
				}

				for _, row := range rows {
					tokensOfInterest := md060RowTokens(row)
					for i := 0; i < len(tokensOfInterest); i++ {
						tok := tokensOfInterest[i]
						if tok.Type != mm.TypeTableCellDivider {
							continue
						}

						startColumn := tok.StartColumn
						startLine := tok.StartLine

						if i-1 >= 0 {
							previous := tokensOfInterest[i-1]
							if previous.Type == mm.TypeWhitespace {
								if runeLen(previous.Text) != 1 {
									addError(
										&errorsIfCompact,
										startLine,
										startColumn,
										"Table pipe has extra space to the left for style \"compact\"",
									)
								}

								addError(
									&errorsIfTight,
									startLine,
									startColumn,
									"Table pipe has space to the left for style \"tight\"",
								)
							} else {
								addError(
									&errorsIfCompact,
									startLine,
									startColumn,
									"Table pipe is missing space to the left for style \"compact\"",
								)
							}
						}

						if i+1 < len(tokensOfInterest) {
							next := tokensOfInterest[i+1]
							if next.Type == mm.TypeWhitespace {
								if next.EndColumn != row.EndColumn {
									if runeLen(next.Text) != 1 {
										addError(
											&errorsIfCompact,
											startLine,
											startColumn,
											"Table pipe has extra space to the right for style \"compact\"",
										)
									}

									addError(
										&errorsIfTight,
										startLine,
										startColumn,
										"Table pipe has space to the right for style \"tight\"",
									)
								}
							} else {
								addError(
									&errorsIfCompact,
									startLine,
									startColumn,
									"Table pipe is missing space to the right for style \"compact\"",
								)
							}
						}
					}
				}
			}

			// Report errors for whatever (allowed) style has the fewest.
			errorInfos := errorsIfAligned
			if styleCompactAllowed &&
				(len(errorsIfCompact) < len(errorInfos) || !styleAlignedAllowed) {
				errorInfos = errorsIfCompact
			}

			if styleTightAllowed &&
				(len(errorsIfTight) < len(errorInfos) || (!styleAlignedAllowed && !styleCompactAllowed)) {
				errorInfos = errorsIfTight
			}

			for _, ei := range errorInfos {
				onError(types.ErrorInfo{
					LineNumber: ei.lineNumber,
					Detail:     ei.detail,
					Range:      rng(ei.column, 1),
				})
			}
		}
	},
}

// md060RowTokens reconstructs the GFM-table token sequence of interest
// (tableCellDivider, tableContent, whitespace) for a row. The Go parser emits
// cells (tableData/tableDelimiter) and dividers (tableCellDivider) as row
// children, but does not split cells into content/whitespace sub-tokens, so
// each cell's text is decomposed here into optional leading whitespace, content,
// and optional trailing whitespace tokens with rune-based columns.
func md060RowTokens(row *mm.Token) []*mm.Token {
	var out []*mm.Token

	for _, child := range row.Children {
		switch child.Type {
		case mm.TypeTableCellDivider:
			out = append(out, child)
		case mm.TypeTableData, mm.TypeTableDelimiter, mm.TypeTableHeader, mm.TypeTableContent:
			runes := []rune(child.Text)
			n := len(runes)
			// Leading whitespace.
			lead := 0
			for lead < n && (runes[lead] == ' ' || runes[lead] == '\t') {
				lead++
			}
			// Trailing whitespace.
			trail := n
			for trail > lead && (runes[trail-1] == ' ' || runes[trail-1] == '\t') {
				trail--
			}

			startCol := child.StartColumn
			if lead > 0 {
				out = append(out, &mm.Token{
					Type:        mm.TypeWhitespace,
					StartLine:   child.StartLine,
					StartColumn: startCol,
					EndLine:     child.StartLine,
					EndColumn:   startCol + lead - 1,
					Text:        string(runes[:lead]),
				})
			}

			if trail > lead {
				out = append(out, &mm.Token{
					Type:        mm.TypeTableContent,
					StartLine:   child.StartLine,
					StartColumn: startCol + lead,
					EndLine:     child.StartLine,
					EndColumn:   startCol + trail - 1,
					Text:        string(runes[lead:trail]),
				})
			}

			if trail < n {
				out = append(out, &mm.Token{
					Type:        mm.TypeWhitespace,
					StartLine:   child.StartLine,
					StartColumn: startCol + trail,
					EndLine:     child.StartLine,
					EndColumn:   startCol + n - 1,
					Text:        string(runes[trail:]),
				})
			}
		}
	}

	return out
}
