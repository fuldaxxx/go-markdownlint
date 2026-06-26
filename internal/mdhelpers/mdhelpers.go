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

// Package mdhelpers ports helpers/micromark-helpers.cjs: utilities that operate
// on the micromark token tree (filtering, descendant/parent lookup, heading
// helpers, HTML tag info).
package mdhelpers

import (
	"regexp"
	"strings"

	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
)

// NonContentTokens is the set of token types that do not contain content.
var NonContentTokens = map[mm.TokenType]bool{
	mm.TypeBlockQuoteMarker:            true,
	mm.TypeBlockQuotePrefix:            true,
	mm.TypeBlockQuotePrefixWhitespace:  true,
	mm.TypeGfmFootnoteDefinitionIndent: true,
	mm.TypeLineEnding:                  true,
	mm.TypeLineEndingBlank:             true,
	mm.TypeLinePrefix:                  true,
	mm.TypeListItemIndent:              true,
	mm.TypeUndefinedReference:          true,
	mm.TypeUndefinedReferenceCollapsed: true,
	mm.TypeUndefinedReferenceFull:      true,
	mm.TypeUndefinedReferenceShortcut:  true,
}

// AddRangeToSet adds the inclusive integer range [start, end] to set.
func AddRangeToSet(set map[int]struct{}, start, end int) {
	for i := start; i <= end; i++ {
		set[i] = struct{}{}
	}
}

// FilterByPredicate walks tokens depth-first, collecting those for which
// allowed returns true. If transform is non-nil it replaces a token's children
// for the purpose of the walk.
func FilterByPredicate(
	tokens []*mm.Token,
	allowed func(*mm.Token) bool,
	transform func(*mm.Token) []*mm.Token,
) []*mm.Token {
	var result []*mm.Token

	type frame struct {
		array []*mm.Token
		index int
	}

	queue := []*frame{{array: tokens}}
	for len(queue) > 0 {
		cur := queue[len(queue)-1]
		if cur.index < len(cur.array) {
			tok := cur.array[cur.index]
			cur.index++

			if allowed(tok) {
				result = append(result, tok)
			}

			if len(tok.Children) > 0 {
				children := tok.Children
				if transform != nil {
					children = transform(tok)
				}

				queue = append(queue, &frame{array: children})
			}
		} else {
			queue = queue[:len(queue)-1]
		}
	}

	return result
}

// FilterByTypes returns tokens whose type is in types. When the flat list is
// available on the document it is used; otherwise the tree is walked. htmlFlow
// controls whether tokens inside HTML flow blocks are included.
func FilterByTypes(tokens []*mm.Token, types []mm.TokenType, htmlFlow bool) []*mm.Token {
	want := make(map[mm.TokenType]bool, len(types))
	for _, t := range types {
		want[t] = true
	}

	pred := func(tok *mm.Token) bool {
		return want[tok.Type] && (htmlFlow || !tok.InHTMLFlow())
	}

	return FilterByPredicate(tokens, pred, nil)
}

// FilterFlat filters a precomputed flat token list by type.
func FilterFlat(flat []*mm.Token, types []mm.TokenType, htmlFlow bool) []*mm.Token {
	want := make(map[mm.TokenType]bool, len(types))
	for _, t := range types {
		want[t] = true
	}

	var result []*mm.Token

	for _, tok := range flat {
		if want[tok.Type] && (htmlFlow || !tok.InHTMLFlow()) {
			result = append(result, tok)
		}
	}

	return result
}

// GetDescendantsByType returns descendants of parent(s) reached by following
// the given type path. Each step's type element may match one of several types.
func GetDescendantsByType(parents []*mm.Token, typePath [][]mm.TokenType) []*mm.Token {
	tokens := parents

	for _, types := range typePath {
		want := make(map[mm.TokenType]bool, len(types))
		for _, t := range types {
			want[t] = true
		}

		var next []*mm.Token

		for _, tok := range tokens {
			for _, c := range tok.Children {
				if want[c.Type] {
					next = append(next, c)
				}
			}
		}

		tokens = next
	}

	return tokens
}

// GetParentOfType returns the nearest ancestor of token whose type is in types.
func GetParentOfType(token *mm.Token, types []mm.TokenType) *mm.Token {
	want := make(map[mm.TokenType]bool, len(types))
	for _, t := range types {
		want[t] = true
	}

	cur := token.Parent
	for cur != nil && !want[cur.Type] {
		cur = cur.Parent
	}

	return cur
}

