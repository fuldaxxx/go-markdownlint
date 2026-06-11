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
	"strings"
	"testing"
)

// --- Test helpers ---

// findByType returns every token in the flattened document with the given type,
// preserving document (depth-first) order.
func findByType(doc *Document, typ TokenType) []*Token {
	var out []*Token
	for _, t := range doc.Flat {
		if t.Type == typ {
			out = append(out, t)
		}
	}
	return out
}

// firstByType returns the first token of the given type, or nil.
func firstByType(doc *Document, typ TokenType) *Token {
	ts := findByType(doc, typ)
	if len(ts) == 0 {
		return nil
	}
	return ts[0]
}

// countByType counts tokens of the given type.
func countByType(doc *Document, typ TokenType) int {
	return len(findByType(doc, typ))
}

// requireType fails unless exactly one token of typ exists, returning it.
func requireType(t *testing.T, doc *Document, typ TokenType) *Token {
	t.Helper()
	ts := findByType(doc, typ)
	if len(ts) != 1 {
		t.Fatalf("expected exactly 1 %q token, found %d", typ, len(ts))
	}
	return ts[0]
}

// assertSpanStart checks 1-based start line/column of a token.
func assertSpanStart(t *testing.T, tok *Token, line, col int) {
	t.Helper()
	if tok == nil {
		t.Fatalf("nil token (type assertion)")
	}
	if tok.StartLine != line || tok.StartColumn != col {
		t.Errorf("token %q: got start L%d C%d, want L%d C%d", tok.Type, tok.StartLine, tok.StartColumn, line, col)
	}
}

// assertText checks a token's Text.
func assertText(t *testing.T, tok *Token, want string) {
	t.Helper()
	if tok == nil {
		t.Fatalf("nil token, want text %q", want)
	}
	if tok.Text != want {
		t.Errorf("token %q: got text %q, want %q", tok.Type, tok.Text, want)
	}
}

// --- ATX headings ---

func TestATXHeadingBasic(t *testing.T) {
	doc := Parse("# Heading\n")
	h := requireType(t, doc, TypeAtxHeading)
	assertSpanStart(t, h, 1, 1)
	assertText(t, h, "# Heading")

	seq := requireType(t, doc, TypeAtxHeadingSequence)
	assertSpanStart(t, seq, 1, 1)
	assertText(t, seq, "#")

	// Whitespace token sits between the opening sequence and the heading text.
	ws := requireType(t, doc, TypeWhitespace)
	assertSpanStart(t, ws, 1, 2)
	assertText(t, ws, " ")

	htxt := requireType(t, doc, TypeAtxHeadingText)
	assertSpanStart(t, htxt, 1, 3)
	assertText(t, htxt, "Heading")

	// EndColumn is exclusive (one past the last char). "# Heading" len 9, start 1.
	if h.EndColumn != 10 {
		t.Errorf("atxHeading EndColumn: got %d, want 10 (exclusive)", h.EndColumn)
	}
}

func TestATXHeadingClosed(t *testing.T) {
	doc := Parse("## Closed ##\n")
	seqs := findByType(doc, TypeAtxHeadingSequence)
	if len(seqs) != 2 {
		t.Fatalf("closed ATX: want 2 sequences (open+close), got %d", len(seqs))
	}
	assertText(t, seqs[0], "##")
	assertSpanStart(t, seqs[0], 1, 1)
	assertText(t, seqs[1], "##")
	// Closing sequence starts at column 11.
	assertSpanStart(t, seqs[1], 1, 11)

	htxt := requireType(t, doc, TypeAtxHeadingText)
	assertText(t, htxt, "Closed")

	// There is a whitespace token between text and the closing sequence too.
	wss := findByType(doc, TypeWhitespace)
	if len(wss) != 2 {
		t.Fatalf("closed ATX: want 2 whitespace tokens, got %d", len(wss))
	}
}

func TestATXHeadingIndented(t *testing.T) {
	// Indentation < 4 spaces: emitted as a linePrefix sibling preceding the heading.
	doc := Parse("   # Indented\n")
	lp := requireType(t, doc, TypeLinePrefix)
	assertSpanStart(t, lp, 1, 1)
	assertText(t, lp, "   ")

	h := requireType(t, doc, TypeAtxHeading)
	assertSpanStart(t, h, 1, 4)

	// linePrefix precedes the heading in flat order.
	var lpIdx, hIdx = -1, -1
	for i, tk := range doc.Flat {
		if tk == lp {
			lpIdx = i
		}
		if tk == h {
			hIdx = i
		}
	}
	if lpIdx < 0 || hIdx < 0 || lpIdx >= hIdx {
		t.Errorf("linePrefix (idx %d) should precede atxHeading (idx %d)", lpIdx, hIdx)
	}
	// linePrefix is a sibling of the heading, not a child.
	if lp.Parent != nil && lp.Parent.Type == TypeAtxHeading {
		t.Errorf("linePrefix should be a sibling of atxHeading, not its child")
	}
}

