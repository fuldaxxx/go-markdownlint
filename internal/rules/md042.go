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

func init() { register(&md042) }

var md042 = rule.Rule{
	Names:       []string{"MD042", "no-empty-links"},
	Description: "No empty links",
	Tags:        []string{"links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		definitions := p.ReferenceLinkImageData().Definitions
		isReferenceDefinitionHash := func(token *mm.Token) bool {
			def, ok := definitions[strings.TrimSpace(token.Text)]
			return ok && def.Destination == "#"
		}

		links := p.FilterByTypesCached([]mm.TokenType{mm.TypeLink}, false)
		for _, link := range links {
			labelText := mdhelpers.GetDescendantsByType(
				[]*mm.Token{link},
				[][]mm.TokenType{{mm.TypeLabel}, {mm.TypeLabelText}},
			)
			reference := mdhelpers.GetDescendantsByType(
				[]*mm.Token{link},
				[][]mm.TokenType{{mm.TypeReference}},
			)
			resource := mdhelpers.GetDescendantsByType(
				[]*mm.Token{link},
				[][]mm.TokenType{{mm.TypeResource}},
			)
			referenceString := mdhelpers.GetDescendantsByType(
				reference,
				[][]mm.TokenType{{mm.TypeReferenceString}},
			)
			resourceDestinationString := mdhelpers.GetDescendantsByType(resource, [][]mm.TokenType{
				{mm.TypeResourceDestination},
				{mm.TypeResourceDestinationLiteral, mm.TypeResourceDestinationRaw},
				{mm.TypeResourceDestinationString},
			})
			hasLabelText := len(labelText) > 0
			hasReference := len(reference) > 0
			hasResource := len(resource) > 0
			hasReferenceString := len(referenceString) > 0
			hasResourceDestinationString := len(resourceDestinationString) > 0

			error := false
			if hasLabelText &&
				((!hasReference && !hasResource) || (hasReference && !hasReferenceString)) {
				error = isReferenceDefinitionHash(labelText[0])
			} else if hasReferenceString &&
				!hasResourceDestinationString {
				error = isReferenceDefinitionHash(referenceString[0])
			} else if !hasReferenceString &&
				hasResourceDestinationString {
				error = strings.TrimSpace(resourceDestinationString[0].Text) == "#"
			} else if !hasReferenceString &&
				!hasResourceDestinationString {
				error = true
			}

			if error {
				helpers.AddErrorContext(
					onError,
					link.StartLine,
					link.Text,
					false,
					false,
					rng(link.StartColumn, link.EndColumn-link.StartColumn),
					nil,
				)
			}
		}
	},
}
