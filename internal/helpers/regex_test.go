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
	"reflect"
	"testing"
)

func TestNewLineRe(t *testing.T) {
	got := NewLineRe.FindAllString("a\nb\r\nc\rd", -1)
	want := []string{"\n", "\r\n", "\r"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewLineRe matches = %v, want %v", got, want)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{"lf", "a\nb\nc", []string{"a", "b", "c"}},
		{"crlf", "a\r\nb", []string{"a", "b"}},
		{"cr", "a\rb", []string{"a", "b"}},
		{"mixed", "a\nb\r\nc\rd", []string{"a", "b", "c", "d"}},
		{"trailing newline yields empty element", "a\n", []string{"a", ""}},
		{"empty string", "", []string{""}},
		{"no newline", "single", []string{"single"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitLines(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitLines(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestNextLinesRe(t *testing.T) {
	got := NextLinesRe.ReplaceAllString("first\nsecond\nthird", "")
	if got != "first" {
		t.Errorf("NextLinesRe = %q, want %q", got, "first")
	}
	// No newline: unchanged.
	if got := NextLinesRe.ReplaceAllString("only", ""); got != "only" {
		t.Errorf("NextLinesRe no-newline = %q, want %q", got, "only")
	}
}

func TestFrontMatterRe(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"yaml", "---\ntitle: x\n---\nbody", true},
		{"toml", "+++\ntitle = \"x\"\n+++\nbody", true},
		{"json", "{\n\"title\": \"x\"\n}\nbody", true},
		{"no front matter", "# Heading\n", false},
		{"dashes not at start", "text\n---\nx\n---\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := FrontMatterRe.FindStringIndex(tt.input)
			matched := loc != nil && loc[0] == 0
			if matched != tt.wantMatch {
				t.Errorf("FrontMatterRe(%q) matched=%v, want %v", tt.input, matched, tt.wantMatch)
			}
		})
	}
}

func TestInlineCommentStartRe(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMatch bool
	}{
		{"disable", "<!-- markdownlint-disable -->", true},
		{"enable", "<!-- markdownlint-enable MD001 -->", true},
		{"configure-file", "<!-- markdownlint-configure-file {} -->", true},
		{"disable-next-line", "<!-- markdownlint-disable-next-line -->", true},
		{"case insensitive", "<!-- MARKDOWNLINT-DISABLE -->", true},
		{"not a config comment", "<!-- regular comment -->", false},
		{"unknown action", "<!-- markdownlint-foo -->", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InlineCommentStartRe.MatchString(tt.input); got != tt.wantMatch {
				t.Errorf("InlineCommentStartRe(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
		})
	}
}

func TestEndOfLineHTMLEntityRe(t *testing.T) {
	tests := []struct {
		input     string
		wantMatch bool
	}{
		{"&amp;", true},
		{"&#169;", true},
		{"&#xAB;", true},
		{"text &copy;", true},
		{"&amp; more", false},
		{"no entity", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := EndOfLineHTMLEntityRe.MatchString(tt.input); got != tt.wantMatch {
				t.Errorf("EndOfLineHTMLEntityRe(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
		})
	}
}

func TestEndOfLineGemojiCodeRe(t *testing.T) {
	tests := []struct {
		input     string
		wantMatch bool
	}{
		{":smile:", true},
		{":+1:", true},
		{"end with :tada:", true},
		{":smile: trailing", false},
		{"no gemoji", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := EndOfLineGemojiCodeRe.MatchString(tt.input); got != tt.wantMatch {
				t.Errorf("EndOfLineGemojiCodeRe(%q) = %v, want %v", tt.input, got, tt.wantMatch)
			}
		})
	}
}

func TestPunctuationConstants(t *testing.T) {
	if AllPunctuation != ".,;:!?。，；：！？" {
		t.Errorf("AllPunctuation = %q", AllPunctuation)
	}
	if AllPunctuationNoQuestion != ".,;:!。，；：！" {
		t.Errorf("AllPunctuationNoQuestion = %q", AllPunctuationNoQuestion)
	}
}
