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

	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
)

func TestNormalizeReference(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Hello", "hello"},
		{"  Trim  Me  ", "trim me"},
		{"Multi   Space", "multi space"},
		{"With\tTab", "with tab"},
		{"line\nbreak", "line break"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeReference(tt.in); got != tt.want {
				t.Errorf("normalizeReference(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGetReferenceLinkImageData_FullReference(t *testing.T) {
	doc := mm.Parse("[text][ref]\n\n[ref]: https://example.com\n")
	d := GetReferenceLinkImageData(doc.Flat)

	if got, ok := d.References["ref"]; !ok || !reflect.DeepEqual(got, [][3]int{{0, 0, 0}}) {
		t.Errorf("References[ref] = %v (ok=%v), want [[0 0 0]]", got, ok)
	}
	if len(d.Shortcuts) != 0 {
		t.Errorf("expected no shortcuts, got %v", d.Shortcuts)
	}
	def, ok := d.Definitions["ref"]
	if !ok {
		t.Fatalf("missing definition for ref")
	}
	if def.LineIndex != 2 || def.Destination != "https://example.com" {
		t.Errorf("definition = %+v, want {2 https://example.com}", def)
	}
	if !reflect.DeepEqual(d.DefinitionLineIndices, []int{2}) {
		t.Errorf("DefinitionLineIndices = %v, want [2]", d.DefinitionLineIndices)
	}
	if len(d.DuplicateDefinitions) != 0 {
		t.Errorf("expected no duplicates, got %v", d.DuplicateDefinitions)
	}
}

func TestGetReferenceLinkImageData_Shortcut(t *testing.T) {
	doc := mm.Parse("[shortcut]\n\n[shortcut]: https://example.com\n")
	d := GetReferenceLinkImageData(doc.Flat)

	if got, ok := d.Shortcuts["shortcut"]; !ok || !reflect.DeepEqual(got, [][3]int{{0, 0, 0}}) {
		t.Errorf("Shortcuts[shortcut] = %v (ok=%v), want [[0 0 0]]", got, ok)
	}
	if len(d.References) != 0 {
		t.Errorf("expected no full references, got %v", d.References)
	}
	if _, ok := d.Definitions["shortcut"]; !ok {
		t.Errorf("missing definition for shortcut")
	}
}

func TestGetReferenceLinkImageData_Image(t *testing.T) {
	doc := mm.Parse("![img][ref]\n\n[ref]: https://example.com\n")
	d := GetReferenceLinkImageData(doc.Flat)

	if got, ok := d.References["ref"]; !ok || !reflect.DeepEqual(got, [][3]int{{0, 0, 0}}) {
		t.Errorf("References[ref] = %v (ok=%v), want [[0 0 0]]", got, ok)
	}
}

func TestGetReferenceLinkImageData_UndefinedFull(t *testing.T) {
	doc := mm.Parse("[undefined][nope]\n")
	d := GetReferenceLinkImageData(doc.Flat)

	got, ok := d.References["undefined"]
	if !ok {
		t.Fatalf("missing undefined reference; refs=%v shorts=%v", d.References, d.Shortcuts)
	}
	// Full undefined reference uses the whole token text length.
	if !reflect.DeepEqual(got, [][3]int{{0, 0, 17}}) {
		t.Errorf("References[undefined] = %v, want [[0 0 17]]", got)
	}
	if len(d.Definitions) != 0 {
		t.Errorf("expected no definitions, got %v", d.Definitions)
	}
}

func TestGetReferenceLinkImageData_UndefinedShortcut(t *testing.T) {
	doc := mm.Parse("[undefshort]\n")
	d := GetReferenceLinkImageData(doc.Flat)

	got, ok := d.Shortcuts["undefshort"]
	if !ok {
		t.Fatalf("missing undefined shortcut; refs=%v shorts=%v", d.References, d.Shortcuts)
	}
	if !reflect.DeepEqual(got, [][3]int{{0, 0, 12}}) {
		t.Errorf("Shortcuts[undefshort] = %v, want [[0 0 12]]", got)
	}
}

func TestGetReferenceLinkImageData_DuplicateDefinitions(t *testing.T) {
	doc := mm.Parse("[dup]: https://a.com\n[dup]: https://b.com\n\n[dup]\n")
	d := GetReferenceLinkImageData(doc.Flat)

	// First definition wins; the destination is from line 0.
	def, ok := d.Definitions["dup"]
	if !ok || def.LineIndex != 0 || def.Destination != "https://a.com" {
		t.Errorf("Definitions[dup] = %+v (ok=%v), want {0 https://a.com}", def, ok)
	}
	if !reflect.DeepEqual(d.DuplicateDefinitions, []DupDef{{Label: "dup", LineIndex: 1}}) {
		t.Errorf("DuplicateDefinitions = %v, want [{dup 1}]", d.DuplicateDefinitions)
	}
	if !reflect.DeepEqual(d.DefinitionLineIndices, []int{0, 1}) {
		t.Errorf("DefinitionLineIndices = %v, want [0 1]", d.DefinitionLineIndices)
	}
	// The [dup] usage on line 3 resolves to a shortcut.
	if got, ok := d.Shortcuts["dup"]; !ok || !reflect.DeepEqual(got, [][3]int{{3, 0, 0}}) {
		t.Errorf("Shortcuts[dup] = %v (ok=%v), want [[3 0 0]]", got, ok)
	}
}

func TestGetReferenceLinkImageData_Footnote(t *testing.T) {
	doc := mm.Parse("Here is a footnote[^1].\n\n[^1]: The note.\n")
	d := GetReferenceLinkImageData(doc.Flat)

	// The footnote call is recorded as a shortcut keyed by "^1".
	if got, ok := d.Shortcuts["^1"]; !ok || !reflect.DeepEqual(got, [][3]int{{0, 18, 0}}) {
		t.Errorf("Shortcuts[^1] = %v (ok=%v), want [[0 18 0]]", got, ok)
	}
	// The footnote definition is recorded (label prefixed with "^").
	if def, ok := d.Definitions["^1"]; !ok || def.LineIndex != 2 {
		t.Errorf("Definitions[^1] = %+v (ok=%v), want lineIndex 2", def, ok)
	}
	if !reflect.DeepEqual(d.DefinitionLineIndices, []int{2}) {
		t.Errorf("DefinitionLineIndices = %v, want [2]", d.DefinitionLineIndices)
	}
}

func TestGetReferenceLinkImageData_DuplicateFootnoteDefinition(t *testing.T) {
	doc := mm.Parse("[^a]: first\n[^a]: second\n\nUse [^a].\n")
	d := GetReferenceLinkImageData(doc.Flat)

	// First definition wins; second is recorded as a duplicate (label "^a").
	if def, ok := d.Definitions["^a"]; !ok || def.LineIndex != 0 {
		t.Errorf("Definitions[^a] = %+v (ok=%v), want lineIndex 0", def, ok)
	}
	if !reflect.DeepEqual(d.DuplicateDefinitions, []DupDef{{Label: "^a", LineIndex: 1}}) {
		t.Errorf("DuplicateDefinitions = %v, want [{^a 1}]", d.DuplicateDefinitions)
	}
	// The call on line 3 is recorded as a shortcut.
	if got, ok := d.Shortcuts["^a"]; !ok || !reflect.DeepEqual(got, [][3]int{{3, 4, 0}}) {
		t.Errorf("Shortcuts[^a] = %v (ok=%v), want [[3 4 0]]", got, ok)
	}
}

func TestGetReferenceLinkImageData_MultipleSameReference(t *testing.T) {
	doc := mm.Parse("A[^1] B[^1].\n\n[^1]: note.\n")
	d := GetReferenceLinkImageData(doc.Flat)
	// Two calls to the same footnote accumulate two data entries.
	if got, ok := d.Shortcuts["^1"]; !ok || !reflect.DeepEqual(got, [][3]int{{0, 1, 0}, {0, 7, 0}}) {
		t.Errorf("Shortcuts[^1] = %v (ok=%v), want [[0 1 0] [0 7 0]]", got, ok)
	}
}

func TestGetReferenceLinkImageData_Empty(t *testing.T) {
	doc := mm.Parse("Just plain text with no links.\n")
	d := GetReferenceLinkImageData(doc.Flat)
	if len(d.References) != 0 || len(d.Shortcuts) != 0 || len(d.Definitions) != 0 {
		t.Errorf("expected empty data, got %+v", d)
	}
	if d.References == nil || d.Shortcuts == nil || d.Definitions == nil {
		t.Error("maps should be initialized, not nil")
	}
}

func TestRefTokenText(t *testing.T) {
	if got := refTokenText(nil); got != "" {
		t.Errorf("refTokenText(nil) = %q, want empty", got)
	}
	tok := &mm.Token{
		Children: []*mm.Token{
			{Type: mm.TypeData, Text: "ab"},
			{Type: mm.TypeBlockQuotePrefix, Text: "> "}, // excluded
			{Type: mm.TypeData, Text: "cd"},
		},
	}
	if got := refTokenText(tok); got != "abcd" {
		t.Errorf("refTokenText = %q, want abcd", got)
	}
}

func TestAnyChildType(t *testing.T) {
	tok := &mm.Token{Children: []*mm.Token{{Type: mm.TypeResource}}}
	if !anyChildType(tok, mm.TypeResource) {
		t.Error("expected resource child to be found")
	}
	if anyChildType(tok, mm.TypeLabel) {
		t.Error("did not expect label child")
	}
}

func TestFirstDescendant(t *testing.T) {
	child := &mm.Token{Type: mm.TypeLabelText, Text: "inner"}
	label := &mm.Token{Type: mm.TypeLabel, Children: []*mm.Token{child}}
	root := &mm.Token{Type: mm.TypeLink, Children: []*mm.Token{label}}

	if got := firstDescendant(root, mm.TypeLabel, mm.TypeLabelText); got != child {
		t.Errorf("firstDescendant = %v, want child %v", got, child)
	}
	if got := firstDescendant(root, mm.TypeReference); got != nil {
		t.Errorf("firstDescendant(missing) = %v, want nil", got)
	}
}
