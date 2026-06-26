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
	"sort"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md044) }

var md044IgnoredChildTypes = map[mm.TokenType]bool{
	mm.TypeCodeFencedFence: true,
	mm.TypeDefinition:      true,
	mm.TypeReference:       true,
	mm.TypeResource:        true,
}

var (
	md044LeadingNonWordRe  = regexp.MustCompile(`^\W`)
	md044TrailingNonWordRe = regexp.MustCompile(`\W$`)
)

var md044 = rule.Rule{
	Names:       []string{"MD044", "proper-names"},
	Description: "Proper names should have the correct capitalization",
	Tags:        []string{"spelling"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD044

		names := c.Names
		if names == nil {
			names = []string{}
		}
		// names.sort((a, b) => (b.length - a.length) || a.localeCompare(b))
		sort.SliceStable(names, func(i, j int) bool {
			li, lj := runeLen(names[i]), runeLen(names[j])
			if li != lj {
				return li > lj
			}

			return names[i] < names[j]
		})

		if len(names) == 0 {
			return
		}

		namesSet := map[string]bool{}
		for _, n := range names {
			namesSet[n] = true
		}

		codeBlocks := types.BoolOr(c.CodeBlocks, true)
		htmlElements := types.BoolOr(c.HTMLElements, true)

		scannedTypes := map[mm.TokenType]bool{mm.TypeData: true}
		if codeBlocks {
			scannedTypes[mm.TypeCodeFlowValue] = true
			scannedTypes[mm.TypeCodeTextData] = true
		}

		if htmlElements {
			scannedTypes[mm.TypeHTMLFlowData] = true
			scannedTypes[mm.TypeHTMLTextData] = true
		}

		contentTokens := mdhelpers.FilterByPredicate(
			p.MicromarkTokens(),
			func(token *mm.Token) bool { return scannedTypes[token.Type] },
			func(token *mm.Token) []*mm.Token {
				var filtered []*mm.Token

				for _, t := range token.Children {
					if !md044IgnoredChildTypes[t.Type] {
						filtered = append(filtered, t)
					}
				}

				return filtered
			},
		)

		var exclusions []helpers.FileRange

		scannedTokens := map[*mm.Token]bool{}

		for _, name := range names {
			escapedName := helpers.EscapeForRegExp(name)

			startNamePattern := `\b_*`
			if md044LeadingNonWordRe.MatchString(name) {
				startNamePattern = ""
			}

			endNamePattern := `_*\b`
			if md044TrailingNonWordRe.MatchString(name) {
				endNamePattern = ""
			}

			namePattern := `(?i)(` + startNamePattern + `)(` + escapedName + `)` + endNamePattern
			nameRe := regexp.MustCompile(namePattern)

			for _, token := range contentTokens {
				text := token.Text

				matches := nameRe.FindAllStringSubmatchIndex(text, -1)
				for _, loc := range matches {
					leftMatch := ""
					if loc[2] >= 0 {
						leftMatch = text[loc[2]:loc[3]]
					}

					nameMatch := text[loc[4]:loc[5]]
					// match.index is rune-based column offset within token text.
					matchIndex := runeLen(text[:loc[0]])
					column := token.StartColumn + matchIndex + runeLen(leftMatch)
					length := runeLen(nameMatch)
					lineNumber := token.StartLine
					nameRange := helpers.FileRange{
						StartLine:   lineNumber,
						StartColumn: column,
						EndLine:     lineNumber,
						EndColumn:   column + length - 1,
					}
					overlapsExclusion := false

					for _, ex := range exclusions {
						if helpers.HasOverlap(ex, nameRange) {
							overlapsExclusion = true
							break
						}
					}

					if !namesSet[nameMatch] && !overlapsExclusion {
						var autolinkRanges []helpers.FileRange

						if !scannedTokens[token] {
							for _, tok := range mdhelpers.FilterByTypes(micromarkParseFlatMD044(text), []mm.TokenType{mm.TypeLiteralAutolink}, false) {
								autolinkRanges = append(autolinkRanges, helpers.FileRange{
									StartLine:   lineNumber,
									StartColumn: token.StartColumn + tok.StartColumn - 1,
									EndLine:     lineNumber,
									EndColumn:   token.EndColumn + tok.EndColumn - 1,
								})
							}

							exclusions = append(exclusions, autolinkRanges...)
							scannedTokens[token] = true
						}

						overlapsAutolink := false

						for _, ar := range autolinkRanges {
							if helpers.HasOverlap(ar, nameRange) {
								overlapsAutolink = true
								break
							}
						}

						if !overlapsAutolink {
							helpers.AddErrorDetailIf(
								onError,
								token.StartLine,
								name,
								nameMatch,
								"",
								"",
								rng(column, length),
								fixReplace(column, length, name),
							)
						}
					}

					exclusions = append(exclusions, nameRange)
				}
			}
		}
	},
}

// micromarkParseFlatMD044 parses text and returns the top-level token list, for
// detecting literalAutolink tokens.
func micromarkParseFlatMD044(text string) []*mm.Token {
	return mm.Parse(text).Children
}
