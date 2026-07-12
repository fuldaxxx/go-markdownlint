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

func init() { register(&md045) }

var (
	md045AltRe        = helpers.GetHTMLAttributeRe("alt")
	md045AriaHiddenRe = helpers.GetHTMLAttributeRe("aria-hidden")
)

var md045 = rule.Rule{
	Names:       []string{"MD045", "no-alt-text"},
	Description: "Images should have alternate text (alt text)",
	Tags:        []string{"accessibility", "images"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		// Process Markdown images
		images := p.FilterByTypesCached([]mm.TokenType{mm.TypeImage}, false)
		for _, image := range images {
			labelTexts := mdhelpers.GetDescendantsByType(
				[]*mm.Token{image},
				[][]mm.TokenType{{mm.TypeLabel}, {mm.TypeLabelText}},
			)
			empty := false

			for _, labelText := range labelTexts {
				if len(labelText.Text) == 0 {
					empty = true
					break
				}
			}

			if empty {
				var r *[2]int
				if image.StartLine == image.EndLine {
					r = rng(image.StartColumn, image.EndColumn-image.StartColumn)
				}

				helpers.AddError(onError, image.StartLine, "", "", r, nil)
			}
		}

		// Process HTML images
		htmlTexts := p.FilterByTypesCached([]mm.TokenType{mm.TypeHTMLText}, true)
		for _, htmlText := range htmlTexts {
			text := htmlText.Text

			htmlTagInfo := mdhelpers.GetHTMLTagInfo(htmlText)
			if htmlTagInfo != nil &&
				!htmlTagInfo.Close &&
				strings.ToLower(htmlTagInfo.Name) == "img" &&
				!md045AltRe.MatchString(text) {
				ariaHidden := ""
				if m := md045AriaHiddenRe.FindStringSubmatch(text); m != nil {
					ariaHidden = strings.ToLower(m[1])
				}

				if ariaHidden != "true" {
					length := runeLen(helpers.NextLinesRe.ReplaceAllString(text, ""))
					helpers.AddError(
						onError,
						htmlText.StartLine,
						"",
						"",
						rng(htmlText.StartColumn, length),
						nil,
					)
				}
			}
		}
	},
}
