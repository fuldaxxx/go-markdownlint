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
			// Only flag paragraphs at the top level of the document (or embedded
			// directly in an HTML block). Paragraphs nested in block quotes or
			// list items are not headings-in-disguise and must be ignored.
			//
			// Upstream markdownlint wraps each flow paragraph in a "content"
			// token and checks content.parent; this port's tokenizer attaches the
			// paragraph directly to its container, so a top-level paragraph has a
			// nil parent. Checking for the (absent) "content" wrapper here skipped
			// every paragraph and stopped MD036 from ever firing.
			parent := token.Parent

			okParent := parent == nil ||
				(parent.Type == mm.TypeHTMLFlow && parent.Parent == nil)
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
				if len(textToken.Children) != 1 || textToken.Children[0].Type != mm.TypeData {
					continue
				}

				// This port's tokenizer keeps the emphasis text on the child data
				// token rather than on the emphasis-text token itself, so read the
				// text from the child. Using textToken.Text (empty here) made the
				// trailing-punctuation check always pass and flagged emphasized
				// sentences like "**Done.**".
				innerText := textToken.Text
				if innerText == "" {
					innerText = textToken.Children[0].Text
				}

				if punctuationRe.MatchString(innerText) {
					continue
				}

				helpers.AddErrorContext(
					onError,
					textToken.StartLine,
					innerText,
					false,
					false,
					nil,
					nil,
				)
			}
		}
	},
}

func init() { register(&md036) }
