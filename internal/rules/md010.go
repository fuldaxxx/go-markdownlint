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
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md010) }

var md010TabRe = regexp.MustCompile(`\t+`)

var md010 = rule.Rule{
	Names:       []string{"MD010", "no-hard-tabs"},
	Description: "Hard tabs",
	Tags:        []string{"whitespace", "hard_tab"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD010
		includeCode := types.BoolOr(c.CodeBlocks, true)

		ignoreCodeLanguages := map[string]struct{}{}
		for _, lang := range c.IgnoreCodeLanguages {
			ignoreCodeLanguages[strings.ToLower(lang)] = struct{}{}
		}

		spaceMultiplier := 1
		if c.SpacesPerTab != nil {
			spaceMultiplier = types.IntOr(c.SpacesPerTab, 1)
			if spaceMultiplier < 0 {
				spaceMultiplier = 0
			}
		}

		var exclusionTypes []mm.TokenType

		if includeCode {
			if len(ignoreCodeLanguages) > 0 {
				exclusionTypes = append(exclusionTypes, mm.TypeCodeFenced)
			}
		} else {
			exclusionTypes = append(
				exclusionTypes,
				mm.TypeCodeFenced,
				mm.TypeCodeIndented,
				mm.TypeCodeText,
			)
		}

		var codeTokens []*mm.Token

		for _, token := range p.FilterByTypesCached(exclusionTypes, false) {
			if token.Type == mm.TypeCodeFenced && len(ignoreCodeLanguages) > 0 {
				fenceInfos := mdhelpers.GetDescendantsByType([]*mm.Token{token}, [][]mm.TokenType{
					{mm.TypeCodeFencedFence},
					{mm.TypeCodeFencedFenceInfo},
				})
				every := true

				for _, fenceInfo := range fenceInfos {
					if _, ok := ignoreCodeLanguages[strings.ToLower(fenceInfo.Text)]; !ok {
						every = false
						break
					}
				}

				if !every {
					continue
				}
			}

			codeTokens = append(codeTokens, token)
		}

		type codeRange struct {
			startLine, startColumn, endLine, endColumn int
		}

		codeRanges := make([]codeRange, 0, len(codeTokens))
		for _, token := range codeTokens {
			codeFenced := token.Type == mm.TypeCodeFenced

			cr := codeRange{
				startLine:   token.StartLine,
				startColumn: token.StartColumn,
				endLine:     token.EndLine,
				endColumn:   token.EndColumn,
			}
			if codeFenced {
				cr.startLine = token.StartLine + 1
				cr.startColumn = 0
				cr.endLine = token.EndLine - 1
				cr.endColumn = math.MaxInt32
			}

			codeRanges = append(codeRanges, cr)
		}

		for lineIndex, line := range p.Lines {
			for _, m := range md010TabRe.FindAllStringIndex(line, -1) {
				// Convert byte offsets to rune column.
				column := runeLen(line[:m[0]]) + 1
				length := runeLen(line[m[0]:m[1]])
				lineNumber := lineIndex + 1
				rangeFR := helpers.FileRange{
					StartLine:   lineNumber,
					StartColumn: column,
					EndLine:     lineNumber,
					EndColumn:   column + length - 1,
				}
				overlap := false

				for _, cr := range codeRanges {
					if helpers.HasOverlap(
						helpers.FileRange{
							StartLine:   cr.startLine,
							StartColumn: cr.startColumn,
							EndLine:     cr.endLine,
							EndColumn:   cr.endColumn,
						},
						rangeFR,
					) {
						overlap = true
						break
					}
				}

				if !overlap {
					helpers.AddError(
						onError,
						lineNumber,
						"Column: "+strconv.Itoa(column),
						"",
						rng(column, length),
						fixReplace(column, length, strings.Repeat(" ", length*spaceMultiplier)),
					)
				}
			}
		}
	},
}
