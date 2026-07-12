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
	"strconv"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md001) }

var md001 = rule.Rule{
	Names:       []string{"MD001", "heading-increment"},
	Description: "Heading levels should only increment by one level at a time",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD001
		hasTitle := helpers.FrontMatterHasTitle(
			p.FrontMatterLines,
			types.StringOr(c.FrontMatterTitle, ""),
			c.FrontMatterTitle != nil,
		)

		prevLevel := 1<<53 - 1 // Number.MAX_SAFE_INTEGER analog
		if hasTitle {
			prevLevel = 1
		}

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			level := mdhelpers.GetHeadingLevel(heading)
			if level > prevLevel {
				helpers.AddErrorDetailIf(
					onError,
					heading.StartLine,
					"h"+strconv.Itoa(prevLevel+1),
					"h"+strconv.Itoa(level),
					"",
					"",
					nil,
					nil,
				)
			}

			prevLevel = level
		}
	},
}
