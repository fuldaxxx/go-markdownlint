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

var md020Re = regexp.MustCompile(
	`^(#+)([ \t]*)([^# \t\\]|[^# \t][^#]*?[^# \t\\])([ \t]*)((?:\\#)?)(#+)(\s*)$`,
)

var md020 = rule.Rule{
	Names:       []string{"MD020", "no-missing-space-closed-atx"},
	Description: "No space inside hashes on closed atx style heading",
	Tags:        []string{"headings", "atx_closed", "spaces"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		ignoreBlockLineNumbers := map[int]struct{}{}
		for _, b := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented, mm.TypeHTMLFlow}, false) {
			mdhelpers.AddRangeToSet(ignoreBlockLineNumbers, b.StartLine, b.EndLine)
		}

		for lineIndex, line := range p.Lines {
			if _, skip := ignoreBlockLineNumbers[lineIndex+1]; skip {
				continue
			}

			match := md020Re.FindStringSubmatch(line)
			if match == nil {
				continue
			}

			leftHash := match[1]
			leftSpaceLength := runeLen(match[2])
			content := match[3]
			rightSpaceLength := runeLen(match[4])
			rightEscape := match[5]
			rightHash := match[6]
			trailSpaceLength := runeLen(match[7])

			leftHashLength := runeLen(leftHash)
			rightHashLength := runeLen(rightHash)
			left := leftSpaceLength == 0
			right := rightSpaceLength == 0 || rightEscape != ""

			rightEscapeReplacement := ""
			if rightEscape != "" {
				rightEscapeReplacement = rightEscape + " "
			}

			if left || right {
				lineLen := runeLen(line)

				var r *[2]int
				if left {
					r = rng(1, leftHashLength+1)
				} else {
					r = rng(lineLen-trailSpaceLength-rightHashLength, rightHashLength+1)
				}

				helpers.AddErrorContext(
					onError,
					lineIndex+1,
					strings.TrimSpace(line),
					left,
					right,
					r,
					fixReplace(
						1,
						lineLen,
						leftHash+" "+content+" "+rightEscapeReplacement+rightHash,
					),
				)
			}
		}
	},
}

func init() { register(&md020) }
