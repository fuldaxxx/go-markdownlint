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
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md053) }

var md053LinkReferenceDefinitionRe = regexp.MustCompile(`^ {0,3}\[([^\]]*[^\\])\]:`)

var md053 = rule.Rule{
	Names:       []string{"MD053", "link-image-reference-definitions"},
	Description: "Link and image reference definitions should be needed",
	Tags:        []string{"images", "links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD053
		ignored := map[string]struct{}{}

		if ig := c.IgnoredDefinitions; ig != nil {
			for _, l := range ig {
				ignored[l] = struct{}{}
			}
		} else {
			ignored["//"] = struct{}{}
		}

		lines := p.Lines
		data := p.ReferenceLinkImageData()

		singleLineDefinition := func(line string) bool {
			return len(
				strings.TrimSpace(md053LinkReferenceDefinitionRe.ReplaceAllString(line, "")),
			) > 0
		}
		deleteFixInfo := fixDelete(0, -1)

		// Look for unused link references (unreferenced by any link/image).
		for label, def := range data.Definitions {
			lineIndex := def.LineIndex

			if _, ig := ignored[label]; ig {
				continue
			}

			if _, ok := data.References[label]; ok {
				continue
			}

			if _, ok := data.Shortcuts[label]; ok {
				continue
			}

			line := lines[lineIndex]

			var fix *types.FixInfo
			if singleLineDefinition(line) {
				fix = deleteFixInfo
			}

			helpers.AddError(
				onError,
				lineIndex+1,
				"Unused link or image reference definition: \""+label+"\"",
				helpers.Ellipsify(line, false, false),
				rng(1, runeLen(line)),
				fix,
			)
		}

		// Look for duplicate link references (defined more than once).
		for _, dup := range data.DuplicateDefinitions {
			label := dup.Label
			lineIndex := dup.LineIndex

			if _, ig := ignored[label]; ig {
				continue
			}

			line := lines[lineIndex]

			var fix *types.FixInfo
			if singleLineDefinition(line) {
				fix = deleteFixInfo
			}

			helpers.AddError(
				onError,
				lineIndex+1,
				"Duplicate link or image reference definition: \""+label+"\"",
				helpers.Ellipsify(line, false, false),
				rng(1, runeLen(line)),
				fix,
			)
		}
	},
}
