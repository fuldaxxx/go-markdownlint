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
	"net/url"
	"regexp"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md054) }

var (
	md054BackslashEscapeRe = regexp.MustCompile(
		`\\([!"#$%&'()*+,\-./:;<=>?@\[\\\]^_` + "`" + `{|}~])`,
	)
	md054AutolinkDisallowedRe = regexp.MustCompile(`[ <>]`)
	md054LabelBracketRe       = regexp.MustCompile(`[\[\]]`)
	md054DestParenRe          = regexp.MustCompile(`[()]`)
)

func md054RemoveBackslashEscapes(text string) string {
	return md054BackslashEscapeRe.ReplaceAllString(text, "$1")
}

// md054AutolinkAble reports whether the destination parses as an absolute URL
// and contains no disallowed characters.
func md054AutolinkAble(destination string) bool {
	u, err := url.Parse(destination)
	// An absolute URL (with a scheme) is required.
	if err != nil || u.Scheme == "" {
		return false
	}

	return !md054AutolinkDisallowedRe.MatchString(destination)
}

var md054 = rule.Rule{
	Names:       []string{"MD054", "link-image-style"},
	Description: "Link and image style",
	Tags:        []string{"images", "links"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD054
		autolink := types.BoolOr(c.Autolink, true)
		inline := types.BoolOr(c.Inline, true)
		full := types.BoolOr(c.Full, true)
		collapsed := types.BoolOr(c.Collapsed, true)
		shortcut := types.BoolOr(c.Shortcut, true)

		urlInline := types.BoolOr(c.URLInline, true)
		if autolink && inline && full && collapsed && shortcut && urlInline {
			return
		}

		definitions := p.ReferenceLinkImageData().Definitions
		descText := func(token *mm.Token, path [][]mm.TokenType) (string, bool) {
			ds := mdhelpers.GetDescendantsByType([]*mm.Token{token}, path)
			if len(ds) > 0 {
				return ds[0].Text, true
			}

			return "", false
		}

		links := p.FilterByTypesCached(
			[]mm.TokenType{mm.TypeAutolink, mm.TypeImage, mm.TypeLink},
			false,
		)
		for _, link := range links {
			label := ""
			destination := ""
			image := link.Type == mm.TypeImage
			isError := false

			if link.Type == mm.TypeAutolink {
				dest, hasDest := descText(
					link,
					[][]mm.TokenType{{mm.TypeAutolinkEmail, mm.TypeAutolinkProtocol}},
				)
				destination = dest
				label = destination
				isError = !autolink && hasDest && dest != ""
			} else {
				lbl, _ := descText(link, [][]mm.TokenType{{mm.TypeLabel}, {mm.TypeLabelText}})
				label = lbl

				dest, hasDest := descText(link, [][]mm.TokenType{
					{mm.TypeResource},
					{mm.TypeResourceDestination},
					{mm.TypeResourceDestinationLiteral, mm.TypeResourceDestinationRaw},
					{mm.TypeResourceDestinationString},
				})
				if hasDest {
					destination = dest
					title, hasTitle := descText(
						link,
						[][]mm.TokenType{
							{mm.TypeResource},
							{mm.TypeResourceTitle},
							{mm.TypeResourceTitleString},
						},
					)
					_ = title
					isError = !inline || (!urlInline &&
						autolink &&
						!image &&
						!hasTitle &&
						(label == destination) &&
						md054AutolinkAble(destination))
				} else {
					references := mdhelpers.GetDescendantsByType(
						[]*mm.Token{link},
						[][]mm.TokenType{{mm.TypeReference}},
					)
					isShortcut := len(references) == 0
					referenceString, hasReferenceString := descText(
						link,
						[][]mm.TokenType{{mm.TypeReference}, {mm.TypeReferenceString}},
					)
					isCollapsed := !hasReferenceString

					key := referenceString
					if !hasReferenceString {
						key = label
					}

					if def, ok := definitions[key]; ok {
						destination = def.Destination
					} else {
						destination = ""
					}

					isError = destination != "" &&
						((isShortcut && !shortcut) || (!isShortcut && isCollapsed && !collapsed) || (!isShortcut && !isCollapsed && !full))
				}
			}

			if isError {
				var (
					r       *[2]int
					fixInfo *types.FixInfo
				)

				if link.StartLine == link.EndLine {
					column := link.StartColumn
					length := link.EndColumn - link.StartColumn
					r = rng(column, length)
					insertText := ""
					canInline := inline && label != ""

					canAutolink := autolink && !image && md054AutolinkAble(destination)
					if canInline && (urlInline || !canAutolink) {
						prefix := ""
						if image {
							prefix = "!"
						}

						escapedLabel := md054LabelBracketRe.ReplaceAllString(label, `\$0`)
						escapedDestination := md054DestParenRe.ReplaceAllString(destination, `\$0`)
						insertText = prefix + "[" + escapedLabel + "](" + escapedDestination + ")"
					} else if canAutolink {
						insertText = "<" + md054RemoveBackslashEscapes(destination) + ">"
					}

					if insertText != "" {
						fixInfo = fixReplace(column, length, insertText)
					}
				}

				helpers.AddErrorContext(
					onError,
					link.StartLine,
					helpers.NextLinesRe.ReplaceAllString(link.Text, ""),
					false,
					false,
					r,
					fixInfo,
				)
			}
		}
	},
}
