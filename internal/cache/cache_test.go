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

package cache

import (
	"reflect"
	"testing"

	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
)

func parse(md string) *mm.Document { return mm.Parse(md) }

func TestNewAndTokens(t *testing.T) {
	doc := parse("# Heading\n\nbody\n")
	c := New(doc.Flat)
	if c == nil {
		t.Fatal("New returned nil")
	}
	if !reflect.DeepEqual(c.Tokens(), doc.Flat) {
		t.Errorf("Tokens() != doc.Flat")
	}
}

func TestNewEmpty(t *testing.T) {
	c := New(nil)
	if got := c.Tokens(); len(got) != 0 {
		t.Errorf("want empty, got %d", len(got))
	}
	got := c.FilterByTypes([]mm.TokenType{mm.TypeParagraph}, false)
	if len(got) != 0 {
		t.Errorf("want 0, got %d", len(got))
	}
}

func TestFilterByTypesMemoization(t *testing.T) {
	doc := parse("# One\n\n## Two\n")
	c := New(doc.Flat)

	first := c.FilterByTypes([]mm.TokenType{mm.TypeAtxHeading}, false)
	if len(first) != 2 {
		t.Fatalf("want 2 headings, got %d", len(first))
	}
	second := c.FilterByTypes([]mm.TokenType{mm.TypeAtxHeading}, false)

	// Memoized call must return the identical slice (same backing array & len).
	if &first[0] != &second[0] || len(first) != len(second) {
		t.Errorf("memoized call returned a different slice")
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("results differ")
	}
}

func TestFilterByTypesDistinctKeys(t *testing.T) {
	doc := parse("# H\n\ntext\n")
	c := New(doc.Flat)

	headings := c.FilterByTypes([]mm.TokenType{mm.TypeAtxHeading}, false)
	paragraphs := c.FilterByTypes([]mm.TokenType{mm.TypeParagraph}, false)

	if len(headings) != 1 {
		t.Errorf("headings = %d, want 1", len(headings))
	}
	if len(paragraphs) != 1 {
		t.Errorf("paragraphs = %d, want 1", len(paragraphs))
	}
}

func TestFilterByTypesHTMLFlowKey(t *testing.T) {
	doc := parse("<!-- comment -->\n")
	c := New(doc.Flat)
	// Different htmlFlow flag => different cache key.
	a := c.FilterByTypes([]mm.TokenType{mm.TypeHTMLFlow}, false)
	b := c.FilterByTypes([]mm.TokenType{mm.TypeHTMLFlow}, true)
	if len(a) != 1 || len(b) != 1 {
		t.Errorf("a=%d b=%d, want 1 each", len(a), len(b))
	}
}

func TestReferenceLinkImageDataMemoization(t *testing.T) {
	doc := parse("[txt][ref]\n\n[ref]: http://example.com\n")
	c := New(doc.Flat)

	first := c.ReferenceLinkImageData()
	if first == nil {
		t.Fatal("nil reference data")
	}
	second := c.ReferenceLinkImageData()
	// Memoized: identical pointer.
	if first != second {
		t.Errorf("expected memoized identical pointer")
	}

	if _, ok := first.Definitions["ref"]; !ok {
		t.Errorf("expected definition 'ref', got %#v", first.Definitions)
	}
	if _, ok := first.References["ref"]; !ok {
		t.Errorf("expected reference 'ref', got %#v", first.References)
	}
}

func TestReferenceLinkImageDataEmpty(t *testing.T) {
	doc := parse("just text\n")
	c := New(doc.Flat)
	data := c.ReferenceLinkImageData()
	if data == nil {
		t.Fatal("nil data")
	}
	if len(data.Definitions) != 0 || len(data.References) != 0 {
		t.Errorf("expected empty, got %#v", data)
	}
}
