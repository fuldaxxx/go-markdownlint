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

package helpers

import (
	"strings"
	"testing"
)

func TestStringify(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"string", "hello", "hello"},
		{"empty string", "", ""},
		{"int", 42, "42"},
		{"negative int", -7, "-7"},
		{"int64", int64(123456789012), "123456789012"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"float whole", 3.0, "3"},
		{"float fractional", 3.5, "3.5"},
		{"float negative whole", -10.0, "-10"},
		{"nil", nil, ""},
		{"fallback uint", uint(9), "9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stringify(tt.in); got != tt.want {
				t.Errorf("Stringify(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEllipsify(t *testing.T) {
	long := "0123456789012345678901234567890123456789" // 40 ASCII chars
	tests := []struct {
		name       string
		text       string
		start, end bool
		want       string
	}{
		{"short returns as-is", "short text", false, false, "short text"},
		{"exactly 30 returns as-is", strings.Repeat("a", 30), true, true, strings.Repeat("a", 30)},
		{"start only", long, true, false, long[:30] + "..."},
		{"end only", long, false, true, "..." + long[len(long)-30:]},
		{"both", long, true, true, long[:15] + "..." + long[len(long)-15:]},
		{"neither defaults to start", long, false, false, long[:30] + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ellipsify(tt.text, tt.start, tt.end); got != tt.want {
				t.Errorf("Ellipsify(%q, %v, %v) = %q, want %q", tt.text, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestEllipsifyMultibyte(t *testing.T) {
	// 40 multibyte runes; operations must be rune-based, not byte-based.
	r := []rune(strings.Repeat("é", 40))
	text := string(r)
	got := Ellipsify(text, true, true)
	want := string(r[:15]) + "..." + string(r[len(r)-15:])
	if got != want {
		t.Errorf("Ellipsify multibyte = %q, want %q", got, want)
	}
	// A 30-rune multibyte string (but >30 bytes) must be returned unchanged.
	short := string([]rune(strings.Repeat("ü", 30)))
	if got := Ellipsify(short, true, true); got != short {
		t.Errorf("Ellipsify 30-rune multibyte changed string: %q", got)
	}
}

func TestIsString(t *testing.T) {
	if !IsString("x") {
		t.Error("IsString(string) should be true")
	}
	if IsString(3) {
		t.Error("IsString(int) should be false")
	}
	if IsString(nil) {
		t.Error("IsString(nil) should be false")
	}
}

func TestIsEmptyString(t *testing.T) {
	if !IsEmptyString("") {
		t.Error("empty should be empty")
	}
	if IsEmptyString(" ") {
		t.Error("space is not empty")
	}
}

func TestIsBlankLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{"empty", "", true},
		{"spaces", "   ", true},
		{"tabs", "\t\t", true},
		{"text", "hello", false},
		{"only gt", ">", true},
		{"gt with spaces", "  >  > ", true},
		{"block comment only", "<!-- comment -->", true},
		{"comment with gt", "> <!-- c -->", true},
		{"comment then text", "<!-- c --> word", false},
		{"text then comment", "word <!-- c -->", false},
		{"unterminated comment removes rest", "before <!-- after", false},
		{"unterminated comment only after marker", "<!-- trailing", true},
		{"end marker before start", "--> tail", false},
		{"end then start", "--> <!--", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlankLine(tt.line); got != tt.want {
				t.Errorf("IsBlankLine(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestClearHTMLCommentText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no comment unchanged", "plain text", "plain text"},
		// Only non-space, non-newline characters are replaced with "."; spaces
		// inside the comment are preserved.
		{"simple block comment cleared", "<!-- secret -->", "<!-- ...... -->"},
		{"inline comment in paragraph", "text <!-- hide -->", "text <!-- .... -->"},
		{"empty comment unchanged", "<!---->", "<!---->"},
		{"unterminated comment unchanged", "<!-- no end", "<!-- no end"},
		{"preserves spaces converting newlines kept", "<!-- a\nb -->", "<!-- .\n. -->"},
		// A space immediately before a newline is itself replaced with "." so it
		// is not mistaken for trailing whitespace (trailingSpaceRe branch).
		{"trailing space before newline cleared", "<!-- a \nb -->", "<!-- ..\n. -->"},
		// invalid (content starts with ">") is not cleared
		{"directive-like not cleared", "<!-->bad-->", "<!-->bad-->"},
		// In a non-block (inline) context, content containing "--" is invalid
		// and not cleared. As a block-level comment the "--" check is skipped.
		{"inline double dash not cleared", "x <!--a--b-->", "x <!--a--b-->"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClearHTMLCommentText(tt.in); got != tt.want {
				t.Errorf("ClearHTMLCommentText(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestClearHTMLCommentTextTableCells(t *testing.T) {
	// A comment that begins inside a table cell line (starts with "|") and
	// spans multiple cells (contains a newline) is NOT cleared.
	in := "| <!-- a\nb --> |"
	if got := ClearHTMLCommentText(in); got != in {
		t.Errorf("table-spanning comment should be unchanged, got %q", got)
	}
	// A comment inside a table cell that does not span cells IS cleared.
	in2 := "| <!-- hide --> |"
	want2 := "| <!-- .... --> |"
	if got := ClearHTMLCommentText(in2); got != want2 {
		t.Errorf("single-cell comment = %q, want %q", got, want2)
	}
}

func TestEscapeForRegExp(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"a.b", `a\.b`},
		{"(x)", `\(x\)`},
		{"a+b*c?", `a\+b\*c\?`},
		{`a\b`, `a\\b`},
		{"[abc]", `\[abc\]`},
		{"a|b", `a\|b`},
		{"$^", `\$\^`},
		{"plain", "plain"},
		{"a-b/c{d}", `a\-b\/c\{d\}`},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := EscapeForRegExp(tt.in); got != tt.want {
				t.Errorf("EscapeForRegExp(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetHTMLAttributeRe(t *testing.T) {
	re := GetHTMLAttributeRe("href")
	tests := []struct {
		name      string
		input     string
		wantMatch bool
		wantVal   string
	}{
		{"double quotes", `<a href="https://x.com">`, true, "https://x.com"},
		{"single quotes", `<a href='val'>`, true, "val"},
		{"no quotes", `<a href=plain>`, true, "plain"},
		{"case insensitive", `<a HREF="up">`, true, "up"},
		{"absent", `<a class="c">`, false, ""},
		{"spaces around equals", `<a href = "spaced">`, true, "spaced"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := re.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if m == nil {
					t.Fatalf("expected match for %q", tt.input)
				}
				if m[1] != tt.wantVal {
					t.Errorf("captured %q, want %q", m[1], tt.wantVal)
				}
			} else if m != nil {
				t.Errorf("expected no match for %q, got %v", tt.input, m)
			}
		})
	}
}

func TestFrontMatterHasTitle(t *testing.T) {
	tests := []struct {
		name       string
		lines      []string
		pattern    string
		patternSet bool
		want       bool
	}{
		{"default finds title", []string{"title: Hello"}, "", false, true},
		{"default quoted title", []string{`"title": Hello`}, "", false, true},
		{"default case insensitive", []string{"TITLE = Hi"}, "", false, true},
		{"default no title", []string{"author: Me"}, "", false, false},
		{"empty pattern set disables", []string{"title: Hello"}, "", true, false},
		{"custom pattern matches", []string{"name: foo"}, `^name\s*:`, true, true},
		{"custom pattern no match", []string{"title: foo"}, `^name\s*:`, true, false},
		{"empty lines no title", nil, "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FrontMatterHasTitle(tt.lines, tt.pattern, tt.patternSet); got != tt.want {
				t.Errorf("FrontMatterHasTitle(%v, %q, %v) = %v, want %v",
					tt.lines, tt.pattern, tt.patternSet, got, tt.want)
			}
		})
	}
}

func TestHasOverlap(t *testing.T) {
	mk := func(sl, sc, el, ec int) FileRange {
		return FileRange{StartLine: sl, StartColumn: sc, EndLine: el, EndColumn: ec}
	}
	tests := []struct {
		name string
		a, b FileRange
		want bool
	}{
		{"overlapping same line", mk(1, 1, 1, 10), mk(1, 5, 1, 15), true},
		{"disjoint same line", mk(1, 1, 1, 4), mk(1, 6, 1, 10), false},
		{"touching at boundary", mk(1, 1, 1, 5), mk(1, 5, 1, 10), true},
		{"disjoint different lines", mk(1, 1, 1, 10), mk(3, 1, 3, 10), false},
		{"b before a still detected", mk(2, 1, 2, 10), mk(1, 1, 1, 5), false},
		{"contained range", mk(1, 1, 5, 10), mk(2, 1, 3, 5), true},
		{"multi-line overlap", mk(1, 5, 3, 5), mk(2, 1, 4, 1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasOverlap(tt.a, tt.b); got != tt.want {
				t.Errorf("HasOverlap(%+v, %+v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			// Symmetry: HasOverlap should give the same answer regardless of order.
			if got := HasOverlap(tt.b, tt.a); got != tt.want {
				t.Errorf("HasOverlap(swapped) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandTildePath(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		homedir string
		want    string
	}{
		{"no homedir returns file", "~/foo", "", "~/foo"},
		{"bare tilde", "~", "/home/u", "/home/u"},
		{"tilde slash", "~/docs/a.md", "/home/u", "/home/u/docs/a.md"},
		{"tilde backslash", `~\docs\a.md`, "/home/u", `/home/u\docs\a.md`},
		{"no tilde unchanged", "/abs/path", "/home/u", "/abs/path"},
		{"tilde in middle unchanged", "a/~/b", "/home/u", "a/~/b"},
		{"tildeuser unchanged", "~user/x", "/home/u", "~user/x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandTildePath(tt.file, tt.homedir); got != tt.want {
				t.Errorf("ExpandTildePath(%q, %q) = %q, want %q", tt.file, tt.homedir, got, tt.want)
			}
		})
	}
}
