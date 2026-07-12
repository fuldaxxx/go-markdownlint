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

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

var md026 = rule.Rule{
	Names:       []string{"MD026", "no-trailing-punctuation"},
	Description: "Trailing punctuation in heading",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		punctuation := helpers.AllPunctuationNoQuestion
		if v := p.Config.MD026.Punctuation; v != nil {
			punctuation = helpers.Stringify(v)
		}

		if punctuation == "" {
			// Empty character class matches nothing; RE2 rejects [] so it is handled specially.
			return
		}

		trailingPunctuationRe := regexp.MustCompile(
			`\s*[` + helpers.EscapeForRegExp(punctuation) + `]+$`,
		)

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeadingText, mm.TypeSetextHeadingText}, false) {
			endColumn := heading.EndColumn
			endLine := heading.EndLine
			text := heading.Text

			fullMatch := trailingPunctuationRe.FindString(text)
			if fullMatch != "" &&
				!helpers.EndOfLineHTMLEntityRe.MatchString(text) &&
				!helpers.EndOfLineGemojiCodeRe.MatchString(text) {
				length := runeLen(fullMatch)
				column := endColumn - length
				helpers.AddError(
					onError,
					endLine,
					"Punctuation: '"+fullMatch+"'",
					"",
					rng(column, length),
					fixDelete(column, length),
				)
			}
		}
	},
}

func init() { register(&md026) }
