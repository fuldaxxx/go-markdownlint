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

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

var md036EmphasisTypes = [][]mm.TokenType{
	{mm.TypeEmphasis, mm.TypeEmphasisText},
	{mm.TypeStrong, mm.TypeStrongText},
}

var md036 = rule.Rule{
	Names:       []string{"MD036", "no-emphasis-as-heading"},
	Description: "Emphasis used instead of a heading",
	Tags:        []string{"headings", "emphasis"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		punctuation := helpers.AllPunctuation
		if v := p.Config.MD036.Punctuation; v != nil {
			punctuation = helpers.Stringify(v)
		}

		punctuationRe := regexp.MustCompile(`[` + punctuation + `]$`)

		isParagraphChildMeaningful := func(token *mm.Token) bool {
			return token.Type != mm.TypeHTMLText &&
				(token.Type != mm.TypeData || strings.TrimSpace(token.Text) != "")
		}

		var paragraphTokens []*mm.Token

		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeParagraph}, true) {
			parent := token.Parent
			if parent == nil || parent.Type != mm.TypeContent {
				continue
			}

			grandparent := parent.Parent

			okParent := grandparent == nil ||
				(grandparent.Type == mm.TypeHTMLFlow && grandparent.Parent == nil)
			if !okParent {
				continue
			}

			meaningful := 0

			for _, c := range token.Children {
				if isParagraphChildMeaningful(c) {
					meaningful++
				}
			}

			if meaningful == 1 {
				paragraphTokens = append(paragraphTokens, token)
			}
		}

		for _, emphasisType := range md036EmphasisTypes {
			textTokens := mdhelpers.GetDescendantsByType(
				paragraphTokens,
				[][]mm.TokenType{emphasisType[:1], emphasisType[1:]},
			)
			for _, textToken := range textTokens {
				if len(textToken.Children) == 1 &&
					textToken.Children[0].Type == mm.TypeData &&
					!punctuationRe.MatchString(textToken.Text) {
					helpers.AddErrorContext(
						onError,
						textToken.StartLine,
						textToken.Text,
						false,
						false,
						nil,
						nil,
					)
				}
			}
		}
	},
}

func init() { register(&md036) }
