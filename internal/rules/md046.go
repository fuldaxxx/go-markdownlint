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
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md046) }

var md046 = rule.Rule{
	Names:       []string{"MD046", "code-block-style"},
	Description: "Code block style",
	Tags:        []string{"code"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		tokenTypeToStyle := func(tokenType mm.TokenType) string {
			if tokenType == mm.TypeCodeFenced {
				return "fenced"
			}

			return "indented"
		}

		c := p.Config.MD046

		expectedStyle := types.StringOr(c.Style, "consistent")
		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented}, false) {
			if expectedStyle == "consistent" {
				expectedStyle = tokenTypeToStyle(token.Type)
			}

			helpers.AddErrorDetailIf(
				onError,
				token.StartLine,
				expectedStyle,
				tokenTypeToStyle(token.Type),
				"",
				"",
				nil,
				nil,
			)
		}
	},
}
