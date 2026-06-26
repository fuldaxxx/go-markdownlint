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

package mdhelpers

import (
	"reflect"
	"testing"

	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
)

// parseFlat parses markdown and returns the flat token list.
func parseFlat(t *testing.T, md string) []*mm.Token {
	t.Helper()
	return mm.Parse(md).Flat
}

// parseTop parses markdown and returns the top-level tokens.
func parseTop(t *testing.T, md string) []*mm.Token {
	t.Helper()
	return mm.Parse(md).Children
}

// firstOfType returns the first token of the given type from a flat list.
func firstOfType(flat []*mm.Token, typ mm.TokenType) *mm.Token {
	for _, tok := range flat {
		if tok.Type == typ {
			return tok
		}
	}
	return nil
}

func TestAddRangeToSet(t *testing.T) {
	tests := []struct {
		name       string
		start, end int
		want       []int
	}{
		{"single", 3, 3, []int{3}},
		{"range", 1, 4, []int{1, 2, 3, 4}},
		{"empty_when_end_lt_start", 5, 2, nil},
		{"negative", -2, 0, []int{-2, -1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := map[int]struct{}{}
			AddRangeToSet(set, tt.start, tt.end)
			if len(set) != len(tt.want) {
				t.Fatalf("len = %d, want %d (%v)", len(set), len(tt.want), tt.want)
			}
			for _, v := range tt.want {
				if _, ok := set[v]; !ok {
					t.Errorf("missing %d in %v", v, set)
				}
			}
		})
	}
}

func TestAddRangeToSetMergesIntoExisting(t *testing.T) {
	set := map[int]struct{}{10: {}}
	AddRangeToSet(set, 1, 2)
	for _, v := range []int{1, 2, 10} {
		if _, ok := set[v]; !ok {
			t.Errorf("missing %d", v)
		}
	}
}

func TestFilterByPredicate(t *testing.T) {
	flat := parseTop(t, "# Heading 1\n")
	// allowed selects atxHeadingSequence anywhere in the tree.
	got := FilterByPredicate(flat, func(tok *mm.Token) bool {
		return tok.Type == mm.TypeAtxHeadingSequence
	}, nil)
	if len(got) != 1 {
		t.Fatalf("want 1 seq, got %d", len(got))
	}
	if got[0].Text != "#" {
		t.Errorf("seq text = %q", got[0].Text)
	}
}