func TestATXHeadingEmpty(t *testing.T) {
	// "# " with only a hash and trailing space: no heading text token.
	doc := Parse("# \n")
	if requireType(t, doc, TypeAtxHeading) == nil {
		t.Fatal("expected atxHeading")
	}
	if countByType(doc, TypeAtxHeadingText) != 0 {
		t.Errorf("empty heading should have no atxHeadingText")
	}

	doc2 := Parse("#\n")
	requireType(t, doc2, TypeAtxHeading)
	if countByType(doc2, TypeAtxHeadingText) != 0 {
		t.Errorf("bare '#' heading should have no atxHeadingText")
	}
	if countByType(doc2, TypeWhitespace) != 0 {
		t.Errorf("bare '#' heading should have no whitespace token")
	}
}

func TestATXHeadingLevels(t *testing.T) {
	for n := 1; n <= 6; n++ {
		hashes := strings.Repeat("#", n)
		doc := Parse(hashes + " Title\n")
		seq := firstByType(doc, TypeAtxHeadingSequence)
		assertText(t, seq, hashes)
	}
	// 7 hashes is not a heading (becomes a paragraph).
	doc := Parse("####### Too many\n")
	if countByType(doc, TypeAtxHeading) != 0 {
		t.Errorf("7 hashes should not be an ATX heading")
	}
	requireType(t, doc, TypeParagraph)
}

// --- Setext headings ---

func TestSetextHeading(t *testing.T) {
	doc := Parse("Setext\n======\n")
	h := requireType(t, doc, TypeSetextHeading)
	assertSpanStart(t, h, 1, 1)
	if h.EndLine != 2 {
		t.Errorf("setext heading EndLine: got %d, want 2", h.EndLine)
	}
	htxt := requireType(t, doc, TypeSetextHeadingText)
	assertSpanStart(t, htxt, 1, 1)

	uln := requireType(t, doc, TypeSetextHeadingLine)
	assertSpanStart(t, uln, 2, 1)
	assertText(t, uln, "======")
}

func TestSetextHeadingDash(t *testing.T) {
	// NOTE: "Title\n---\n" is NOT a dash setext heading in this parser. The "---"
	// line matches the thematic-break rule first, so the result is a paragraph
	// followed by a thematicBreak (a CommonMark divergence: CommonMark prefers the
	// setext interpretation when there is a preceding paragraph line).
	doc := Parse("Title\n---\n")
	if countByType(doc, TypeSetextHeading) != 0 {
		t.Errorf("dash after paragraph is treated as thematicBreak, not setext")
	}
	requireType(t, doc, TypeParagraph)
	requireType(t, doc, TypeThematicBreak)
}

func TestSetextHeadingMultiline(t *testing.T) {
	doc := Parse("line one\nline two\n======\n")
	htxt := requireType(t, doc, TypeSetextHeadingText)
	if htxt.EndLine != 2 {
		t.Errorf("multiline setext text EndLine: got %d, want 2", htxt.EndLine)
	}
	// A lineEnding joins the two text lines.
	if countByType(doc, TypeLineEnding) != 1 {
		t.Errorf("expected 1 lineEnding in multiline setext text")
	}
}

// --- Thematic breaks ---

func TestThematicBreak(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"stars", "***\n"},
		{"dashes", "---\n"},
		{"underscores", "___\n"},
		{"spaced", "* * *\n"},
		{"many", "----------\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := Parse(tc.in)
			tb := requireType(t, doc, TypeThematicBreak)
			assertSpanStart(t, tb, 1, 1)
		})
	}
}

// --- Fenced code ---

func TestFencedCodeBacktick(t *testing.T) {
	doc := Parse("```go\ncode line\n```\n")
	code := requireType(t, doc, TypeCodeFenced)
	assertSpanStart(t, code, 1, 1)

	fences := findByType(doc, TypeCodeFencedFence)
	if len(fences) != 2 {
		t.Fatalf("want 2 fences (open+close), got %d", len(fences))
	}
	assertSpanStart(t, fences[0], 1, 1)
	assertText(t, fences[0], "```go")
	assertSpanStart(t, fences[1], 3, 1)

	seqs := findByType(doc, TypeCodeFencedFenceSeq)
	if len(seqs) != 1 {
		// Only the opening fence emits a sequence in this parser.
		t.Errorf("want 1 codeFencedFenceSequence, got %d", len(seqs))
	}
	assertText(t, seqs[0], "```")

	info := requireType(t, doc, TypeCodeFencedFenceInfo)
	assertText(t, info, "go")
	assertSpanStart(t, info, 1, 4)

	val := requireType(t, doc, TypeCodeFlowValue)
	assertText(t, val, "code line")
	assertSpanStart(t, val, 2, 1)
}

func TestFencedCodeTilde(t *testing.T) {
	doc := Parse("~~~\nx\n~~~\n")
	requireType(t, doc, TypeCodeFenced)
	seq := firstByType(doc, TypeCodeFencedFenceSeq)
	assertText(t, seq, "~~~")
}

func TestFencedCodeNoInfo(t *testing.T) {
	doc := Parse("```\nplain\n```\n")
	requireType(t, doc, TypeCodeFenced)
	if countByType(doc, TypeCodeFencedFenceInfo) != 0 {
		t.Errorf("fence with no info should have no info token")
	}
}

// --- Indented code ---

func TestIndentedCode(t *testing.T) {
	doc := Parse("    indented code\n")
	code := requireType(t, doc, TypeCodeIndented)
	assertSpanStart(t, code, 1, 1)
	val := requireType(t, doc, TypeCodeFlowValue)
	assertText(t, val, "indented code")
	assertSpanStart(t, val, 1, 5)
}

