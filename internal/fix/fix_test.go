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

package fix

import (
	"testing"

	"github.com/ldmonster/go-markdownlint/internal/types"
)

// fi is a small constructor helper for FixInfo literals.
func fi(line, col, del int, text string) types.FixInfo {
	return types.FixInfo{LineNumber: line, EditColumn: col, DeleteCount: del, InsertText: text}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name       string
		in         types.FixInfo
		lineNumber int
		want       types.FixInfo
	}{
		{
			name:       "all zero uses defaults and provided lineNumber",
			in:         types.FixInfo{},
			lineNumber: 7,
			want:       types.FixInfo{LineNumber: 7, EditColumn: 1, DeleteCount: 0, InsertText: ""},
		},
		{
			name:       "explicit lineNumber wins over argument",
			in:         types.FixInfo{LineNumber: 3},
			lineNumber: 7,
			want:       types.FixInfo{LineNumber: 3, EditColumn: 1},
		},
		{
			name:       "editColumn 0 normalizes to 1",
			in:         types.FixInfo{LineNumber: 1, EditColumn: 0},
			lineNumber: 0,
			want:       types.FixInfo{LineNumber: 1, EditColumn: 1},
		},
		{
			name:       "non-zero fields are preserved",
			in:         types.FixInfo{LineNumber: 2, EditColumn: 5, DeleteCount: 3, InsertText: "x"},
			lineNumber: 9,
			want:       types.FixInfo{LineNumber: 2, EditColumn: 5, DeleteCount: 3, InsertText: "x"},
		},
		{
			name:       "deleteCount -1 preserved",
			in:         types.FixInfo{LineNumber: 4, DeleteCount: -1},
			lineNumber: 0,
			want:       types.FixInfo{LineNumber: 4, EditColumn: 1, DeleteCount: -1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalize(tc.in, tc.lineNumber)
			if got != tc.want {
				t.Fatalf("normalize(%+v, %d) = %+v, want %+v", tc.in, tc.lineNumber, got, tc.want)
			}
		})
	}
}

func TestApplyFix(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		fi         types.FixInfo
		lineEnding string
		wantLine   string
		wantOK     bool
	}{
		{
			name:     "insert at editColumn",
			line:     "abc",
			fi:       fi(0, 2, 0, "X"),
			wantLine: "aXbc",
			wantOK:   true,
		},
		{
			name:     "default editColumn 0 inserts at start (col 1)",
			line:     "abc",
			fi:       fi(0, 0, 0, "X"),
			wantLine: "Xabc",
			wantOK:   true,
		},
		{
			name:     "insert at column 1 explicitly",
			line:     "abc",
			fi:       fi(0, 1, 0, ">"),
			wantLine: ">abc",
			wantOK:   true,
		},
		{
			name:     "delete with deleteCount>0",
			line:     "abcdef",
			fi:       fi(0, 2, 2, ""),
			wantLine: "adef",
			wantOK:   true,
		},
		{
			name:     "replace = delete + insert",
			line:     "abcdef",
			fi:       fi(0, 2, 2, "XY"),
			wantLine: "aXYdef",
			wantOK:   true,
		},
		{
			name:     "whole-line delete returns ok=false",
			line:     "abc",
			fi:       fi(0, 1, -1, ""),
			wantLine: "",
			wantOK:   false,
		},
		{
			name:     "editColumn beyond line length clamps to end (append)",
			line:     "abc",
			fi:       fi(0, 99, 0, "Z"),
			wantLine: "abcZ",
			wantOK:   true,
		},
		{
			name:     "deleteCount beyond end clamps",
			line:     "abc",
			fi:       fi(0, 2, 99, "Z"),
			wantLine: "aZ",
			wantOK:   true,
		},
		{
			name:       "\\n in insertText replaced with lineEnding default",
			line:       "abc",
			fi:         fi(0, 4, 0, "\nX"),
			lineEnding: "\n",
			wantLine:   "abc\nX",
			wantOK:     true,
		},
		{
			name:       "\\n in insertText replaced with CRLF lineEnding",
			line:       "abc",
			fi:         fi(0, 4, 0, "\nX"),
			lineEnding: "\r\n",
			wantLine:   "abc\r\nX",
			wantOK:     true,
		},
		{
			name:     "empty line, insert",
			line:     "",
			fi:       fi(0, 1, 0, "hi"),
			wantLine: "hi",
			wantOK:   true,
		},
		{
			name:     "multibyte/rune columns: emoji line, delete second rune",
			line:     "a😀b",
			fi:       fi(0, 2, 1, ""),
			wantLine: "ab",
			wantOK:   true,
		},
		{
			name:     "multibyte/rune columns: insert after emoji",
			line:     "😀😀",
			fi:       fi(0, 2, 0, "X"),
			wantLine: "😀X😀",
			wantOK:   true,
		},
		{
			name:     "multibyte/rune columns: replace emoji with ascii",
			line:     "x😀y",
			fi:       fi(0, 2, 1, "Z"),
			wantLine: "xZy",
			wantOK:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			le := tc.lineEnding
			if le == "" {
				le = "\n"
			}
			got, ok := ApplyFix(tc.line, tc.fi, le)
			if got != tc.wantLine || ok != tc.wantOK {
				t.Fatalf("ApplyFix(%q, %+v, %q) = (%q, %v), want (%q, %v)",
					tc.line, tc.fi, le, got, ok, tc.wantLine, tc.wantOK)
			}
		})
	}
}

