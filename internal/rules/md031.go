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

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md031) }

var (
	md031CodeFencePrefixRe = regexp.MustCompile("^(.*?)[`~]")
	md031NonGtRe           = regexp.MustCompile(`[^>]`)
)

var md031 = rule.Rule{
	Names:       []string{"MD031", "blanks-around-fences"},
	Description: "Fenced code blocks should be surrounded by blank lines",
	Tags:        []string{"code", "blank_lines"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		lines := p.Lines
		addError := func(lineNumber int, top bool) {
			line := lines[lineNumber-1]
			m := md031CodeFencePrefixRe.FindStringSubmatch(line)

			var fixInfo *types.FixInfo

			if m != nil {
				prefix := m[1]
				insertText := strings.TrimSpace(md031NonGtRe.ReplaceAllString(prefix, " ")) + "\n"

				fixLineNumber := lineNumber
				if !top {
					fixLineNumber = lineNumber + 1
				}

				fixInfo = &types.FixInfo{LineNumber: fixLineNumber, InsertText: insertText}
			}

			helpers.AddErrorContext(
				onError,
				lineNumber,
				strings.TrimSpace(line),
				false,
				false,
				nil,
				fixInfo,
			)
		}

		c := p.Config.MD031

		listItems := types.BoolOr(c.ListItems, true)
		for _, codeBlock := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced}, false) {
			if listItems ||
				mdhelpers.GetParentOfType(
					codeBlock,
					[]mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered},
				) == nil {
				if !helpers.IsBlankLine(md031LineAt(lines, codeBlock.StartLine-2)) {
					addError(codeBlock.StartLine, true)
				}

				if !helpers.IsBlankLine(md031LineAt(lines, codeBlock.EndLine)) &&
					!helpers.IsBlankLine(md031LineAt(lines, codeBlock.EndLine-1)) {
					addError(codeBlock.EndLine, false)
				}
			}
		}
	},
}

// md031LineAt returns lines[index] (0-based) or "" when out of range. An empty
// string is treated as a blank line by isBlankLine, so out-of-range indices
// count as blank.
func md031LineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}

	return lines[index]
}
