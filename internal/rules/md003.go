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
	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md003) }

var md003 = rule.Rule{
	Names:       []string{"MD003", "heading-style"},
	Description: "Heading style",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD003
		style := types.StringOr(c.Style, "consistent")

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			styleForToken := mdhelpers.GetHeadingStyle(heading)
			if style == "consistent" {
				style = styleForToken
			}

			if styleForToken == style {
				continue
			}

			h12 := mdhelpers.GetHeadingLevel(heading) <= 2
			setextWithAtx := style == "setext_with_atx" &&
				((h12 && styleForToken == "setext") || (!h12 && styleForToken == "atx"))

			setextWithAtxClosed := style == "setext_with_atx_closed" &&
				((h12 && styleForToken == "setext") || (!h12 && styleForToken == "atx_closed"))
			if setextWithAtx || setextWithAtxClosed {
				continue
			}

			expected := style
			switch style {
			case "setext_with_atx":
				if h12 {
					expected = "setext"
				} else {
					expected = "atx"
				}
			case "setext_with_atx_closed":
				if h12 {
					expected = "setext"
				} else {
					expected = "atx_closed"
				}
			}

			helpers.AddErrorDetailIf(
				onError,
				heading.StartLine,
				expected,
				styleForToken,
				"",
				"",
				nil,
				nil,
			)
		}
	},
}
