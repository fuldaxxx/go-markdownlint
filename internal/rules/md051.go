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
	"strconv"
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md051) }

var (
	md051IdRe           = helpers.GetHTMLAttributeRe("id")
	md051NameRe         = helpers.GetHTMLAttributeRe("name")
	md051AnchorRe       = regexp.MustCompile(`\{(#[a-z\d]+(?:[-_][a-z\d]+)*)\}`)
	md051LineFragmentRe = regexp.MustCompile(`^#(?:L\d+(?:C\d+)?-L\d+(?:C\d+)?|L\d+)$`)
	// Removes any char that is not a letter, mark, number, connector
	// punctuation, hyphen, or space.
	md051FragmentStripRe = regexp.MustCompile(`[^\p{L}\p{M}\p{N}\p{Pc}\- ]`)
)

var md051ChildrenExclude = map[mm.TokenType]bool{
	mm.TypeImage:     true,
	mm.TypeReference: true,
	mm.TypeResource:  true,
}

var md051TokensInclude = map[mm.TokenType]bool{
	mm.TypeCharacterEscapeValue: true,
	mm.TypeCodeTextData:         true,
	mm.TypeData:                 true,
	mm.TypeMathTextData:         true,
}

// md051EncodeURIComponent percent-encodes every byte except the unreserved set
// A-Za-z0-9 and -_.!~*'()
func md051EncodeURIComponent(s string) string {
	const upperhex = "0123456789ABCDEF"

	unreserved := func(c byte) bool {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			return true
		}

		switch c {
		case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
			return true
		}

		return false
	}

	var b strings.Builder

	for i := 0; i < len(s); i++ {
		c := s[i]
		if unreserved(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&0x0f])
		}
	}

	return b.String()
}

// md051ConvertHeadingToHTMLFragment ports convertHeadingToHTMLFragment.
func md051ConvertHeadingToHTMLFragment(headingText *mm.Token) string {
	inlineTokens := mdhelpers.FilterByPredicate(
		headingText.Children,
		func(token *mm.Token) bool { return md051TokensInclude[token.Type] },
		func(token *mm.Token) []*mm.Token {
			if md051ChildrenExclude[token.Type] {
				return []*mm.Token{}
			}

			return token.Children
		},
	)

	var sb strings.Builder
	for _, t := range inlineTokens {
		sb.WriteString(t.Text)
	}

	inlineText := strings.ToLower(sb.String())
	inlineText = md051FragmentStripRe.ReplaceAllString(inlineText, "")
	inlineText = strings.ReplaceAll(inlineText, " ", "-")

	return "#" + md051EncodeURIComponent(inlineText)
}

// md051UnescapeStringTokenText ports unescapeStringTokenText.
func md051UnescapeStringTokenText(token *mm.Token) string {
	children := mdhelpers.FilterByTypes(
		token.Children,
		[]mm.TokenType{mm.TypeCharacterEscapeValue, mm.TypeData},
		false,
	)

	// This port's tokenizer stores the destination/definition text directly on
	// the string token instead of nesting data/characterEscape children under
	// it. Fall back to the token's own text when there are no such children;
	// otherwise the fragment reads as empty and MD051 can never fire.
	if len(children) == 0 {
		return token.Text
	}

	var sb strings.Builder
	for _, c := range children {
		sb.WriteString(c.Text)
	}

	return sb.String()
}

