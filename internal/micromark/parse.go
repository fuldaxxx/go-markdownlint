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

package micromark

import (
	"regexp"
	"strings"
)

// Parse parses a Markdown document into a micromark-compatible token tree.
// Front matter must already have been removed by the caller. Line and column
// values are 1-based; columns count runes (a tab counts as one column).
func Parse(markdown string) *Document {
	lines := splitKeepEnds(markdown)
	// First pass: parse blocks to discover link reference definition labels.
	// Inline parsing during this pass over-creates link/image tokens, but we only
	// use the resulting `definition` tokens to build the defined-label set.
	pre := &parser{lines: lines}
	preRoot := &Token{
		Type:        TypeData,
		Text:        "ROOT",
		StartLine:   -1,
		StartColumn: -1,
		EndLine:     -1,
		EndColumn:   -1,
	}
	pre.parseBlocks(preRoot, 0, len(pre.lines))
	defined := collectDefinedLabels(preRoot.Children)

	// Second pass: re-parse with the defined-label set so inline parsing can
	// decide which bracket forms are real links/images vs. undefined references.
	p := &parser{lines: lines, definedLabels: defined}
	root := &Token{
		Type:        TypeData,
		Text:        "ROOT",
		StartLine:   -1,
		StartColumn: -1,
		EndLine:     -1,
		EndColumn:   -1,
	}
	p.parseBlocks(root, 0, len(p.lines))

	doc := &Document{Children: root.Children}
	for _, c := range root.Children {
		c.Parent = nil
	}

	doc.Flat = flatten(root.Children)
	// Normalize to micromark's exclusive end-column convention: EndColumn points
	// one past the last character of the token (so length == EndColumn-StartColumn).
	// The block/inline tokenizers build inclusive end columns; shift them by one.
	for _, t := range doc.Flat {
		t.EndColumn++
	}

	return doc
}

// line is one source line with its content (excluding the newline) and the
// newline sequence that followed it ("" for the last line if unterminated).
type line struct {
	text string
	eol  string
}

type parser struct {
	lines []line
	// definedLabels is the set of normalized link reference definition labels
	// discovered in a first parsing pass. nil during the first pass.
	definedLabels map[string]bool
}

var refWhitespaceRe = regexp.MustCompile(`\s+`)

// normalizeLabel normalizes a reference/definition label the same way
// helpers.GetReferenceLinkImageData does: lowercase, trim, collapse internal
// whitespace runs to a single space.
func normalizeLabel(s string) string {
	return refWhitespaceRe.ReplaceAllString(strings.TrimSpace(strings.ToLower(s)), " ")
}

// collectDefinedLabels walks a parsed token tree and returns the set of
// normalized definition labels (from definitionLabelString tokens).
func collectDefinedLabels(tokens []*Token) map[string]bool {
	set := map[string]bool{}

	var walk func([]*Token)

	walk = func(ts []*Token) {
		for _, t := range ts {
			if t.Type == TypeDefinitionLabelString {
				set[normalizeLabel(t.Text)] = true
			}

			walk(t.Children)
		}
	}
	walk(tokens)

	return set
}

var newLineSplitRe = regexp.MustCompile(`\r\n|\r|\n`)

func splitKeepEnds(s string) []line {
	var out []line

	i := 0
	for i < len(s) {
		loc := newLineSplitRe.FindStringIndex(s[i:])
		if loc == nil {
			out = append(out, line{text: s[i:], eol: ""})
			return out
		}

		out = append(out, line{text: s[i : i+loc[0]], eol: s[i+loc[0] : i+loc[1]]})
		i += loc[1]
	}
	// Trailing newline => a final empty line with no eol (matches split semantics
	// where the document text ends with a newline but no further content).
	if len(s) == 0 {
		out = append(out, line{})
	}

	return out
}

// flatten returns a depth-first flat list of tokens.
func flatten(tokens []*Token) []*Token {
	var (
		out  []*Token
		walk func([]*Token)
	)

	walk = func(ts []*Token) {
		for _, t := range ts {
			out = append(out, t)
			walk(t.Children)
		}
	}
	walk(tokens)

	return out
}

// runeLen returns the number of runes in s.
func runeLen(s string) int { return len([]rune(s)) }

// leadingSpaces returns the count of leading space/tab characters (as columns,
// tab = 1).
func leadingSpaces(s string) int {
	n := 0

	for _, r := range s {
		if r == ' ' || r == '\t' {
			n++
		} else {
			break
		}
	}

	return n
}

// addChild appends child to parent and sets the parent pointer.
func addChild(parent, child *Token) {
	child.Parent = parent
	parent.Children = append(parent.Children, child)
}

var (
	atxRe          = regexp.MustCompile(`^(\s*)(#{1,6})(\s+.*?|\s*)$`)
	thematicRe     = regexp.MustCompile(`^ {0,3}((\* *){3,}|(- *){3,}|(_ *){3,})$`)
	fenceOpenRe    = regexp.MustCompile("^(\\s*)(`{3,}|~{3,})(.*)$")
	setextRe       = regexp.MustCompile(`^ {0,3}(=+|-+)\s*$`)
	blockquoteRe   = regexp.MustCompile(`^ {0,3}>`)
	htmlBlockRe    = regexp.MustCompile(`(?i)^ {0,3}<(/?)([a-z][a-z0-9-]*|!--)`)
	bulletRe       = regexp.MustCompile(`^(\s*)([-*+])(\s+)(.*)$`)
	orderedRe      = regexp.MustCompile(`^(\s*)(\d{1,9})([.)])(\s+)(.*)$`)
	bulletEmptyRe  = regexp.MustCompile(`^(\s*)([-*+])(\s*)$`)
	orderedEmptyRe = regexp.MustCompile(`^(\s*)(\d{1,9})([.)])(\s*)$`)
	defRe          = regexp.MustCompile(`^ {0,3}\[([^\]]+)\]:\s*(\S+)(?:\s+(.*))?$`)
	tableDelimRe   = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$`)
)

func isBlank(s string) bool { return strings.TrimSpace(s) == "" }
