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

func init() { register(&md037) }

var (
	md037StartRe = regexp.MustCompile(`^\s+\S`)
	md037EndRe   = regexp.MustCompile(`\S\s+$`)
)

var md037 = rule.Rule{
	Names:       []string{"MD037", "no-space-in-emphasis"},
	Description: "Spaces inside emphasis markers",
	Tags:        []string{"whitespace", "emphasis"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		lines := p.Lines

		markers := []string{"_", "__", "___", "*", "**", "***"}

		isMarker := map[string]bool{}
		for _, m := range markers {
			isMarker[m] = true
		}

		tokens := mdhelpers.FilterByPredicate(
			p.MicromarkTokens(),
			func(token *mm.Token) bool {
				for _, child := range token.Children {
					if child.Type == mm.TypeData {
						return true
					}
				}

				return false
			},
			nil,
		)

		for _, token := range tokens {
			// Build lists of bare tokens for each emphasis marker type.
			emphasisTokensByMarker := map[string][]*mm.Token{}
			for _, m := range markers {
				emphasisTokensByMarker[m] = nil
			}

			for _, child := range token.Children {
				text := child.Text
				if child.Type == mm.TypeData && runeLen(text) <= 3 && isMarker[text] {
					if !child.InHTMLFlow() {
						emphasisTokensByMarker[text] = append(emphasisTokensByMarker[text], child)
					}
				}
			}

			// Process bare tokens for each emphasis marker type.
			for _, marker := range markers {
				emphasisTokens := emphasisTokensByMarker[marker]
				for i := 0; i+1 < len(emphasisTokens); i += 2 {
					// Process start token of start/end pair.
					startToken := emphasisTokens[i]
					startLineRunes := []rune(lines[startToken.StartLine-1])

					startSlice := string(startLineRunes[startToken.EndColumn-1:])
					if m := md037StartRe.FindString(startSlice); m != "" {
						startSpaceCharacter := m
						startContext := marker + startSpaceCharacter
						column := startToken.EndColumn
						count := runeLen(startSpaceCharacter) - 1
						helpers.AddError(
							onError,
							startToken.StartLine,
							"",
							startContext,
							rng(column, count),
							fixDelete(column, count),
						)
					}

					// Process end token of start/end pair.
					endToken := emphasisTokens[i+1]
					endLineRunes := []rune(lines[endToken.StartLine-1])

					endSlice := string(endLineRunes[:endToken.StartColumn-1])
					if m := md037EndRe.FindString(endSlice); m != "" {
						endSpaceCharacter := m
						endContext := endSpaceCharacter + marker
						column := endToken.StartColumn - (runeLen(endSpaceCharacter) - 1)
						count := runeLen(endSpaceCharacter) - 1
						helpers.AddError(
							onError,
							endToken.StartLine,
							"",
							endContext,
							rng(column, count),
							fixDelete(column, count),
						)
					}
				}
			}
		}
	},
}
