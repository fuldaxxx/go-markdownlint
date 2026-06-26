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
	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md040) }

var md040 = rule.Rule{
	Names:       []string{"MD040", "fenced-code-language"},
	Description: "Fenced code blocks should have a language specified",
	Tags:        []string{"code", "language"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD040
		allowed := c.AllowedLanguages

		allowedSet := make(map[string]bool, len(allowed))
		for _, a := range allowed {
			allowedSet[a] = true
		}

		languageOnly := types.BoolOr(c.LanguageOnly, false)

		for _, fencedCode := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced}, false) {
			fences := mdhelpers.GetDescendantsByType(
				[]*mm.Token{fencedCode},
				[][]mm.TokenType{{mm.TypeCodeFencedFence}},
			)
			if len(fences) == 0 {
				continue
			}

			openingFence := fences[0]
			startLine := openingFence.StartLine
			text := openingFence.Text
			infos := mdhelpers.GetDescendantsByType(
				[]*mm.Token{openingFence},
				[][]mm.TokenType{{mm.TypeCodeFencedFenceInfo}},
			)

			info := ""
			if len(infos) > 0 {
				info = infos[0].Text
			}

			if info == "" {
				helpers.AddErrorContext(onError, startLine, text, false, false, nil, nil)
			} else if len(allowed) > 0 && !allowedSet[info] {
				helpers.AddError(onError, startLine, `"`+info+`" is not allowed`, "", nil, nil)
			}

			if languageOnly &&
				len(
					mdhelpers.GetDescendantsByType(
						[]*mm.Token{openingFence},
						[][]mm.TokenType{{mm.TypeCodeFencedFenceMeta}},
					),
				) > 0 {
				helpers.AddError(
					onError,
					startLine,
					`Info string contains more than language: "`+text+`"`,
					"",
					nil,
					nil,
				)
			}
		}
	},
}
