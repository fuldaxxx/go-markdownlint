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

func init() { register(&md048) }

var md048 = rule.Rule{
	Names:       []string{"MD048", "code-fence-style"},
	Description: "Code fence style",
	Tags:        []string{"code"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		fencedCodeBlockStyleFor := func(markup string) string {
			if len(markup) > 0 && markup[0] == '~' {
				return "tilde"
			}

			return "backtick"
		}

		c := p.Config.MD048
		expectedStyle := types.StringOr(c.Style, "consistent")

		for _, codeFenced := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced}, false) {
			descendants := mdhelpers.GetDescendantsByType(
				[]*mm.Token{codeFenced},
				[][]mm.TokenType{{mm.TypeCodeFencedFence}, {mm.TypeCodeFencedFenceSeq}},
			)
			if len(descendants) == 0 {
				continue
			}

			seq := descendants[0]
			if expectedStyle == "consistent" {
				expectedStyle = fencedCodeBlockStyleFor(seq.Text)
			}

			helpers.AddErrorDetailIf(
				onError,
				seq.StartLine,
				expectedStyle,
				fencedCodeBlockStyleFor(seq.Text),
				"",
				"",
				nil,
				nil,
			)
		}
	},
}