func TestFilterByPredicateNoneAllowed(t *testing.T) {
	flat := parseTop(t, "# Heading\n")
	got := FilterByPredicate(flat, func(*mm.Token) bool { return false }, nil)
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

func TestFilterByPredicateAllAllowed(t *testing.T) {
	top := parseTop(t, "# H\n")
	flat := parseFlat(t, "# H\n")
	got := FilterByPredicate(top, func(*mm.Token) bool { return true }, nil)
	// Walking the tree visits every node; should match the flat count.
	if len(got) != len(flat) {
		t.Errorf("walk count = %d, flat count = %d", len(got), len(flat))
	}
}

func TestFilterByPredicateTransform(t *testing.T) {
	top := parseTop(t, "# Heading\n")
	heading := top[0]
	// transform replaces children of the heading with a single synthetic token.
	synthetic := &mm.Token{Type: "synthetic"}
	got := FilterByPredicate(top, func(tok *mm.Token) bool {
		return tok.Type == "synthetic" || tok.Type == mm.TypeAtxHeading
	}, func(tok *mm.Token) []*mm.Token {
		if tok == heading {
			return []*mm.Token{synthetic}
		}
		return tok.Children
	})
	// Expect the heading plus the synthetic child substituted in.
	foundSynthetic := false
	for _, g := range got {
		if g.Type == "synthetic" {
			foundSynthetic = true
		}
	}
	if !foundSynthetic {
		t.Errorf("transform child not visited: %v", got)
	}
}

func TestFilterByTypes(t *testing.T) {
	top := parseTop(t, "# One\n\n## Two\n")
	got := FilterByTypes(top, []mm.TokenType{mm.TypeAtxHeading}, false)
	if len(got) != 2 {
		t.Fatalf("want 2 headings, got %d", len(got))
	}
	got = FilterByTypes(top, []mm.TokenType{mm.TypeAtxHeadingText}, false)
	if len(got) != 2 {
		t.Fatalf("want 2 heading texts, got %d", len(got))
	}
}

func TestFilterByTypesMultipleTypes(t *testing.T) {
	top := parseTop(t, "# H\n\ntext\n")
	got := FilterByTypes(top, []mm.TokenType{mm.TypeAtxHeading, mm.TypeParagraph}, false)
	if len(got) != 2 {
		t.Fatalf("want 2 (heading+paragraph), got %d", len(got))
	}
}

func TestFilterByTypesEmptyInput(t *testing.T) {
	got := FilterByTypes(nil, []mm.TokenType{mm.TypeParagraph}, false)
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

func TestFilterByTypesHTMLFlowFlag(t *testing.T) {
	// htmlFlow=false excludes tokens inside HTML-flow reparse. A plain comment
	// produces only htmlFlow/htmlFlowData; none are marked InHTMLFlow here, so
	// confirm filtering still returns the htmlFlow token itself.
	top := parseTop(t, "<!-- comment -->\n")
	got := FilterByTypes(top, []mm.TokenType{mm.TypeHTMLFlow}, false)
	if len(got) != 1 {
		t.Fatalf("want 1 htmlFlow, got %d", len(got))
	}
	gotTrue := FilterByTypes(top, []mm.TokenType{mm.TypeHTMLFlow}, true)
	if len(gotTrue) != 1 {
		t.Fatalf("htmlFlow=true: want 1, got %d", len(gotTrue))
	}
}

func TestFilterFlat(t *testing.T) {
	flat := parseFlat(t, "# One\n\nbody\n")
	got := FilterFlat(flat, []mm.TokenType{mm.TypeAtxHeading}, false)
	if len(got) != 1 {
		t.Fatalf("want 1 heading, got %d", len(got))
	}
	none := FilterFlat(flat, []mm.TokenType{mm.TypeTable}, false)
	if len(none) != 0 {
		t.Fatalf("want 0 tables, got %d", len(none))
	}
}

func TestFilterFlatMatchesFilterByTypes(t *testing.T) {
	md := "# One\n\n## Two\n\nbody **bold**\n"
	flat := parseFlat(t, md)
	top := parseTop(t, md)
	types := []mm.TokenType{mm.TypeAtxHeading, mm.TypeParagraph}
	a := FilterFlat(flat, types, false)
	b := FilterByTypes(top, types, false)
	if len(a) != len(b) {
		t.Fatalf("FilterFlat=%d FilterByTypes=%d", len(a), len(b))
	}
}

func TestGetDescendantsByType(t *testing.T) {
	top := parseTop(t, "# Hello\n")
	heading := firstOfType(parseFlat(t, "# Hello\n"), mm.TypeAtxHeading)
	_ = top
	got := GetDescendantsByType([]*mm.Token{heading}, [][]mm.TokenType{
		{mm.TypeAtxHeadingText},
	})
	if len(got) != 1 || got[0].Text != "Hello" {
		t.Fatalf("got %#v", got)
	}
}

func TestGetDescendantsByTypeNestedPath(t *testing.T) {
	flat := parseFlat(t, "# Hello\n")
	heading := firstOfType(flat, mm.TypeAtxHeading)
	// atxHeadingText -> data
	got := GetDescendantsByType([]*mm.Token{heading}, [][]mm.TokenType{
		{mm.TypeAtxHeadingText},
		{mm.TypeData},
	})
	if len(got) != 1 || got[0].Text != "Hello" {
		t.Fatalf("got %#v", got)
	}
}

func TestGetDescendantsByTypeNoMatch(t *testing.T) {
	flat := parseFlat(t, "# Hello\n")
	heading := firstOfType(flat, mm.TypeAtxHeading)
	got := GetDescendantsByType([]*mm.Token{heading}, [][]mm.TokenType{
		{mm.TypeTable},
	})
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

func TestGetDescendantsByTypeMultipleAlternatives(t *testing.T) {
	flat := parseFlat(t, "Setext\n======\n")
	heading := firstOfType(flat, mm.TypeSetextHeading)
	got := GetDescendantsByType([]*mm.Token{heading}, [][]mm.TokenType{
		{mm.TypeAtxHeadingText, mm.TypeSetextHeadingText},
	})
	if len(got) != 1 || got[0].Type != mm.TypeSetextHeadingText {
		t.Fatalf("got %#v", got)
	}
}

func TestGetParentOfType(t *testing.T) {
	flat := parseFlat(t, "# Hello\n")
	data := firstOfType(flat, mm.TypeData)
	// data is inside atxHeadingText inside atxHeading.
	parent := GetParentOfType(data, []mm.TokenType{mm.TypeAtxHeading})
	if parent == nil || parent.Type != mm.TypeAtxHeading {
		t.Fatalf("got %#v", parent)
	}
	// Nearest ancestor of two candidate types.
	near := GetParentOfType(data, []mm.TokenType{mm.TypeAtxHeading, mm.TypeAtxHeadingText})
	if near == nil || near.Type != mm.TypeAtxHeadingText {
		t.Fatalf("want nearest atxHeadingText, got %#v", near)
	}
}

func TestGetParentOfTypeNotFound(t *testing.T) {
	flat := parseFlat(t, "# Hello\n")
	data := firstOfType(flat, mm.TypeData)
	parent := GetParentOfType(data, []mm.TokenType{mm.TypeTable})
	if parent != nil {
		t.Fatalf("want nil, got %#v", parent)
	}
}

func TestGetHeadingLevel(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want int
	}{
		{"h1", "# One\n", 1},
		{"h2", "## Two\n", 2},
		{"h6", "###### Six\n", 6},
		{"setext_h1", "Title\n=====\n", 1},
		// NOTE: setext h2 ("Title\n-----\n") is a known parser divergence; this
		// implementation tokenizes the "-----" line as a thematicBreak rather
		// than a setextHeading, so it is covered separately below by
		// hand-constructing a setextHeadingLine starting with '-'.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flat := parseFlat(t, tt.md)
			h := firstOfType(flat, mm.TypeAtxHeading)
			if h == nil {
				h = firstOfType(flat, mm.TypeSetextHeading)
			}
			if got := GetHeadingLevel(h); got != tt.want {
				t.Errorf("level = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetHeadingLevelSetextDash(t *testing.T) {
	// Hand-constructed setext heading with a '-' underline -> level 2.
	h := &mm.Token{Type: mm.TypeSetextHeading, Children: []*mm.Token{
		{Type: mm.TypeSetextHeadingText, Children: []*mm.Token{{Type: mm.TypeData, Text: "Title"}}},
		{Type: mm.TypeSetextHeadingLine, Text: "-----"},
	}}
	if got := GetHeadingLevel(h); got != 2 {
		t.Errorf("want 2, got %d", got)
	}
}

func TestGetHeadingLevelNoSequence(t *testing.T) {
	// Hand-constructed heading with no sequence child defaults to level 1.
	h := &mm.Token{Type: mm.TypeAtxHeading}
	if got := GetHeadingLevel(h); got != 1 {
		t.Errorf("want 1, got %d", got)
	}
}

func TestGetHeadingLevelClampedAboveSix(t *testing.T) {
	// Hand-constructed sequence longer than 6 hashes clamps to 6.
	h := &mm.Token{Type: mm.TypeAtxHeading, Children: []*mm.Token{
		{Type: mm.TypeAtxHeadingSequence, Text: "#######"},
	}}
	if got := GetHeadingLevel(h); got != 6 {
		t.Errorf("want 6, got %d", got)
	}
}

func TestGetHeadingStyle(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want string
	}{
		{"atx", "# Open\n", "atx"},
		{"atx_closed", "# Closed #\n", "atx_closed"},
		{"setext", "Title\n=====\n", "setext"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flat := parseFlat(t, tt.md)
			h := firstOfType(flat, mm.TypeAtxHeading)
			if h == nil {
				h = firstOfType(flat, mm.TypeSetextHeading)
			}
			if got := GetHeadingStyle(h); got != tt.want {
				t.Errorf("style = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetHeadingText(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want string
	}{
		{"atx", "# Hello world\n", "Hello world"},
		{"setext", "Title here\n==========\n", "Title here"},
		{"with_html_excluded", "# Some <b>bold</b> text\n", "Some bold text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flat := parseFlat(t, tt.md)
			h := firstOfType(flat, mm.TypeAtxHeading)
			if h == nil {
				h = firstOfType(flat, mm.TypeSetextHeading)
			}
			if got := GetHeadingText(h); got != tt.want {
				t.Errorf("text = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetHeadingTextNewlinesCollapsed(t *testing.T) {
	// Hand-construct a heading text whose data contains a newline.
	h := &mm.Token{Type: mm.TypeAtxHeading, Children: []*mm.Token{
		{Type: mm.TypeAtxHeadingText, Children: []*mm.Token{
			{Type: mm.TypeData, Text: "a\nb"},
		}},
	}}
	if got := GetHeadingText(h); got != "a b" {
		t.Errorf("want %q, got %q", "a b", got)
	}
}

func TestGetHTMLTagInfo(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantNil   bool
		wantClose bool
		wantName  string
	}{
		{"open", "<b>", false, false, "b"},
		{"close", "</b>", false, true, "b"},
		{"open_with_attrs", `<a href="x">`, false, false, "a"},
		{"comment_not_a_tag", "<!-- c -->", true, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &mm.Token{Type: mm.TypeHTMLText, Text: tt.text}
			got := GetHTMLTagInfo(tok)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("want nil, got %#v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("want non-nil")
			}
			if got.Close != tt.wantClose || got.Name != tt.wantName {
				t.Errorf("got %#v, want {Close:%v Name:%q}", got, tt.wantClose, tt.wantName)
			}
		})
	}
}

func TestGetHTMLTagInfoWrongType(t *testing.T) {
	tok := &mm.Token{Type: mm.TypeData, Text: "<b>"}
	if got := GetHTMLTagInfo(tok); got != nil {
		t.Fatalf("want nil for non-htmlText, got %#v", got)
	}
}

func TestGetHTMLTagInfoFromParse(t *testing.T) {
	flat := parseFlat(t, "Some <b>bold</b> text\n")
	var infos []*HTMLTagInfo
	for _, tok := range flat {
		if tok.Type == mm.TypeHTMLText {
			if info := GetHTMLTagInfo(tok); info != nil {
				infos = append(infos, info)
			}
		}
	}
	if len(infos) != 2 {
		t.Fatalf("want 2 tags, got %d", len(infos))
	}
	if infos[0].Close || infos[0].Name != "b" {
		t.Errorf("open tag = %#v", infos[0])
	}
	if !infos[1].Close || infos[1].Name != "b" {
		t.Errorf("close tag = %#v", infos[1])
	}
}

func TestGetBlockQuotePrefixText(t *testing.T) {
	top := parseTop(t, "> a\n> b\n")
	// Line 1 prefix is "> " -> trimmed "&gt;" then "\n", repeated count times.
	got := GetBlockQuotePrefixText(top, 1, 1)
	if got != ">\n" {
		t.Errorf("got %q, want %q", got, ">\n")
	}
	got2 := GetBlockQuotePrefixText(top, 1, 2)
	if got2 != ">\n>\n" {
		t.Errorf("got %q, want %q", got2, ">\n>\n")
	}
}

func TestGetBlockQuotePrefixTextNoPrefixForLine(t *testing.T) {
	top := parseTop(t, "> a\n")
	// A line with no matching prefix produces just trimmed empty + newline.
	got := GetBlockQuotePrefixText(top, 99, 1)
	if got != "\n" {
		t.Errorf("got %q, want %q", got, "\n")
	}
}

func TestIsHTMLFlowComment(t *testing.T) {
	tests := []struct {
		name string
		typ  mm.TokenType
		text string
		want bool
	}{
		{"valid_comment", mm.TypeHTMLFlow, "<!-- hello -->", true},
		{"not_htmlflow", mm.TypeParagraph, "<!-- hello -->", false},
		{"no_prefix", mm.TypeHTMLFlow, "<div>x</div>", false},
		{"starts_with_gt", mm.TypeHTMLFlow, "<!-->-->", false},
		{"starts_with_arrow", mm.TypeHTMLFlow, "<!--->-->", false},
		{"ends_with_dash", mm.TypeHTMLFlow, "<!--x--->", false},
		{"empty_comment", mm.TypeHTMLFlow, "<!---->", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &mm.Token{Type: tt.typ, Text: tt.text}
			if got := IsHTMLFlowComment(tok); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHTMLFlowCommentFromParse(t *testing.T) {
	top := parseTop(t, "<!-- a real comment -->\n")
	tok := firstOfType(top, mm.TypeHTMLFlow)
	if tok == nil {
		t.Fatal("no htmlFlow token")
	}
	if !IsHTMLFlowComment(tok) {
		t.Errorf("expected comment to be detected: %q", tok.Text)
	}
}

func TestIsDocfxTab(t *testing.T) {
	flat := parseFlat(t, "# [tab](#tab/csharp)\n")
	h := firstOfType(flat, mm.TypeAtxHeading)
	if !IsDocfxTab(h) {
		t.Errorf("expected docfx tab")
	}
}

func TestIsDocfxTabNegatives(t *testing.T) {
	tests := []struct {
		name string
		md   string
	}{
		{"plain_heading", "# Hello\n"},
		{"link_but_not_tab", "# [home](#home)\n"},
		{"setext", "Title\n=====\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flat := parseFlat(t, tt.md)
			h := firstOfType(flat, mm.TypeAtxHeading)
			if h == nil {
				h = firstOfType(flat, mm.TypeSetextHeading)
			}
			if IsDocfxTab(h) {
				t.Errorf("expected not docfx tab for %q", tt.md)
			}
		})
	}
}

func TestIsDocfxTabNil(t *testing.T) {
	if IsDocfxTab(nil) {
		t.Errorf("nil should not be a docfx tab")
	}
}

func TestNonContentTokens(t *testing.T) {
	// Spot-check membership and a non-member.
	want := []mm.TokenType{
		mm.TypeBlockQuoteMarker,
		mm.TypeBlockQuotePrefix,
		mm.TypeBlockQuotePrefixWhitespace,
		mm.TypeGfmFootnoteDefinitionIndent,
		mm.TypeLineEnding,
		mm.TypeLineEndingBlank,
		mm.TypeLinePrefix,
		mm.TypeListItemIndent,
		mm.TypeUndefinedReference,
		mm.TypeUndefinedReferenceCollapsed,
		mm.TypeUndefinedReferenceFull,
		mm.TypeUndefinedReferenceShortcut,
	}
	for _, typ := range want {
		if !NonContentTokens[typ] {
			t.Errorf("%q should be a non-content token", typ)
		}
	}
	if NonContentTokens[mm.TypeParagraph] {
		t.Errorf("paragraph should not be a non-content token")
	}
	if len(NonContentTokens) != len(want) {
		t.Errorf("set size = %d, want %d", len(NonContentTokens), len(want))
	}
}

// TestFilterByTypesResultStability sanity-checks that repeated filtering yields
// equal slices (no hidden state mutation).
func TestFilterByTypesResultStability(t *testing.T) {
	top := parseTop(t, "# One\n\n## Two\n")
	a := FilterByTypes(top, []mm.TokenType{mm.TypeAtxHeading}, false)
	b := FilterByTypes(top, []mm.TokenType{mm.TypeAtxHeading}, false)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("results differ between calls")
	}
}
