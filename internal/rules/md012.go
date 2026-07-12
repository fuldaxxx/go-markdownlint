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

func init() { register(&md012) }

var md012 = rule.Rule{
	Names:       []string{"MD012", "no-multiple-blanks"},
	Description: "Multiple consecutive blank lines",
	Tags:        []string{"whitespace", "blank_lines"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD012
		maximum := types.IntOr(c.Maximum, 1)

		codeBlockLines := map[int]struct{}{}
		for _, cb := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented}, false) {
			mdhelpers.AddRangeToSet(codeBlockLines, cb.StartLine, cb.EndLine)
		}

		count := 0

		for lineIndex, line := range p.Lines {
			_, inCode := codeBlockLines[lineIndex+1]
			if inCode || strings.TrimSpace(line) != "" {
				count = 0
			} else {
				count++
			}

			if maximum < count {
				helpers.AddErrorDetailIf(
					onError,
					lineIndex+1,
					maximum,
					count,
					"",
					"",
					nil,
					fixDelete(0, -1),
				)
			}
		}
	},
}