// errWithFix builds a types.Error carrying a FixInfo, mirroring how rules emit
// fixes into ApplyFixes.
func errWithFix(lineNumber int, f types.FixInfo) types.Error {
	cp := f
	return types.Error{LineNumber: lineNumber, FixInfo: &cp}
}

func TestApplyFixes(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		errors []types.Error
		want   string
	}{
		{
			name:   "no fixes returns input unchanged",
			input:  "alpha\nbeta\n",
			errors: nil,
			want:   "alpha\nbeta\n",
		},
		{
			name:  "error without FixInfo is ignored",
			input: "alpha\nbeta",
			errors: []types.Error{
				{LineNumber: 1, FixInfo: nil},
			},
			want: "alpha\nbeta",
		},
		{
			name:  "single insert uses error lineNumber when FixInfo.LineNumber==0",
			input: "alpha\nbeta",
			errors: []types.Error{
				errWithFix(2, fi(0, 1, 0, "X")),
			},
			want: "alpha\nXbeta",
		},
		{
			name:  "FixInfo.LineNumber overrides error lineNumber",
			input: "alpha\nbeta",
			errors: []types.Error{
				errWithFix(1, fi(2, 1, 0, "X")),
			},
			want: "alpha\nXbeta",
		},
		{
			name:  "whole-line delete removes the line",
			input: "alpha\nbeta\ngamma",
			errors: []types.Error{
				errWithFix(2, fi(0, 0, -1, "")),
			},
			want: "alpha\ngamma",
		},
		{
			name:  "bottom-to-top ordering: two inserts on different lines",
			input: "a\nb\nc",
			errors: []types.Error{
				errWithFix(1, fi(0, 1, 0, "1")),
				errWithFix(3, fi(0, 1, 0, "3")),
			},
			want: "1a\nb\n3c",
		},
		{
			name:  "de-duplication of identical fixes applies once",
			input: "abc",
			errors: []types.Error{
				errWithFix(1, fi(0, 2, 0, "-")),
				errWithFix(1, fi(0, 2, 0, "-")),
			},
			want: "a-bc",
		},
		{
			// insert (no delete) + delete (no insert) at same position collapse
			// into a single replace.
			name:  "insert+delete collapse at same position",
			input: "abc",
			errors: []types.Error{
				errWithFix(1, fi(0, 2, 0, "X")), // insert "X" at col 2
				errWithFix(1, fi(0, 2, 1, "")),  // delete 1 char at col 2
			},
			want: "aXc",
		},
		{
			// Two overlapping fixes on the same line: only the first (sorted)
			// should be applied; the overlapping one is skipped.
			name:  "overlap skipping: only one applied",
			input: "abcdef",
			errors: []types.Error{
				errWithFix(1, fi(0, 2, 3, "X")), // replace cols 2-4
				errWithFix(1, fi(0, 3, 1, "Y")), // overlaps cols 3
			},
			// Sorted right-to-left: col 3 fix applies first -> "abYdef",
			// then col 2 fix overlaps (editIndex+deleteCount=1+3=4 > lastEditIndex=2),
			// so it is skipped.
			want: "abYdef",
		},
		{
			name:  "non-overlapping fixes on same line both applied (right-to-left)",
			input: "abcdef",
			errors: []types.Error{
				errWithFix(1, fi(0, 2, 0, "X")),
				errWithFix(1, fi(0, 5, 0, "Y")),
			},
			want: "aXbcdYef",
		},
		{
			name:  "line-delete sorts last relative to same-line edits",
			input: "abc\ndef",
			errors: []types.Error{
				errWithFix(1, fi(0, 1, -1, "")), // delete line 1
				errWithFix(1, fi(0, 2, 0, "Z")), // edit line 1
			},
			// Edit applies first (delete sorts last), then the line is removed.
			want: "def",
		},
		{
			name:  "longest-insert-first at same position",
			input: "abc",
			errors: []types.Error{
				errWithFix(1, fi(0, 2, 0, "S")),
				errWithFix(1, fi(0, 2, 0, "LONG")),
			},
			// Both are inserts at the same column (col 2) with deleteCount 0.
			// Sorted longest-first: "LONG" applies; "S" overlaps and is skipped
			// (editIndex+0=1 <= lastEditIndex-1=0 is false).
			want: "aLONGbc",
		},
		{
			name:  "CRLF input preserves CRLF line endings",
			input: "alpha\r\nbeta\r\ngamma",
			errors: []types.Error{
				errWithFix(2, fi(0, 1, 0, "X")),
			},
			want: "alpha\r\nXbeta\r\ngamma",
		},
		{
			name:  "CRLF input: inserted newline uses CRLF",
			input: "alpha\r\nbeta",
			errors: []types.Error{
				errWithFix(1, fi(0, 6, 0, "\nmid")),
			},
			want: "alpha\r\nmid\r\nbeta",
		},
		{
			name:  "out-of-range lineNumber is skipped",
			input: "alpha\nbeta",
			errors: []types.Error{
				errWithFix(99, fi(0, 1, 0, "X")),
				errWithFix(0, fi(0, 1, 0, "Y")), // lineIndex -1 -> skipped
			},
			want: "alpha\nbeta",
		},
		{
			name:  "multiple line deletes remove all referenced lines",
			input: "a\nb\nc\nd",
			errors: []types.Error{
				errWithFix(2, fi(0, 1, -1, "")),
				errWithFix(4, fi(0, 1, -1, "")),
			},
			want: "a\nc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyFixes(tc.input, tc.errors)
			if got != tc.want {
				t.Fatalf("ApplyFixes(%q, ...) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestApplyFixesFrontMatterOffset verifies that FixInfo.LineNumber (which rules
// set as an absolute content line number, already accounting for any front
// matter offset applied upstream) is honored over the error's LineNumber.
func TestApplyFixesFrontMatterOffset(t *testing.T) {
	input := "---\ntitle: x\n---\nbody1\nbody2"
	// A rule reporting on content line 2 (body2) but whose FixInfo carries the
	// absolute line number 5 within the full document including front matter.
	errs := []types.Error{
		errWithFix(2, fi(5, 1, 0, "> ")),
	}
	got := ApplyFixes(input, errs)
	want := "---\ntitle: x\n---\nbody1\n> body2"
	if got != want {
		t.Fatalf("front-matter offset: got %q, want %q", got, want)
	}
}

// TestApplyFixInsert checks a representative insert against the applyFix
// semantics:
//
//	line[:editIndex] + insertText (with newlines expanded to lineEnding) + line[editIndex+deleteCount:]
//
// For line "Hello", editColumn 6 (editIndex 5), deleteCount 0, insert " world":
// "Hello" + " world" + "" => "Hello world".
func TestApplyFixInsert(t *testing.T) {
	got, ok := ApplyFix("Hello", fi(0, 6, 0, " world"), "\n")
	if !ok || got != "Hello world" {
		t.Fatalf("got (%q,%v), want (%q,true)", got, ok, "Hello world")
	}
}

// TestApplyFixReplaceNewline checks replace + newline expansion.
// ApplyFix("# Title", {editColumn:1, deleteCount:2, insertText:"## "}) =>
// "" + "## " + "Title" => "## Title".
func TestApplyFixReplaceNewline(t *testing.T) {
	got, ok := ApplyFix("# Title", fi(0, 1, 2, "## "), "\n")
	if !ok || got != "## Title" {
		t.Fatalf("got (%q,%v), want (%q,true)", got, ok, "## Title")
	}
}

// NOTE: ApplyFix clamps editIndex and endIndex to the rune length of the line
// (fix.go lines 38-44), silently tolerating out-of-range indices. For the
// inputs markdownlint rules actually produce (columns derived from the same
// line) this never matters; this test pins the clamping behavior for an
// editColumn far beyond the line so a future change is caught.
func TestApplyFixClampCharacterization(t *testing.T) {
	// editColumn 100 on a 3-rune line clamps to appending at the end.
	got, ok := ApplyFix("abc", fi(0, 100, 5, "!"), "\n")
	if !ok || got != "abc!" {
		t.Fatalf("clamp characterization: got (%q,%v), want (%q,true)", got, ok, "abc!")
	}
}
