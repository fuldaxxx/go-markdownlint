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

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md013) }

// Regular expression for a line that is not wrappable.
var md013NotWrappableRe = regexp.MustCompile(`^(?:[#>\s]*\s)?\S*$`)

// Trailing run of non-whitespace (possibly empty) at end of line.
var md013TrailingNonWhitespaceRe = regexp.MustCompile(`\S*$`)

var md013 = rule.Rule{
	Names:       []string{"MD013", "line-length"},
	Description: "Line length",
	Tags:        []string{"line_length"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD013
		lineLength := types.IntOr(c.LineLength, 80)
		headingLineLength := types.IntOr(c.HeadingLineLength, lineLength)
		codeLineLength := types.IntOr(c.CodeBlockLineLength, lineLength)
		strict := types.BoolOr(c.Strict, false)
		stern := types.BoolOr(c.Stern, false)
		includeCodeBlocks := types.BoolOr(c.CodeBlocks, true)
		includeTables := types.BoolOr(c.Tables, true)
		includeHeadings := types.BoolOr(c.Headings, true)

		headingLineNumbers := map[int]struct{}{}
		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			mdhelpers.AddRangeToSet(headingLineNumbers, heading.StartLine, heading.EndLine)
		}

		codeBlockLineNumbers := map[int]struct{}{}
		for _, codeBlock := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented}, false) {
			mdhelpers.AddRangeToSet(codeBlockLineNumbers, codeBlock.StartLine, codeBlock.EndLine)
		}

		tableLineNumbers := map[int]struct{}{}
		for _, table := range p.FilterByTypesCached([]mm.TokenType{mm.TypeTable}, false) {
			mdhelpers.AddRangeToSet(tableLineNumbers, table.StartLine, table.EndLine)
		}

		linkLineNumbers := map[int]struct{}{}
		for _, link := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAutolink, mm.TypeImage, mm.TypeLink, mm.TypeLiteralAutolink}, false) {
			mdhelpers.AddRangeToSet(linkLineNumbers, link.StartLine, link.EndLine)
		}

		paragraphDataLineNumbers := map[int]struct{}{}

		for _, paragraph := range p.FilterByTypesCached([]mm.TokenType{mm.TypeParagraph}, false) {
			for _, data := range mdhelpers.GetDescendantsByType([]*mm.Token{paragraph}, [][]mm.TokenType{{mm.TypeData}}) {
				mdhelpers.AddRangeToSet(paragraphDataLineNumbers, data.StartLine, data.EndLine)
			}
		}

		linkOnlyLineNumbers := map[int]struct{}{}

		for lineNumber := range linkLineNumbers {
			if _, ok := paragraphDataLineNumbers[lineNumber]; !ok {
				linkOnlyLineNumbers[lineNumber] = struct{}{}
			}
		}

		definitionLineIndices := map[int]struct{}{}
		for _, idx := range p.ReferenceLinkImageData().DefinitionLineIndices {
			definitionLineIndices[idx] = struct{}{}
		}

		for lineIndex, line := range p.Lines {
			lineNumber := lineIndex + 1
			_, isHeading := headingLineNumbers[lineNumber]
			_, inCode := codeBlockLineNumbers[lineNumber]
			_, inTable := tableLineNumbers[lineNumber]

			maxLength := lineLength
			if inCode {
				maxLength = codeLineLength
			} else if isHeading {
				maxLength = headingLineLength
			}
			// If not strict/stern, the last run of non-whitespace is allowed to go
			// beyond the limit as long as it begins within the limit.
			text := line
			if !strict && !stern {
				text = md013TrailingNonWhitespaceRe.ReplaceAllString(line, "#")
			}

			_, isLinkOnly := linkOnlyLineNumbers[lineNumber]

			_, isDefinition := definitionLineIndices[lineIndex]
			if (maxLength > 0) &&
				(includeCodeBlocks || !inCode) &&
				(includeTables || !inTable) &&
				(includeHeadings || !isHeading) &&
				!isDefinition &&
				(strict ||
					((!stern || !md013NotWrappableRe.MatchString(line)) &&
						!isLinkOnly)) &&
				(runeLen(text) > maxLength) {
				helpers.AddErrorDetailIf(
					onError,
					lineNumber,
					maxLength,
					runeLen(line),
					"",
					"",
					rng(maxLength+1, runeLen(line)-maxLength),
					nil,
				)
			}
		}
	},
}
