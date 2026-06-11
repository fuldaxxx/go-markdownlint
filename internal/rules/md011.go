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
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md011) }

// reversedLinkRe = /(^|[^\\])\(([^()]+)\)\[([^\]^][^\]]*)\](?!\()/g
// Go RE2 has no lookahead, so the (?!\() is handled in code.
var md011ReversedLinkRe = regexp.MustCompile(`(^|[^\\])\(([^()]+)\)\[([^\]^][^\]]*)\]`)

var md011 = rule.Rule{
	Names:       []string{"MD011", "no-reversed-links"},
	Description: "Reversed link syntax",
	Tags:        []string{"links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		ignoreBlockLineNumbers := map[int]struct{}{}
		for _, ignoreBlock := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented, mm.TypeMathFlow}, false) {
			mdhelpers.AddRangeToSet(
				ignoreBlockLineNumbers,
				ignoreBlock.StartLine,
				ignoreBlock.EndLine,
			)
		}

		ignoreTexts := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeCodeText, mm.TypeMathText},
			false,
		)
		// byteToRune maps a byte offset within s to a rune index.
		byteToRune := func(s string, byteOff int) int {
			return runeLen(s[:byteOff])
		}

		for lineIndex, line := range p.Lines {
			lineNumber := lineIndex + 1
			if _, skip := ignoreBlockLineNumbers[lineNumber]; skip {
				continue
			}

			matches := md011ReversedLinkRe.FindAllStringSubmatchIndex(line, -1)
			for _, loc := range matches {
				sub := func(g int) string {
					if loc[2*g] < 0 {
						return ""
					}

					return line[loc[2*g]:loc[2*g+1]]
				}
				reversedLink := sub(0)
				preChar := sub(1)
				linkText := sub(2)
				linkDestination := sub(3)

				// Manual (?!\() negative lookahead: skip if next char is '('.
				if loc[1] < len(line) && line[loc[1]] == '(' {
					continue
				}

				if strings.HasSuffix(linkText, "\\") || strings.HasSuffix(linkDestination, "\\") {
					continue
				}

				matchIndex := byteToRune(line, loc[0])
				column := matchIndex + runeLen(preChar) + 1
				length := runeLen(reversedLink) - runeLen(preChar)
				rangeFR := helpers.FileRange{
					StartLine:   lineNumber,
					StartColumn: column,
					EndLine:     lineNumber,
					EndColumn:   column + length - 1,
				}
				overlap := false

				for _, ignoreText := range ignoreTexts {
					if helpers.HasOverlap(
						helpers.FileRange{
							StartLine:   ignoreText.StartLine,
							StartColumn: ignoreText.StartColumn,
							EndLine:     ignoreText.EndLine,
							EndColumn:   ignoreText.EndColumn,
						},
						rangeFR,
					) {
						overlap = true
						break
					}
				}

				if overlap {
					continue
				}

				helpers.AddError(
					onError,
					lineNumber,
					string([]rune(reversedLink)[runeLen(preChar):]),
					"",
					rng(column, length),
					fixReplace(column, length, "["+linkText+"]("+linkDestination+")"),
				)
			}
		}
	},
}
