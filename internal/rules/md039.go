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

func init() { register(&md039) }

var (
	md039StartRe     = regexp.MustCompile(`^[^\S\r\n]+`)
	md039EndRe       = regexp.MustCompile(`[^\S\r\n]+$`)
	md039WhitespRe   = regexp.MustCompile(`\s+`)
	md039TrimStartRe = regexp.MustCompile(`^\s+`)
	md039TrimEndRe   = regexp.MustCompile(`\s+$`)
)

var md039 = rule.Rule{
	Names:       []string{"MD039", "no-space-in-links"},
	Description: "Spaces inside link text",
	Tags:        []string{"whitespace", "links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		addLabelSpaceError := func(label, labelText *mm.Token, isStart bool) {
			var re *regexp.Regexp
			if isStart {
				re = md039StartRe
			} else {
				re = md039EndRe
			}

			match := re.FindString(labelText.Text)

			var (
				r   *[2]int
				fix *types.FixInfo
			)

			if match != "" {
				var col int
				if isStart {
					col = labelText.StartColumn
				} else {
					col = labelText.EndColumn - runeLen(match)
				}

				r = rng(col, runeLen(match))
				fix = fixDelete(col, runeLen(match))
			}

			var lineNumber int
			if isStart {
				lineNumber = labelText.StartLine
				if match == "" {
					lineNumber++
				}
			} else {
				lineNumber = labelText.EndLine
				if match == "" {
					lineNumber--
				}
			}

			context := md039WhitespRe.ReplaceAllString(label.Text, " ")
			helpers.AddErrorContext(
				onError,
				lineNumber,
				context,
				isStart,
				!isStart,
				r,
				fix,
			)
		}

		var labels []*mm.Token

		for _, label := range p.FilterByTypesCached([]mm.TokenType{mm.TypeLabel}, false) {
			if label.Parent != nil && label.Parent.Type == mm.TypeLink {
				labels = append(labels, label)
			}
		}

		for _, label := range labels {
			for _, labelText := range label.Children {
				if labelText.Type != mm.TypeLabelText {
					continue
				}

				if md039TrimStartRe.MatchString(labelText.Text) {
					addLabelSpaceError(label, labelText, true)
				}

				if md039TrimEndRe.MatchString(labelText.Text) {
					addLabelSpaceError(label, labelText, false)
				}
			}
		}
	},
}