// GetHeadingLevel returns the level (1-6) of a heading token.
func GetHeadingLevel(heading *mm.Token) int {
	level := 1

	var seq *mm.Token

	for _, c := range heading.Children {
		if c.Type == mm.TypeAtxHeadingSequence || c.Type == mm.TypeSetextHeadingLine {
			seq = c
			break
		}
	}

	if seq == nil {
		return level
	}

	text := seq.Text
	if len(text) > 0 && text[0] == '#' {
		level = len(text)
		if level > 6 {
			level = 6
		}
	} else if len(text) > 0 && text[0] == '-' {
		level = 2
	}

	return level
}

// GetHeadingStyle returns the style of a heading token: "atx", "atx_closed", or
// "setext".
func GetHeadingStyle(heading *mm.Token) string {
	if heading.Type == mm.TypeSetextHeading {
		return "setext"
	}

	count := 0

	for _, c := range heading.Children {
		if c.Type == mm.TypeAtxHeadingSequence {
			count++
		}
	}

	if count == 1 {
		return "atx"
	}

	return "atx_closed"
}

var newLineRe = regexp.MustCompile(`\r\n?|\n`)

// GetHeadingText returns the text content of a heading token, excluding HTML
// and collapsing newlines to spaces.
func GetHeadingText(heading *mm.Token) string {
	descendants := GetDescendantsByType([]*mm.Token{heading}, [][]mm.TokenType{
		{mm.TypeAtxHeadingText, mm.TypeSetextHeadingText},
	})

	var sb strings.Builder

	for _, d := range descendants {
		for _, c := range d.Children {
			if c.Type != mm.TypeHTMLText {
				sb.WriteString(c.Text)
			}
		}
	}

	return newLineRe.ReplaceAllString(sb.String(), " ")
}

// HTMLTagInfo describes an HTML tag.
type HTMLTagInfo struct {
	Close bool
	Name  string
}

var htmlTagNameRe = regexp.MustCompile(`^<([^!>][^/\s>]*)`)

// GetHTMLTagInfo returns information about the tag in an htmlText token, or nil.
func GetHTMLTagInfo(token *mm.Token) *HTMLTagInfo {
	if token.Type != mm.TypeHTMLText {
		return nil
	}

	m := htmlTagNameRe.FindStringSubmatch(token.Text)
	if m == nil {
		return nil
	}

	name := m[1]

	close := strings.HasPrefix(name, "/")
	if close {
		name = name[1:]
	}

	return &HTMLTagInfo{Close: close, Name: name}
}

// GetBlockQuotePrefixText returns the blockquote prefix text for the given line,
// repeated count times.
func GetBlockQuotePrefixText(tokens []*mm.Token, lineNumber, count int) string {
	prefixes := FilterByTypes(
		tokens,
		[]mm.TokenType{mm.TypeBlockQuotePrefix, mm.TypeLinePrefix},
		false,
	)

	var sb strings.Builder

	for _, p := range prefixes {
		if p.StartLine == lineNumber {
			sb.WriteString(p.Text)
		}
	}

	one := strings.TrimRight(sb.String(), " \t") + "\n"

	return strings.Repeat(one, count)
}

// IsHTMLFlowComment reports whether token is an htmlFlow containing an HTML
// comment.
func IsHTMLFlowComment(token *mm.Token) bool {
	if token.Type != mm.TypeHTMLFlow {
		return false
	}

	text := token.Text
	if !strings.HasPrefix(text, "<!--") || !strings.HasSuffix(text, "-->") {
		return false
	}

	comment := text[4 : len(text)-3]

	return !strings.HasPrefix(comment, ">") &&
		!strings.HasPrefix(comment, "->") &&
		!strings.HasSuffix(comment, "-")
}

var docfxTabSyntaxRe = regexp.MustCompile(`^#tab/`)

// IsDocfxTab reports whether heading looks like a Docfx tab.
func IsDocfxTab(heading *mm.Token) bool {
	if heading == nil || heading.Type != mm.TypeAtxHeading {
		return false
	}

	headingTexts := GetDescendantsByType(
		[]*mm.Token{heading},
		[][]mm.TokenType{{mm.TypeAtxHeadingText}},
	)
	if len(headingTexts) == 1 && len(headingTexts[0].Children) == 1 &&
		headingTexts[0].Children[0].Type == mm.TypeLink {
		dests := FilterByTypes(
			headingTexts[0].Children[0].Children,
			[]mm.TokenType{mm.TypeResourceDestinationString},
			false,
		)

		return len(dests) == 1 && docfxTabSyntaxRe.MatchString(dests[0].Text)
	}

	return false
}
