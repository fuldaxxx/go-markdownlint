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

func init() { register(&md027) }

var md027ListTypes = []mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered}

var md027 = rule.Rule{
	Names:       []string{"MD027", "no-multiple-space-blockquote"},
	Description: "Multiple spaces after blockquote symbol",
	Tags:        []string{"blockquote", "whitespace", "indentation"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD027
		includeListItems := types.BoolOr(c.ListItems, true)
		topTokens := p.MicromarkTokens()

		isListType := func(t mm.TokenType) bool {
			for _, lt := range md027ListTypes {
				if lt == t {
					return true
				}
			}

			return false
		}

		for _, token := range p.FilterByTypesCached([]mm.TokenType{mm.TypeLinePrefix}, false) {
			parent := token.Parent
			codeIndented := parent != nil && parent.Type == mm.TypeCodeIndented

			siblings := topTokens
			if parent != nil {
				siblings = parent.Children
			}
			// Locate token index within siblings.
			idx := -1

			for i, s := range siblings {
				if s == token {
					idx = i
					break
				}
			}

			var prevType, nextType mm.TokenType
			if idx-1 >= 0 {
				prevType = siblings[idx-1].Type
			}

			if idx+1 < len(siblings) {
				nextType = siblings[idx+1].Type
			}

			if !codeIndented &&
				prevType == mm.TypeBlockQuotePrefix &&
				(includeListItems || (!isListType(nextType) &&
					mdhelpers.GetParentOfType(token, md027ListTypes) == nil)) {
				startColumn := token.StartColumn
				startLine := token.StartLine
				length := runeLen(token.Text)
				line := p.Lines[startLine-1]
				helpers.AddErrorContext(
					onError,
					startLine,
					line,
					false,
					false,
					rng(startColumn, length),
					fixDelete(startColumn, length),
				)
			}
		}
	},
}