func TestIndentedCodeMultiline(t *testing.T) {
	doc := Parse("    line1\n    line2\n")
	vals := findByType(doc, TypeCodeFlowValue)
	if len(vals) != 2 {
		t.Fatalf("want 2 codeFlowValue, got %d", len(vals))
	}
	assertText(t, vals[0], "line1")
	assertText(t, vals[1], "line2")
}

// --- Lists ---

func TestUnorderedList(t *testing.T) {
	doc := Parse("- item\n")
	requireType(t, doc, TypeListUnordered)
	prefix := requireType(t, doc, TypeListItemPrefix)
	assertSpanStart(t, prefix, 1, 1)
	assertText(t, prefix, "- ")

	marker := requireType(t, doc, TypeListItemMarker)
	assertText(t, marker, "-")
	assertSpanStart(t, marker, 1, 1)

	pw := requireType(t, doc, TypeListItemPrefixWhitespace)
	assertText(t, pw, " ")
	assertSpanStart(t, pw, 1, 2)

	// content is a paragraph inside the list.
	requireType(t, doc, TypeParagraph)

	// No listItemValue for unordered.
	if countByType(doc, TypeListItemValue) != 0 {
		t.Errorf("unordered list should not have listItemValue")
	}
}

func TestOrderedList(t *testing.T) {
	doc := Parse("1. one\n")
	requireType(t, doc, TypeListOrdered)
	val := requireType(t, doc, TypeListItemValue)
	assertText(t, val, "1")
	assertSpanStart(t, val, 1, 1)

	marker := requireType(t, doc, TypeListItemMarker)
	assertText(t, marker, ".")
	assertSpanStart(t, marker, 1, 2)

	pw := requireType(t, doc, TypeListItemPrefixWhitespace)
	assertSpanStart(t, pw, 1, 3)
}

func TestOrderedListParen(t *testing.T) {
	doc := Parse("3) three\n")
	val := requireType(t, doc, TypeListItemValue)
	assertText(t, val, "3")
	marker := requireType(t, doc, TypeListItemMarker)
	assertText(t, marker, ")")
}

func TestListMultipleItems(t *testing.T) {
	doc := Parse("- a\n- b\n")
	if countByType(doc, TypeListItemPrefix) != 2 {
		t.Errorf("want 2 list item prefixes")
	}
	if countByType(doc, TypeListUnordered) != 1 {
		t.Errorf("two same-type items form one list")
	}
	prefixes := findByType(doc, TypeListItemPrefix)
	assertSpanStart(t, prefixes[1], 2, 1)
}

func TestListNestedContent(t *testing.T) {
	// Continuation line indented to the content column is part of the same item.
	doc := Parse("- first\n  still first\n")
	paras := findByType(doc, TypeParagraph)
	if len(paras) != 1 {
		t.Fatalf("want 1 paragraph spanning the two lines, got %d", len(paras))
	}
	if paras[0].EndLine != 2 {
		t.Errorf("continuation paragraph EndLine: got %d, want 2", paras[0].EndLine)
	}
}

func TestListIndentedPrefix(t *testing.T) {
	// A nested/indented list item carries a linePrefix child for its indent.
	doc := Parse("  - indented item\n")
	prefix := requireType(t, doc, TypeListItemPrefix)
	lp := firstByType(doc, TypeLinePrefix)
	if lp == nil {
		t.Fatal("expected linePrefix for indented list item")
	}
	assertText(t, lp, "  ")
	if lp.Parent != prefix {
		t.Errorf("linePrefix should be child of listItemPrefix")
	}
}

// --- Block quotes ---

func TestBlockQuote(t *testing.T) {
	doc := Parse("> quote\n")
	requireType(t, doc, TypeBlockQuote)
	prefix := requireType(t, doc, TypeBlockQuotePrefix)
	assertSpanStart(t, prefix, 1, 1)
	assertText(t, prefix, "> ")

	marker := requireType(t, doc, TypeBlockQuoteMarker)
	assertText(t, marker, ">")
	assertSpanStart(t, marker, 1, 1)

	pw := requireType(t, doc, TypeBlockQuotePrefixWhitespace)
	assertText(t, pw, " ")
	assertSpanStart(t, pw, 1, 2)

	requireType(t, doc, TypeParagraph)
}

func TestBlockQuoteExtraSpace(t *testing.T) {
	// An extra space after the "> " prefix becomes a linePrefix sibling
	// (MD027 relies on this).
	doc := Parse(">  extra\n")
	lp := firstByType(doc, TypeLinePrefix)
	if lp == nil {
		t.Fatal("expected linePrefix for extra space in blockquote")
	}
	assertText(t, lp, " ")
	assertSpanStart(t, lp, 1, 3)
}

func TestBlockQuoteMultiline(t *testing.T) {
	doc := Parse("> a\n> b\n")
	if countByType(doc, TypeBlockQuotePrefix) != 2 {
		t.Errorf("want 2 blockquote prefixes")
	}
	if countByType(doc, TypeBlockQuote) != 1 {
		t.Errorf("two quoted lines form one blockquote")
	}
}

func TestBlockQuoteNested(t *testing.T) {
	doc := Parse("> > deep\n")
	if countByType(doc, TypeBlockQuote) != 2 {
		t.Errorf("want 2 nested blockquotes, got %d", countByType(doc, TypeBlockQuote))
	}
}

// --- Paragraphs ---

