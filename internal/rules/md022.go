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

const md022DefaultLines = 1

// md022GetLinesFunction returns a function mapping a heading to its required
// blank-line count. The config value may be a number or an array of up to 6
// per-level values.
func md022GetLinesFunction(v interface{}) func(*mm.Token) int {
	if v != nil {
		if arr := md022AsArray(v); arr != nil {
			linesArray := [6]int{}
			for i := range linesArray {
				linesArray[i] = md022DefaultLines
			}

			for index, value := range arr {
				if index >= 6 {
					break
				}

				linesArray[index] = md022CoerceNumber(value)
			}

			return func(heading *mm.Token) int {
				return linesArray[mdhelpers.GetHeadingLevel(heading)-1]
			}
		}
	}

	lines := md022DefaultLines
	if v != nil {
		lines = md022CoerceNumber(v)
	}

	return func(*mm.Token) int { return lines }
}

func md022AsArray(v interface{}) []interface{} {
	switch t := v.(type) {
	case []interface{}:
		return t
	case []int:
		out := make([]interface{}, len(t))
		for i, e := range t {
			out[i] = e
		}

		return out
	case []float64:
		out := make([]interface{}, len(t))
		for i, e := range t {
			out[i] = e
		}

		return out
	default:
		return nil
	}
}

func md022CoerceNumber(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case bool:
		if t {
			return 1
		}

		return 0
	default:
		return md022DefaultLines
	}
}

var md022 = rule.Rule{
	Names:       []string{"MD022", "blanks-around-headings"},
	Description: "Headings should be surrounded by blank lines",
	Tags:        []string{"headings", "blank_lines"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		getLinesAbove := md022GetLinesFunction(p.Config.MD022.LinesAbove)
		getLinesBelow := md022GetLinesFunction(p.Config.MD022.LinesBelow)
		lines := p.Lines
		blockQuotePrefixes := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeBlockQuotePrefix, mm.TypeLinePrefix},
			false,
		)
		blankLine := func(idx int) bool {
			// Out-of-bounds indices are treated as blank lines.
			if idx < 0 || idx >= len(lines) {
				return true
			}

			return helpers.IsBlankLine(lines[idx])
		}

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			startLine := heading.StartLine
			endLine := heading.EndLine
			line := strings.TrimSpace(lines[startLine-1])

			linesAbove := getLinesAbove(heading)
			if linesAbove >= 0 {
				actualAbove := 0
				for i := 0; i < linesAbove && blankLine(startLine-2-i); i++ {
					actualAbove++
				}

				helpers.AddErrorDetailIf(
					onError,
					startLine,
					linesAbove,
					actualAbove,
					"Above",
					line,
					nil,
					&types.FixInfo{
						InsertText: mdhelpers.GetBlockQuotePrefixText(
							blockQuotePrefixes,
							startLine-1,
							linesAbove-actualAbove,
						),
					},
				)
			}

			linesBelow := getLinesBelow(heading)
			if linesBelow >= 0 {
				actualBelow := 0
				for i := 0; i < linesBelow && blankLine(endLine+i); i++ {
					actualBelow++
				}

				helpers.AddErrorDetailIf(
					onError,
					startLine,
					linesBelow,
					actualBelow,
					"Below",
					line,
					nil,
					&types.FixInfo{
						LineNumber: endLine + 1,
						InsertText: mdhelpers.GetBlockQuotePrefixText(
							blockQuotePrefixes,
							endLine+1,
							linesBelow-actualBelow,
						),
					},
				)
			}
		}
	},
}

func init() { register(&md022) }