var md051 = rule.Rule{
	Names:       []string{"MD051", "link-fragments"},
	Description: "Link fragments should be valid",
	Tags:        []string{"links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD051
		ignoreCase := types.BoolOr(c.IgnoreCase, false)
		ignoredPattern := types.StringOr(c.IgnoredPattern, "")

		ignoredPatternSrc := ignoredPattern
		if ignoredPatternSrc == "" {
			ignoredPatternSrc = "^$"
		}

		ignoredPatternRe := regexp.MustCompile(ignoredPatternSrc)

		// fragments preserves insertion order for the mixed-case key lookup.
		fragments := map[string]int{"#top": 0}

		var fragmentOrder []string

		fragmentOrder = append(fragmentOrder, "#top")
		setFragment := func(key string, val int) {
			if _, ok := fragments[key]; !ok {
				fragmentOrder = append(fragmentOrder, key)
			}

			fragments[key] = val
		}

		// Process headings
		headingTexts := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeAtxHeadingText, mm.TypeSetextHeadingText},
			false,
		)
		for _, headingText := range headingTexts {
			fragment := md051ConvertHeadingToHTMLFragment(headingText)
			if fragment != "#" {
				count := fragments[fragment]
				if count != 0 {
					setFragment(fragment+"-"+strconv.Itoa(count), 0)
				}

				setFragment(fragment, count+1)

				for _, m := range md051AnchorRe.FindAllStringSubmatch(headingText.Text, -1) {
					anchor := m[1]
					if _, ok := fragments[anchor]; !ok {
						setFragment(anchor, 1)
					}
				}
			}
		}

		// Process HTML anchors
		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeHTMLText}, true) {
			htmlTagInfo := mdhelpers.GetHTMLTagInfo(token)
			if htmlTagInfo != nil && !htmlTagInfo.Close {
				var anchorMatch []string
				if m := md051IdRe.FindStringSubmatch(token.Text); m != nil {
					anchorMatch = m
				} else if strings.ToLower(htmlTagInfo.Name) == "a" {
					if m := md051NameRe.FindStringSubmatch(token.Text); m != nil {
						anchorMatch = m
					}
				}

				if len(anchorMatch) > 0 {
					setFragment("#"+anchorMatch[1], 0)
				}
			}
		}

		// Process link and definition fragments
		parentChilds := [][2]mm.TokenType{
			{mm.TypeLink, mm.TypeResourceDestinationString},
			{mm.TypeDefinition, mm.TypeDefinitionDestinationString},
		}
		for _, pc := range parentChilds {
			parentType, definitionType := pc[0], pc[1]
			all := p.FilterByTypesCached([]mm.TokenType{parentType}, false)

			var links []*mm.Token

			for _, link := range all {
				if link.Parent != nil && link.Parent.Type == mm.TypeAtxHeadingText &&
					mdhelpers.IsDocfxTab(link.Parent.Parent) {
					continue
				}

				links = append(links, link)
			}

			for _, link := range links {
				definitions := mdhelpers.FilterByTypes(
					link.Children,
					[]mm.TokenType{definitionType},
					false,
				)
				for _, definition := range definitions {
					endColumn := definition.EndColumn
					startColumn := definition.StartColumn
					text := md051UnescapeStringTokenText(definition)

					textSliceOne := ""
					if runeLen(text) > 1 {
						textSliceOne = string([]rune(text)[1:])
					} else if runeLen(text) == 1 {
						textSliceOne = ""
					}

					encodedText := "#" + md051EncodeURIComponent(textSliceOne)

					_, hasFragment := fragments[encodedText]
					if runeLen(text) > 1 &&
						strings.HasPrefix(text, "#") &&
						!hasFragment &&
						!md051LineFragmentRe.MatchString(encodedText) &&
						!ignoredPatternRe.MatchString(textSliceOne) {
						context := ""

						var (
							rangePtr *[2]int
							fixInfo  *types.FixInfo
						)

						if link.StartLine == link.EndLine {
							context = link.Text
							rangePtr = rng(link.StartColumn, link.EndColumn-link.StartColumn)
							fixInfo = fixDelete(startColumn, endColumn-startColumn)
						}

						textLower := strings.ToLower(text)
						mixedCaseKey := ""
						foundMixed := false

						for _, key := range fragmentOrder {
							if textLower == strings.ToLower(key) {
								mixedCaseKey = key
								foundMixed = true

								break
							}
						}

						if foundMixed {
							if fixInfo != nil {
								fixInfo.InsertText = mixedCaseKey
							}

							if !ignoreCase && mixedCaseKey != text {
								helpers.AddError(
									onError,
									link.StartLine,
									"Expected: "+mixedCaseKey+"; Actual: "+text,
									context,
									rangePtr,
									fixInfo,
								)
							}
						} else {
							helpers.AddError(onError, link.StartLine, "", context, rangePtr, nil)
						}
					}
				}
			}
		}
	},
}