func TestParagraph(t *testing.T) {
	doc := Parse("para text\n")
	p := requireType(t, doc, TypeParagraph)
	assertSpanStart(t, p, 1, 1)
	d := requireType(t, doc, TypeData)
	assertText(t, d, "para text")
}

func TestParagraphMultiline(t *testing.T) {
	doc := Parse("one\ntwo\n")
	p := requireType(t, doc, TypeParagraph)
	if p.EndLine != 2 {
		t.Errorf("multiline paragraph EndLine: got %d, want 2", p.EndLine)
	}
	if countByType(doc, TypeLineEnding) != 1 {
		t.Errorf("expected 1 lineEnding between paragraph lines")
	}
}

func TestParagraphsSeparatedByBlank(t *testing.T) {
	doc := Parse("a\n\n\nb\n")
	paras := findByType(doc, TypeParagraph)
	if len(paras) != 2 {
		t.Fatalf("want 2 paragraphs, got %d", len(paras))
	}
	assertSpanStart(t, paras[0], 1, 1)
	assertSpanStart(t, paras[1], 4, 1)
}

// --- Definitions ---

func TestDefinition(t *testing.T) {
	doc := Parse("[ref]: /url\n")
	def := requireType(t, doc, TypeDefinition)
	assertSpanStart(t, def, 1, 1)

	ls := requireType(t, doc, TypeDefinitionLabelString)
	assertText(t, ls, "ref")
	assertSpanStart(t, ls, 1, 2)

	dd := requireType(t, doc, TypeDefinitionDestination)
	assertText(t, dd, "/url")
	requireType(t, doc, TypeDefinitionDestinationString)
}

func TestDefinitionWithTitle(t *testing.T) {
	doc := Parse(`[ref]: /url "Title"` + "\n")
	requireType(t, doc, TypeDefinition)
	dd := requireType(t, doc, TypeDefinitionDestination)
	assertText(t, dd, "/url")
}

// --- GFM tables ---

func TestTable(t *testing.T) {
	doc := Parse("| a | b |\n| - | - |\n| 1 | 2 |\n")
	requireType(t, doc, TypeTable)
	requireType(t, doc, TypeTableHeader)
	requireType(t, doc, TypeTableDelimiterRow)

	rows := findByType(doc, TypeTableRow)
	if len(rows) != 1 {
		t.Fatalf("want 1 body tableRow, got %d", len(rows))
	}
	assertSpanStart(t, rows[0], 3, 1)

	// Header has 2 data cells.
	datas := findByType(doc, TypeTableData)
	// header(2) + body(2) = 4
	if len(datas) != 4 {
		t.Errorf("want 4 tableData cells, got %d", len(datas))
	}
	assertText(t, datas[0], " a ")

	// Delimiter row uses tableDelimiter cells, not tableData.
	delims := findByType(doc, TypeTableDelimiter)
	if len(delims) != 2 {
		t.Errorf("want 2 tableDelimiter cells, got %d", len(delims))
	}

	// Cell dividers (pipes).
	if countByType(doc, TypeTableCellDivider) != 9 {
		t.Errorf("want 9 cell dividers across 3 rows, got %d", countByType(doc, TypeTableCellDivider))
	}
}

func TestTableNoLeadingPipe(t *testing.T) {
	// NOTE: a delimiter row whose first cell starts with "- " ("- | -") is parsed
	// as a list item before table detection runs, so no table is produced. The
	// table detector requires the second line to match the delimiter regex while
	// not being mistaken for a list; use leading pipes to avoid this.
	doc := Parse("a | b\n- | -\n")
	if countByType(doc, TypeTable) != 0 {
		t.Errorf("'- | -' delimiter row is consumed as a list item, not a table")
	}
	requireType(t, doc, TypeListUnordered)
}

func TestTablePipeDelimited(t *testing.T) {
	// With explicit outer pipes the delimiter row is recognized and a table forms.
	doc := Parse("| a | b |\n|---|---|\n")
	requireType(t, doc, TypeTable)
	requireType(t, doc, TypeTableHeader)
	requireType(t, doc, TypeTableDelimiterRow)
}

// --- Inline: code spans ---

func TestCodeSpan(t *testing.T) {
	doc := Parse("`code`\n")
	ct := requireType(t, doc, TypeCodeText)
	assertSpanStart(t, ct, 1, 1)
	assertText(t, ct, "`code`")

	seqs := findByType(doc, TypeCodeTextSequence)
	if len(seqs) != 2 {
		t.Fatalf("want 2 codeTextSequence, got %d", len(seqs))
	}
	assertText(t, seqs[0], "`")
	assertSpanStart(t, seqs[0], 1, 1)
	assertSpanStart(t, seqs[1], 1, 6)

	cd := requireType(t, doc, TypeCodeTextData)
	assertText(t, cd, "code")
	assertSpanStart(t, cd, 1, 2)
}

func TestCodeSpanMultiBacktick(t *testing.T) {
	doc := Parse("``a`b``\n")
	ct := requireType(t, doc, TypeCodeText)
	assertText(t, ct, "``a`b``")
	cd := requireType(t, doc, TypeCodeTextData)
	assertText(t, cd, "a`b")
}

func TestCodeSpanUnclosed(t *testing.T) {
	// No closing run: backtick stays plain data, no codeText.
	doc := Parse("`unclosed\n")
	if countByType(doc, TypeCodeText) != 0 {
		t.Errorf("unclosed code span should not produce codeText")
	}
}

