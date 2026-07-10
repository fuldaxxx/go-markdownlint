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
	"unicode"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md009) }

var md009 = rule.Rule{
	Names:       []string{"MD009", "no-trailing-spaces"},
	Description: "Trailing spaces",
	Tags:        []string{"whitespace"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD009
		brSpaces := types.IntOr(c.BrSpaces, 2)
		includeCode := types.BoolOr(c.CodeBlocks, false)
		listItemEmptyLines := types.BoolOr(c.ListItemEmptyLines, false)
		strict := types.BoolOr(c.Strict, false)

		codeBlockLines := map[int]struct{}{}

		if !includeCode {
			for _, cb := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced}, false) {
				mdhelpers.AddRangeToSet(codeBlockLines, cb.StartLine+1, cb.EndLine-1)
			}

			for _, cb := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeIndented}, false) {
				mdhelpers.AddRangeToSet(codeBlockLines, cb.StartLine, cb.EndLine)
			}
		}

		listItemLines := map[int]struct{}{}

		if listItemEmptyLines {
			for _, lb := range p.FilterByTypesCached([]mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered}, false) {
				mdhelpers.AddRangeToSet(listItemLines, lb.StartLine, lb.EndLine)

				trailingIndent := true

				for i := len(lb.Children) - 1; i >= 0; i-- {
					child := lb.Children[i]
					switch child.Type {
					case mm.TypeContent:
						trailingIndent = false
					case mm.TypeListItemIndent:
						if trailingIndent {
							delete(listItemLines, child.StartLine)
						}
					case mm.TypeListItemPrefix:
						trailingIndent = true
					}
				}
			}
		}

		paragraphLines := map[int]struct{}{}
		codeInlineLines := map[int]struct{}{}

		if strict {
			for _, par := range p.FilterByTypesCached([]mm.TokenType{mm.TypeParagraph}, false) {
				mdhelpers.AddRangeToSet(paragraphLines, par.StartLine, par.EndLine-1)
			}

			for _, ct := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeText}, false) {
				mdhelpers.AddRangeToSet(codeInlineLines, ct.StartLine, ct.EndLine-1)
			}
		}

		expected := brSpaces
		if brSpaces < 2 {
			expected = 0
		}

		for lineIndex, line := range p.Lines {
			lineNumber := lineIndex + 1
			trimmed := strings.TrimRightFunc(line, unicode.IsSpace)

			trailingSpaces := runeLen(line) - runeLen(trimmed)
			if trailingSpaces == 0 {
				continue
			}

			_, inCode := codeBlockLines[lineNumber]
			_, inList := listItemLines[lineNumber]
			_, inPara := paragraphLines[lineNumber]
			_, inCodeInline := codeInlineLines[lineNumber]

			if inCode || inList {
				continue
			}

			if expected != trailingSpaces || (strict && (!inPara || inCodeInline)) {
				column := runeLen(line) - trailingSpaces + 1

				detailExpected := strconv.Itoa(expected)
				if expected != 0 {
					detailExpected = "0 or " + detailExpected
				}

				detail := "Expected: " + detailExpected + "; Actual: " + strconv.Itoa(
					trailingSpaces,
				)
				helpers.AddError(
					onError,
					lineNumber,
					detail,
					"",
					rng(column, trailingSpaces),
					fixDelete(column, trailingSpaces),
				)
			}
		}
	},
}
