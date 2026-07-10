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

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

var md041HeadingTagNameRe = regexp.MustCompile(`^h[1-6]$`)

// md041GetHTMLFlowTagName returns the lowercased HTML tag name of an htmlFlow
// token, or "" if none.
func md041GetHTMLFlowTagName(token *mm.Token) string {
	if token.Type != mm.TypeHTMLFlow {
		return ""
	}

	htmlTexts := mdhelpers.FilterByTypes(token.Children, []mm.TokenType{mm.TypeHTMLText}, true)
	if len(htmlTexts) == 0 {
		return ""
	}

	tagInfo := mdhelpers.GetHTMLTagInfo(htmlTexts[0])
	if tagInfo == nil {
		return ""
	}

	return strings.ToLower(tagInfo.Name)
}

var md041 = rule.Rule{
	Names:       []string{"MD041", "first-line-heading", "first-line-h1"},
	Description: "First line in a file should be a top-level heading",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD041
		allowPreamble := types.BoolOr(c.AllowPreamble, false)
		level := types.IntOr(c.Level, 1)

		tokens := p.MicromarkTokens()
		if helpers.FrontMatterHasTitle(
			p.FrontMatterLines,
			types.StringOr(c.FrontMatterTitle, ""),
			c.FrontMatterTitle != nil,
		) {
			return
		}

		errorLineNumber := 0

		for _, token := range tokens {
			startLine := token.StartLine

			typ := token.Type
			if mdhelpers.NonContentTokens[typ] || mdhelpers.IsHTMLFlowComment(token) {
				continue
			}

			if typ == mm.TypeAtxHeading || typ == mm.TypeSetextHeading {
				if mdhelpers.GetHeadingLevel(token) != level {
					errorLineNumber = startLine
				}

				break
			}

			tagName := md041GetHTMLFlowTagName(token)
			if tagName != "" && md041HeadingTagNameRe.MatchString(tagName) {
				if tagName != "h"+strconv.Itoa(level) {
					errorLineNumber = startLine
				}

				break
			} else if !allowPreamble {
				errorLineNumber = startLine
				break
			}
		}

		if errorLineNumber > 0 {
			helpers.AddErrorContext(
				onError,
				errorLineNumber,
				p.Lines[errorLineNumber-1],
				false,
				false,
				nil,
				nil,
			)
		}
	},
}

func init() { register(&md041) }