// --- Inline: emphasis / strong ---

func TestEmphasis(t *testing.T) {
	doc := Parse("*foo*\n")
	em := requireType(t, doc, TypeEmphasis)
	assertSpanStart(t, em, 1, 1)
	seqs := findByType(doc, TypeEmphasisSequence)
	if len(seqs) != 2 {
		t.Fatalf("want 2 emphasisSequence, got %d", len(seqs))
	}
	assertText(t, seqs[0], "*")
	requireType(t, doc, TypeEmphasisText)
}

func TestEmphasisUnderscore(t *testing.T) {
	doc := Parse("_foo_\n")
	requireType(t, doc, TypeEmphasis)
	seq := firstByType(doc, TypeEmphasisSequence)
	assertText(t, seq, "_")
}

func TestStrong(t *testing.T) {
	doc := Parse("**strong**\n")
	requireType(t, doc, TypeStrong)
	seqs := findByType(doc, TypeStrongSequence)
	if len(seqs) != 2 {
		t.Fatalf("want 2 strongSequence, got %d", len(seqs))
	}
	assertText(t, seqs[0], "**")
	requireType(t, doc, TypeStrongText)
}

func TestStrongEmphasisCombined(t *testing.T) {
	// NOTE: For "***x***" this parser resolves the outer run as emphasis wrapping
	// an inner strong (emphasis > strong), not strong > emphasis. We characterize
	// the actual nesting.
	doc := Parse("***bold em***\n")
	em := requireType(t, doc, TypeEmphasis)
	strong := requireType(t, doc, TypeStrong)
	if strong.Parent == nil || strong.Parent.Type != TypeEmphasisText {
		t.Errorf("expected strong nested inside emphasisText, got parent %v", strong.Parent)
	}
	_ = em
}

func TestEmphasisBareMarkersWithSpaces(t *testing.T) {
	// NOTE: "* foo *" does not form emphasis (surrounding spaces defeat flanking).
	// The first "*" actually triggers a list at the block level; the inner content
	// "foo *" remains plain data tokens (a bare "*" marker), which MD037 inspects.
	doc := Parse("a * foo * b\n")
	if countByType(doc, TypeEmphasis) != 0 {
		t.Errorf("'* foo *' should not form emphasis")
	}
	// The bare '*' markers survive as data tokens.
	var foundStar bool
	for _, d := range findByType(doc, TypeData) {
		if d.Text == "*" {
			foundStar = true
		}
	}
	if !foundStar {
		t.Errorf("expected a bare '*' data marker to survive")
	}
}

func TestEmphasisIntrawordUnderscore(t *testing.T) {
	// NOTE: intraword underscores do not create emphasis; markers stay as data.
	doc := Parse("foo_bar_baz\n")
	if countByType(doc, TypeEmphasis) != 0 {
		t.Errorf("intraword underscore should not form emphasis")
	}
}

// --- Inline: links / images ---

func TestInlineLink(t *testing.T) {
	doc := Parse("[txt](/url)\n")
	link := requireType(t, doc, TypeLink)
	assertSpanStart(t, link, 1, 1)
	lt := requireType(t, doc, TypeLabelText)
	assertText(t, lt, "txt")
	requireType(t, doc, TypeResource)

	rd := requireType(t, doc, TypeResourceDestination)
	assertText(t, rd, "/url")
	requireType(t, doc, TypeResourceDestinationRaw)
	requireType(t, doc, TypeResourceDestinationString)
}

func TestInlineLinkAngleDest(t *testing.T) {
	// An angle-bracketed destination (no spaces) yields a resourceDestinationLiteral
	// wrapping a resourceDestinationString of the inner text.
	doc := Parse("[txt](<http://x.com>)\n")
	requireType(t, doc, TypeLink)
	lit := requireType(t, doc, TypeResourceDestinationLiteral)
	assertText(t, lit, "<http://x.com>")
	rs := requireType(t, doc, TypeResourceDestinationString)
	assertText(t, rs, "http://x.com")
}

func TestInlineLinkAngleDestWithSpace(t *testing.T) {
	// NOTE: this parser splits the destination at the first whitespace, so an
	// angle-bracketed destination containing a space is truncated to "</my" and
	// classified as raw (not literal) — a known divergence.
	doc := Parse("[txt](</my url>)\n")
	requireType(t, doc, TypeLink)
	if countByType(doc, TypeResourceDestinationLiteral) != 0 {
		t.Errorf("space-containing angle dest is not treated as literal")
	}
	rd := requireType(t, doc, TypeResourceDestination)
	assertText(t, rd, "</my")
	requireType(t, doc, TypeResourceDestinationRaw)
}

func TestInlineImage(t *testing.T) {
	doc := Parse("![alt](/img.png)\n")
	img := requireType(t, doc, TypeImage)
	assertSpanStart(t, img, 1, 1)
	lt := requireType(t, doc, TypeLabelText)
	assertText(t, lt, "alt")
	rd := requireType(t, doc, TypeResourceDestination)
	assertText(t, rd, "/img.png")
}

