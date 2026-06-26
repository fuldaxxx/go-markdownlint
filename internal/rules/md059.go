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

func init() { register(&md059) }

var md059AllowedChildrenTypes = map[mm.TokenType]bool{
	mm.TypeCodeText: true,
	mm.TypeHTMLText: true,
}

var md059DefaultProhibitedTexts = []string{
	"click here",
	"here",
	"link",
	"more",
}

var (
	md059NonWordRe    = regexp.MustCompile(`[\W_]+`)
	md059WhitespaceRe = regexp.MustCompile(`\s+`)
)

// md059Normalize removes extra whitespace and punctuation.
func md059Normalize(str string) string {
	str = md059NonWordRe.ReplaceAllString(str, " ")
	str = md059WhitespaceRe.ReplaceAllString(str, " ")
	str = strings.ToLower(str)

	return strings.TrimSpace(str)
}

var md059 = rule.Rule{
	Names:       []string{"MD059", "descriptive-link-text"},
	Description: "Link text should be descriptive",
	Tags:        []string{"accessibility", "links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		var texts []string
		if c := p.Config.MD059; c.ProhibitedTexts != nil {
			texts = *c.ProhibitedTexts
		} else {
			texts = md059DefaultProhibitedTexts
		}

		prohibitedTexts := map[string]struct{}{}
		for _, t := range texts {
			prohibitedTexts[md059Normalize(t)] = struct{}{}
		}

		if len(prohibitedTexts) == 0 {
			return
		}

		links := p.FilterByTypesCached([]mm.TokenType{mm.TypeLink}, false)
		for _, link := range links {
			labelTexts := mdhelpers.GetDescendantsByType(
				[]*mm.Token{link},
				[][]mm.TokenType{{mm.TypeLabel}, {mm.TypeLabelText}},
			)
			for _, labelText := range labelTexts {
				hasAllowedChild := false

				for _, child := range labelText.Children {
					if md059AllowedChildrenTypes[child.Type] {
						hasAllowedChild = true
						break
					}
				}

				if _, prohibited := prohibitedTexts[md059Normalize(labelText.Text)]; !hasAllowedChild &&
					prohibited {
					var r *[2]int
					if labelText.StartLine == labelText.EndLine {
						r = rng(labelText.StartColumn, labelText.EndColumn-labelText.StartColumn)
					}

					context := ""
					if labelText.Parent != nil {
						context = labelText.Parent.Text
					}

					helpers.AddErrorContext(
						onError,
						labelText.StartLine,
						context,
						false,
						false,
						r,
						nil,
					)
				}
			}
		}
	},
}
