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
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

var md023 = rule.Rule{
	Names:       []string{"MD023", "heading-start-left"},
	Description: "Headings must start at the beginning of the line",
	Tags:        []string{"headings", "spaces"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		headings := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeAtxHeading, mm.TypeLinePrefix, mm.TypeSetextHeading},
			false,
		)
		for i := 0; i < len(headings)-1; i++ {
			if headings[i].Type == mm.TypeLinePrefix &&
				headings[i+1].Type != mm.TypeLinePrefix &&
				headings[i].StartLine == headings[i+1].StartLine {
				startColumn := headings[i].StartColumn
				endColumn := headings[i].EndColumn
				length := endColumn - startColumn
				helpers.AddErrorContext(
					onError,
					headings[i].StartLine,
					p.Lines[headings[i].StartLine-1],
					true,
					false,
					rng(startColumn, length),
					fixDelete(startColumn, length),
				)
			}
		}
	},
}

func init() { register(&md023) }