func TestReferenceLinkDefined(t *testing.T) {
	doc := Parse("[txt][ref]\n\n[ref]: /url\n")
	link := requireType(t, doc, TypeLink)
	assertSpanStart(t, link, 1, 1)
	ref := requireType(t, doc, TypeReference)
	assertSpanStart(t, ref, 1, 6)
	rs := requireType(t, doc, TypeReferenceString)
	assertText(t, rs, "ref")
}

func TestCollapsedReferenceDefined(t *testing.T) {
	doc := Parse("[txt][]\n\n[txt]: /url\n")
	requireType(t, doc, TypeLink)
	ref := requireType(t, doc, TypeReference)
	assertText(t, ref, "[]")
}

func TestShortcutReferenceDefined(t *testing.T) {
	doc := Parse("[txt]\n\n[txt]: /url\n")
	requireType(t, doc, TypeLink)
	lt := requireType(t, doc, TypeLabelText)
	assertText(t, lt, "txt")
	// shortcut form has no reference token.
	if countByType(doc, TypeReference) != 0 {
		t.Errorf("shortcut reference should not emit a reference token")
	}
}

func TestUndefinedReferenceShortcut(t *testing.T) {
	doc := Parse("[undef]\n")
	ur := requireType(t, doc, TypeUndefinedReferenceShortcut)
	assertSpanStart(t, ur, 1, 1)
	assertText(t, ur, "[undef]")
	inner := requireType(t, doc, TypeUndefinedReference)
	assertText(t, inner, "[undef]")
	// The concatenated child data equals the label.
	d := firstByType(doc, TypeData)
	assertText(t, d, "undef")
	if countByType(doc, TypeLink) != 0 {
		t.Errorf("undefined shortcut should not be a real link")
	}
}

func TestUndefinedReferenceFull(t *testing.T) {
	doc := Parse("[txt][undef]\n")
	ur := requireType(t, doc, TypeUndefinedReferenceFull)
	assertText(t, ur, "[txt][undef]")
	requireType(t, doc, TypeUndefinedReference)
	if countByType(doc, TypeLink) != 0 {
		t.Errorf("undefined full reference should not be a real link")
	}
}

func TestUndefinedReferenceCollapsed(t *testing.T) {
	doc := Parse("[undef][]\n")
	requireType(t, doc, TypeUndefinedReferenceCollapsed)
}

func TestBracketsNotReference(t *testing.T) {
	// A label containing ']' (nested) does not become an undefined reference;
	// the outer brackets stay data while the inner bracket pair forms a shortcut.
	doc := Parse("[a [nested] b]\n")
	if countByType(doc, TypeUndefinedReferenceShortcut) != 1 {
		t.Errorf("expected exactly 1 inner undefined shortcut, got %d", countByType(doc, TypeUndefinedReferenceShortcut))
	}
	ur := firstByType(doc, TypeUndefinedReferenceShortcut)
	assertText(t, ur, "[nested]")
}

// --- Inline: autolinks ---

func TestAutolinkURI(t *testing.T) {
	// Autolinks resolve inside inline text (a line starting with '<' is parsed as
	// an HTML block, see TestAutolinkAsHTMLBlock).
	doc := Parse("see <http://example.com> here\n")
	a := requireType(t, doc, TypeAutolink)
	assertText(t, a, "<http://example.com>")
	prot := requireType(t, doc, TypeAutolinkProtocol)
	assertText(t, prot, "http://example.com")
}

func TestAutolinkEmail(t *testing.T) {
	doc := Parse("mail <me@example.com> ok\n")
	requireType(t, doc, TypeAutolink)
	em := requireType(t, doc, TypeAutolinkEmail)
	assertText(t, em, "me@example.com")
}

func TestAutolinkAsHTMLBlock(t *testing.T) {
	// NOTE: a line beginning with "<http://...>" is parsed as an HTML block at the
	// block level, so it becomes htmlFlow, not an autolink (a divergence).
	doc := Parse("<http://example.com>\n")
	requireType(t, doc, TypeHTMLFlow)
	if countByType(doc, TypeAutolink) != 0 {
		t.Errorf("leading-'<' line is an HTML block, not an autolink")
	}
}

// --- Inline: GFM literal autolinks ---

func TestLiteralAutolinkHTTP(t *testing.T) {
	doc := Parse("see http://example.com now\n")
	la := requireType(t, doc, TypeLiteralAutolink)
	assertText(t, la, "http://example.com")
	// NOTE: this parser emits the generic literalAutolink type and never the
	// typed subtypes literalAutolinkHttp / literalAutolinkWww / literalAutolinkEmail.
	if countByType(doc, TypeLiteralAutolinkHTTP) != 0 {
		t.Errorf("parser does not emit typed literalAutolinkHttp")
	}
}

func TestLiteralAutolinkWWW(t *testing.T) {
	doc := Parse("visit www.example.com today\n")
	la := requireType(t, doc, TypeLiteralAutolink)
	assertText(t, la, "www.example.com")
}

func TestLiteralAutolinkEmail(t *testing.T) {
	doc := Parse("ping a@b.com please\n")
	la := requireType(t, doc, TypeLiteralAutolink)
	assertText(t, la, "a@b.com")
}

func TestLiteralAutolinkMultiple(t *testing.T) {
	doc := Parse("http://x.com and www.y.com and a@b.com\n")
	las := findByType(doc, TypeLiteralAutolink)
	if len(las) != 3 {
		t.Fatalf("want 3 literal autolinks, got %d", len(las))
	}
	assertText(t, las[0], "http://x.com")
	assertText(t, las[1], "www.y.com")
	assertText(t, las[2], "a@b.com")
}

