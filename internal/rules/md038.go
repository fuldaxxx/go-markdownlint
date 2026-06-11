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

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md038) }

var (
	md038StartRe = regexp.MustCompile(`^(\s+)(\S)`)
	md038EndRe   = regexp.MustCompile(`(\S)(\s+)$`)
)

var md038 = rule.Rule{
	Names:       []string{"MD038", "no-space-in-code"},
	Description: "Spaces inside code span elements",
	Tags:        []string{"whitespace", "code"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		codeTexts := p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeText}, false)
		for _, codeText := range codeTexts {
			datas := mdhelpers.GetDescendantsByType(
				[]*mm.Token{codeText},
				[][]mm.TokenType{{mm.TypeCodeTextData}},
			)
			if len(datas) > 0 {
				paddings := mdhelpers.GetDescendantsByType(
					[]*mm.Token{codeText},
					[][]mm.TokenType{{mm.TypeCodeTextPadding}},
				)

				// Check for extra space at start of code.
				var startPadding *mm.Token
				if len(paddings) > 0 {
					startPadding = paddings[0]
				}

				startData := datas[0]
				startMatch := md038StartRe.FindStringSubmatch(startData.Text)

				startGroup1, startGroup2 := "", ""
				if startMatch != nil {
					startGroup1 = startMatch[1]
					startGroup2 = startMatch[2]
				}

				startBacktick := startGroup2 == "`"

				startCount := runeLen(startGroup1)
				if startBacktick && startPadding == nil {
					startCount--
				}

				startSpaces := startCount > 0

				// Check for extra space at end of code.
				var endPadding *mm.Token
				if len(paddings) > 0 {
					endPadding = paddings[len(paddings)-1]
				}

				endData := datas[len(datas)-1]
				endMatch := md038EndRe.FindStringSubmatch(endData.Text)

				endGroup1, endGroup2 := "", ""
				if endMatch != nil {
					endGroup1 = endMatch[1]
					endGroup2 = endMatch[2]
				}

				endBacktick := endGroup1 == "`"

				endCount := runeLen(endGroup2)
				if endBacktick && endPadding == nil {
					endCount--
				}

				endSpaces := endCount > 0

				// Check if safe to remove 1-space padding.
				removePadding := startSpaces && endSpaces && startPadding != nil &&
					endPadding != nil &&
					!startBacktick &&
					!endBacktick
				context := codeText.Text

				// If extra space at start, report violation.
				if startSpaces {
					var startColumn int
					if removePadding {
						startColumn = startPadding.StartColumn
					} else {
						startColumn = startData.StartColumn
					}

					length := startCount
					if removePadding {
						length += runeLen(startPadding.Text)
					}

					helpers.AddErrorContext(
						onError,
						startData.StartLine,
						context,
						true,
						false,
						rng(startColumn, length),
						fixDelete(startColumn, length),
					)
				}

				// If extra space at end, report violation.
				if endSpaces {
					var endColumn int
					if removePadding {
						endColumn = endPadding.EndColumn
					} else {
						endColumn = endData.EndColumn
					}

					length := endCount
					if removePadding {
						length += runeLen(endPadding.Text)
					}

					helpers.AddErrorContext(
						onError,
						endData.EndLine,
						context,
						false,
						true,
						rng(endColumn-length, length),
						fixDelete(endColumn-length, length),
					)
				}
			}
		}
	},
}
