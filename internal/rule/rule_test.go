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

package rule

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/fuldaxxx/go-markdownlint/internal/cache"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

// buildParams parses md and returns a fully wired RuleParams plus the doc.
func buildParams(t *testing.T, md string) (*RuleParams, *mm.Document) {
	t.Helper()
	doc := mm.Parse(md)
	c := cache.New(doc.Flat)
	lines := []string{"line1", "line2"}
	fm := []string{"---", "title: x", "---"}
	style := "atx"
	cfg := types.Configuration{MD003: types.MD003Config{Style: &style}}
	p := NewRuleParams("MD001", "1.2.3", lines, fm, cfg, doc.Children, c)
	return p, doc
}

func TestNewRuleParamsFields(t *testing.T) {
	p, doc := buildParams(t, "# Heading\n")
	if p.Name != "MD001" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.Version != "1.2.3" {
		t.Errorf("Version = %q", p.Version)
	}
	if !reflect.DeepEqual(p.Lines, []string{"line1", "line2"}) {
		t.Errorf("Lines = %#v", p.Lines)
	}
	if !reflect.DeepEqual(p.FrontMatterLines, []string{"---", "title: x", "---"}) {
		t.Errorf("FrontMatterLines = %#v", p.FrontMatterLines)
	}
	if p.Config.MD003.Style == nil || *p.Config.MD003.Style != "atx" {
		t.Errorf("Config = %#v", p.Config)
	}
	if !reflect.DeepEqual(p.MicromarkTokens(), doc.Children) {
		t.Errorf("MicromarkTokens != doc.Children")
	}
}

func TestMicromarkTokens(t *testing.T) {
	p, _ := buildParams(t, "# One\n\n## Two\n")
	top := p.MicromarkTokens()
	// Two top-level headings.
	count := 0
	for _, tok := range top {
		if tok.Type == mm.TypeAtxHeading {
			count++
		}
	}
	if count != 2 {
		t.Errorf("want 2 top-level headings, got %d", count)
	}
}

func TestFilterByTypesCached(t *testing.T) {
	p, _ := buildParams(t, "# One\n\n## Two\n")
	headings := p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading}, false)
	if len(headings) != 2 {
		t.Fatalf("want 2, got %d", len(headings))
	}
	// Second call is served from the shared cache (identical slice).
	again := p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading}, false)
	if &headings[0] != &again[0] || len(headings) != len(again) {
		t.Errorf("expected cached identical slice")
	}
}

func TestFilterByTypesCachedNoMatch(t *testing.T) {
	p, _ := buildParams(t, "# H\n")
	got := p.FilterByTypesCached([]mm.TokenType{mm.TypeTable}, false)
	if len(got) != 0 {
		t.Errorf("want 0, got %d", len(got))
	}
}

func TestReferenceLinkImageData(t *testing.T) {
	doc := mm.Parse("[txt][ref]\n\n[ref]: http://example.com\n")
	c := cache.New(doc.Flat)
	p := NewRuleParams("MD052", "1.0.0", nil, nil, types.Configuration{}, doc.Children, c)

	data := p.ReferenceLinkImageData()
	if data == nil {
		t.Fatal("nil reference data")
	}
	if _, ok := data.Definitions["ref"]; !ok {
		t.Errorf("expected definition 'ref', got %#v", data.Definitions)
	}
	// Memoized through the shared cache: identical pointer on repeat.
	if p.ReferenceLinkImageData() != data {
		t.Errorf("expected memoized identical pointer")
	}
}

func TestRuleDescriptor(t *testing.T) {
	info, _ := url.Parse("https://example.com/md001")
	called := false
	r := Rule{
		Names:        []string{"MD001", "heading-increment"},
		Description:  "Heading levels should only increment by one level at a time",
		Tags:         []string{"headings"},
		Parser:       types.ParserMicromark,
		Information:  info,
		Asynchronous: false,
		Fn: func(p *RuleParams, onError types.OnError) {
			called = true
			onError(types.ErrorInfo{LineNumber: 1})
		},
	}
	if r.Names[0] != "MD001" || r.Parser != types.ParserMicromark {
		t.Errorf("descriptor = %#v", r)
	}

	p, _ := buildParams(t, "# H\n")
	var errs []types.ErrorInfo
	r.Fn(p, func(e types.ErrorInfo) { errs = append(errs, e) })
	if !called || len(errs) != 1 || errs[0].LineNumber != 1 {
		t.Errorf("Fn did not run as expected: called=%v errs=%#v", called, errs)
	}
}
