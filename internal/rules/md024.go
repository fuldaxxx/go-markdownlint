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

var md024 = rule.Rule{
	Names:       []string{"MD024", "no-duplicate-heading"},
	Description: "Multiple headings with the same content",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD024
		siblingsOnly := types.BoolOr(c.SiblingsOnly, false)
		// knownContents indexed by heading level (1-6); index 0 unused.
		knownContents := map[int][]string{}
		contains := func(level int, text string) bool {
			for _, t := range knownContents[level] {
				if t == text {
					return true
				}
			}

			return false
		}
		lastLevel := 1
		knownLevel := lastLevel

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			headingText := mdhelpers.GetHeadingText(heading)
			if siblingsOnly {
				newLevel := mdhelpers.GetHeadingLevel(heading)
				for lastLevel < newLevel {
					lastLevel++
					knownContents[lastLevel] = nil
				}

				for lastLevel > newLevel {
					knownContents[lastLevel] = nil
					lastLevel--
				}

				knownLevel = newLevel
			}

			if contains(knownLevel, headingText) {
				helpers.AddErrorContext(
					onError,
					heading.StartLine,
					strings.TrimSpace(headingText),
					false,
					false,
					nil,
					nil,
				)
			} else {
				knownContents[knownLevel] = append(knownContents[knownLevel], headingText)
			}
		}
	},
}

func init() { register(&md024) }
