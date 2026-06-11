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
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() {
	register(&md019)
	register(&md021)
}

// md019md021ValidateHeadingSpaces validates heading sequence and whitespace
// length at the start (delta > 0) or end (delta < 0).
func md019md021ValidateHeadingSpaces(onError types.OnError, heading *mm.Token, delta int) {
	children := heading.Children
	startLine := heading.StartLine
	text := heading.Text

	var index int
	if delta > 0 {
		index = 0
	} else {
		index = len(children) - 1
	}

	for index >= 0 && index < len(children) && children[index].Type != mm.TypeAtxHeadingSequence {
		index += delta
	}

	var headingSequence *mm.Token
	if index >= 0 && index < len(children) {
		headingSequence = children[index]
	}

	var whitespace *mm.Token

	wi := index + delta
	if wi >= 0 && wi < len(children) {
		whitespace = children[wi]
	}

	if headingSequence != nil && headingSequence.Type == mm.TypeAtxHeadingSequence &&
		whitespace != nil && whitespace.Type == mm.TypeWhitespace &&
		runeLen(whitespace.Text) > 1 {
		column := whitespace.StartColumn + 1
		length := whitespace.EndColumn - column
		helpers.AddErrorContext(
			onError,
			startLine,
			strings.TrimSpace(text),
			delta > 0,
			delta < 0,
			rng(column, length),
			fixDelete(column, length),
		)
	}
}

var md019 = rule.Rule{
	Names:       []string{"MD019", "no-multiple-space-atx"},
	Description: "Multiple spaces after hash on atx style heading",
	Tags:        []string{"headings", "atx", "spaces"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading}, false) {
			if mdhelpers.GetHeadingStyle(heading) != "atx" {
				continue
			}

			md019md021ValidateHeadingSpaces(onError, heading, 1)
		}
	},
}

var md021 = rule.Rule{
	Names:       []string{"MD021", "no-multiple-space-closed-atx"},
	Description: "Multiple spaces inside hashes on closed atx style heading",
	Tags:        []string{"headings", "atx_closed", "spaces"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading}, false) {
			if mdhelpers.GetHeadingStyle(heading) != "atx_closed" {
				continue
			}

			md019md021ValidateHeadingSpaces(onError, heading, 1)
			md019md021ValidateHeadingSpaces(onError, heading, -1)
		}
	},
}