// --- Inline: character escapes ---

func TestCharacterEscape(t *testing.T) {
	doc := Parse("a\\*b\n")
	esc := requireType(t, doc, TypeCharacterEscape)
	assertText(t, esc, "\\*")
	assertSpanStart(t, esc, 1, 2)
	// The escaped '*' must not form emphasis.
	if countByType(doc, TypeEmphasis) != 0 {
		t.Errorf("escaped '*' should not form emphasis")
	}
}

func TestCharacterEscapeNonPunct(t *testing.T) {
	// Backslash before a non-punctuation char is not an escape; stays data.
	doc := Parse("a\\b\n")
	if countByType(doc, TypeCharacterEscape) != 0 {
		t.Errorf("backslash before letter is not a character escape")
	}
}

// --- Inline: HTML ---

func TestInlineHTML(t *testing.T) {
	doc := Parse("x <span> y\n")
	h := requireType(t, doc, TypeHTMLText)
	assertText(t, h, "<span>")
	assertSpanStart(t, h, 1, 3)
	requireType(t, doc, TypeHTMLTextData)
}

func TestInlineHTMLMultiple(t *testing.T) {
	doc := Parse("x <b><i> y\n")
	if countByType(doc, TypeHTMLText) != 2 {
		t.Errorf("want 2 htmlText tokens, got %d", countByType(doc, TypeHTMLText))
	}
}

func TestHTMLBlock(t *testing.T) {
	doc := Parse("<div>\ncontent\n</div>\n")
	hf := requireType(t, doc, TypeHTMLFlow)
	assertSpanStart(t, hf, 1, 1)
	requireType(t, doc, TypeHTMLFlowData)
}

func TestHTMLBlockComment(t *testing.T) {
	doc := Parse("<!-- a comment -->\n")
	hf := requireType(t, doc, TypeHTMLFlow)
	if !strings.Contains(hf.Text, "comment") {
		t.Errorf("comment HTML block missing content: %q", hf.Text)
	}
}

// --- Structural: edge cases, ordering, parent links ---

func TestEmptyInput(t *testing.T) {
	doc := Parse("")
	if doc == nil {
		t.Fatal("Parse(\"\") returned nil")
	}
	if len(doc.Children) != 0 {
		t.Errorf("empty input should have no children, got %d", len(doc.Children))
	}
	if len(doc.Flat) != 0 {
		t.Errorf("empty input should have empty Flat, got %d", len(doc.Flat))
	}
}

func TestBlankLinesOnly(t *testing.T) {
	doc := Parse("\n\n\n")
	if len(doc.Children) != 0 {
		t.Errorf("blank-only input should produce no top-level tokens, got %d", len(doc.Children))
	}
}

func TestSingleNewline(t *testing.T) {
	doc := Parse("\n")
	if len(doc.Flat) != 0 {
		t.Errorf("single newline should produce no tokens, got %d", len(doc.Flat))
	}
}

func TestParentLinks(t *testing.T) {
	doc := Parse("# Heading\n")
	// Top-level children have nil parent.
	for _, c := range doc.Children {
		if c.Parent != nil {
			t.Errorf("top-level token %q should have nil Parent", c.Type)
		}
	}
	// atxHeadingText's parent is the atxHeading.
	htxt := requireType(t, doc, TypeAtxHeadingText)
	if htxt.Parent == nil || htxt.Parent.Type != TypeAtxHeading {
		t.Errorf("atxHeadingText parent should be atxHeading, got %v", htxt.Parent)
	}
	// the data inside heading text has the heading text as parent.
	d := firstByType(doc, TypeData)
	if d.Parent == nil || d.Parent.Type != TypeAtxHeadingText {
		t.Errorf("heading data parent should be atxHeadingText, got %v", d.Parent)
	}
}

func TestFlatOrdering(t *testing.T) {
	// Flat is a depth-first walk: a parent appears before its children, and
	// children appear before the parent's later siblings.
	doc := Parse("# H\n\npara\n")
	types := make([]TokenType, len(doc.Flat))
	for i, tk := range doc.Flat {
		types[i] = tk.Type
	}
	// Expected prefix: atxHeading, atxHeadingSequence, whitespace, atxHeadingText, data, then paragraph, data.
	wantPrefix := []TokenType{
		TypeAtxHeading, TypeAtxHeadingSequence, TypeWhitespace,
		TypeAtxHeadingText, TypeData, TypeParagraph, TypeData,
	}
	if len(types) != len(wantPrefix) {
		t.Fatalf("flat length: got %d (%v), want %d", len(types), types, len(wantPrefix))
	}
	for i := range wantPrefix {
		if types[i] != wantPrefix[i] {
			t.Errorf("flat[%d]: got %q, want %q", i, types[i], wantPrefix[i])
		}
	}

	// Verify every child appears after its parent in flat order.
	pos := map[*Token]int{}
	for i, tk := range doc.Flat {
		pos[tk] = i
	}
	for _, tk := range doc.Flat {
		if tk.Parent != nil {
			if pos[tk.Parent] >= pos[tk] {
				t.Errorf("token %q appears before its parent %q in Flat", tk.Type, tk.Parent.Type)
			}
		}
	}
}

