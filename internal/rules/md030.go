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
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md030) }

var md030 = rule.Rule{
	Names:       []string{"MD030", "list-marker-space"},
	Description: "Spaces after list markers",
	Tags:        []string{"ol", "ul", "whitespace"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD030
		ulSingle := types.IntOr(c.UlSingle, 1)
		olSingle := types.IntOr(c.OlSingle, 1)
		ulMulti := types.IntOr(c.UlMulti, 1)

		olMulti := types.IntOr(c.OlMulti, 1)

		for _, list := range p.FilterByTypesCached([]mm.TokenType{mm.TypeListOrdered, mm.TypeListUnordered}, false) {
			ordered := list.Type == mm.TypeListOrdered

			var listItemPrefixes []*mm.Token

			for _, token := range list.Children {
				if token.Type == mm.TypeListItemPrefix {
					listItemPrefixes = append(listItemPrefixes, token)
				}
			}

			allSingleLine := (list.EndLine - list.StartLine + 1) == len(listItemPrefixes)

			var expectedSpaces int

			if ordered {
				if allSingleLine {
					expectedSpaces = olSingle
				} else {
					expectedSpaces = olMulti
				}
			} else {
				if allSingleLine {
					expectedSpaces = ulSingle
				} else {
					expectedSpaces = ulMulti
				}
			}

			for _, listItemPrefix := range listItemPrefixes {
				rangeVal := rng(
					listItemPrefix.StartColumn,
					listItemPrefix.EndColumn-listItemPrefix.StartColumn,
				)
				for _, listItemPrefixWhitespace := range listItemPrefix.Children {
					if listItemPrefixWhitespace.Type != mm.TypeListItemPrefixWhitespace {
						continue
					}

					actualSpaces := listItemPrefixWhitespace.EndColumn - listItemPrefixWhitespace.StartColumn
					fixInfo := fixReplace(
						listItemPrefixWhitespace.StartColumn,
						actualSpaces,
						strings.Repeat(" ", expectedSpaces),
					)
					helpers.AddErrorDetailIf(
						onError,
						listItemPrefixWhitespace.StartLine,
						expectedSpaces,
						actualSpaces,
						"",
						"",
						rangeVal,
						fixInfo,
					)
				}
			}
		}
	},
}
