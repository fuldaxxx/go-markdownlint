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
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md052) }

var md052 = rule.Rule{
	Names:       []string{"MD052", "reference-links-images"},
	Description: "Reference links and images should use a label that is defined",
	Tags:        []string{"images", "links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD052
		shortcutSyntax := types.BoolOr(c.ShortcutSyntax, false)
		ignoredLabels := map[string]struct{}{}

		if il := c.IgnoredLabels; il != nil {
			for _, l := range il {
				ignoredLabels[l] = struct{}{}
			}
		} else {
			ignoredLabels["x"] = struct{}{}
		}

		data := p.ReferenceLinkImageData()

		type entry struct {
			label string
			datas [][3]int
		}

		var entries []entry
		for label, datas := range data.References {
			entries = append(entries, entry{label, datas})
		}

		if shortcutSyntax {
			for label, datas := range data.Shortcuts {
				entries = append(entries, entry{label, datas})
			}
		}

		for _, e := range entries {
			if _, defined := data.Definitions[e.label]; defined {
				continue
			}

			if _, ignored := ignoredLabels[e.label]; ignored {
				continue
			}

			for _, d := range e.datas {
				lineIndex, index, length := d[0], d[1], d[2]
				// Context will be incomplete if reporting for a multi-line link.
				line := []rune(p.Lines[lineIndex])

				end := index + length
				if end > len(line) {
					end = len(line)
				}

				start := index
				if start > len(line) {
					start = len(line)
				}

				context := string(line[start:end])
				helpers.AddError(
					onError,
					lineIndex+1,
					"Missing link or image reference definition: \""+e.label+"\"",
					context,
					rng(index+1, runeLen(context)),
					nil,
				)
			}
		}
	},
}
