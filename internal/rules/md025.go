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

var md025 = rule.Rule{
	Names:       []string{"MD025", "single-title", "single-h1"},
	Description: "Multiple top-level headings in the same document",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD025
		level := types.IntOr(c.Level, 1)
		tokens := p.MicromarkTokens()

		var matchingHeadings []*mm.Token

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			if level == mdhelpers.GetHeadingLevel(heading) && !mdhelpers.IsDocfxTab(heading) {
				matchingHeadings = append(matchingHeadings, heading)
			}
		}

		if len(matchingHeadings) == 0 {
			return
		}

		foundFrontMatterTitle := helpers.FrontMatterHasTitle(
			p.FrontMatterLines,
			types.StringOr(c.FrontMatterTitle, ""),
			c.FrontMatterTitle != nil,
		)

		hasTopLevelHeading := foundFrontMatterTitle
		if !hasTopLevelHeading {
			// Check tokens preceding the first matching heading are all non-content.
			// When the matching heading is not found (firstIndex == -1), the slice
			// end becomes len(tokens)-1, i.e. all but the last token.
			firstIndex := -1

			for idx, t := range tokens {
				if t == matchingHeadings[0] {
					firstIndex = idx
					break
				}
			}

			sliceEnd := firstIndex
			if firstIndex < 0 {
				sliceEnd = len(tokens) - 1
				if sliceEnd < 0 {
					sliceEnd = 0
				}
			}

			hasTopLevelHeading = true

			for _, t := range tokens[:sliceEnd] {
				if !mdhelpers.NonContentTokens[t.Type] && !mdhelpers.IsHTMLFlowComment(t) {
					hasTopLevelHeading = false
					break
				}
			}
		}

		if hasTopLevelHeading {
			start := 1
			if foundFrontMatterTitle {
				start = 0
			}

			for _, heading := range matchingHeadings[start:] {
				helpers.AddErrorContext(
					onError,
					heading.StartLine,
					mdhelpers.GetHeadingText(heading),
					false,
					false,
					nil,
					nil,
				)
			}
		}
	},
}

func init() { register(&md025) }