func TestInHTMLFlowFlag(t *testing.T) {
	// Default tokens are not in HTML flow.
	doc := Parse("# H\n")
	for _, tk := range doc.Flat {
		if tk.InHTMLFlow() {
			t.Errorf("token %q unexpectedly marked InHTMLFlow", tk.Type)
		}
	}
}

func TestEndColumnsExclusive(t *testing.T) {
	// Verify end columns are exclusive: EndColumn-StartColumn == rune length for a
	// single-line token with known text.
	doc := Parse("para\n")
	d := requireType(t, doc, TypeData)
	if got := d.EndColumn - d.StartColumn; got != len([]rune(d.Text)) {
		t.Errorf("exclusive end col: EndColumn-StartColumn=%d, want rune len %d", got, len([]rune(d.Text)))
	}
}

func TestNoNewlineAtEOF(t *testing.T) {
	// Document without a trailing newline parses normally.
	doc := Parse("# Heading")
	h := requireType(t, doc, TypeAtxHeading)
	assertText(t, h, "# Heading")
}

func TestMixedDocument(t *testing.T) {
	md := "# Title\n\nA paragraph with `code` and *emphasis*.\n\n- list item\n\n> a quote\n\n```\nfenced\n```\n"
	doc := Parse(md)
	requireType(t, doc, TypeAtxHeading)
	requireType(t, doc, TypeCodeText)
	requireType(t, doc, TypeEmphasis)
	requireType(t, doc, TypeListUnordered)
	requireType(t, doc, TypeBlockQuote)
	requireType(t, doc, TypeCodeFenced)
	// Sanity: all parent links are consistent with Flat ordering.
	pos := map[*Token]int{}
	for i, tk := range doc.Flat {
		pos[tk] = i
	}
	for _, tk := range doc.Flat {
		if tk.Parent != nil && pos[tk.Parent] >= pos[tk] {
			t.Errorf("token %q precedes parent %q", tk.Type, tk.Parent.Type)
		}
	}
}

func TestCRLFLineEndings(t *testing.T) {
	doc := Parse("# Heading\r\npara\r\n")
	requireType(t, doc, TypeAtxHeading)
	requireType(t, doc, TypeParagraph)
}

func TestBlockQuoteLazyContinuation(t *testing.T) {
	// A non-blank line following a "> " line lazily continues the blockquote.
	doc := Parse("> quoted\nlazy line\n")
	requireType(t, doc, TypeBlockQuote)
	// Both lines fold into a single paragraph inside the quote.
	p := requireType(t, doc, TypeParagraph)
	if p.EndLine != 2 {
		t.Errorf("lazy continuation paragraph EndLine: got %d, want 2", p.EndLine)
	}
}

func TestIndentedCodeWithBlankLine(t *testing.T) {
	// A blank line between two indented lines is kept inside the code block.
	doc := Parse("    a\n\n    b\n")
	if countByType(doc, TypeCodeIndented) != 1 {
		t.Errorf("blank line inside indented code should not split it, got %d blocks", countByType(doc, TypeCodeIndented))
	}
	vals := findByType(doc, TypeCodeFlowValue)
	if len(vals) != 2 {
		t.Errorf("want 2 codeFlowValue across the blank, got %d", len(vals))
	}
}

func TestIndentedCodeTrailingBlank(t *testing.T) {
	// A trailing blank line that is not followed by more indented content ends
	// the indented code block.
	doc := Parse("    code\n\nplain\n")
	requireType(t, doc, TypeCodeIndented)
	requireType(t, doc, TypeParagraph)
}

func TestParagraphInterruptedByList(t *testing.T) {
	// A bullet item with content interrupts a paragraph.
	doc := Parse("text\n- item\n")
	// The original line becomes a standalone paragraph; the list item content
	// is a second paragraph inside the list.
	paras := findByType(doc, TypeParagraph)
	if len(paras) != 2 {
		t.Fatalf("want 2 paragraphs (text + list content), got %d", len(paras))
	}
	if paras[0].EndLine != 1 {
		t.Errorf("first paragraph should end on line 1, got %d", paras[0].EndLine)
	}
	requireType(t, doc, TypeListUnordered)
}

func TestParagraphNotInterruptedByOrderedNonOne(t *testing.T) {
	// NOTE: an ordered list starting at a value other than 1 does NOT interrupt a
	// paragraph (CommonMark behavior), so "2. x" folds into the paragraph text.
	doc := Parse("text\n2. not a list\n")
	if countByType(doc, TypeListOrdered) != 0 {
		t.Errorf("ordered list not starting at 1 should not interrupt the paragraph")
	}
	p := requireType(t, doc, TypeParagraph)
	if p.EndLine != 2 {
		t.Errorf("paragraph should absorb the '2.' line, EndLine got %d", p.EndLine)
	}
}

func TestParagraphInterruptedByHeading(t *testing.T) {
	doc := Parse("text\n# heading\n")
	requireType(t, doc, TypeParagraph)
	requireType(t, doc, TypeAtxHeading)
}

func TestLinkLabelEscapedBracket(t *testing.T) {
	// An escaped closing bracket inside the label is not treated as the label end.
	doc := Parse("[a\\]b](/url)\n")
	link := requireType(t, doc, TypeLink)
	assertSpanStart(t, link, 1, 1)
	lt := requireType(t, doc, TypeLabelText)
	assertText(t, lt, "a\\]b")
}
