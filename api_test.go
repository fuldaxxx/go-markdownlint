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

package markdownlint

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// countRule counts how many findings reference the given rule name.
func countRule(errs []Error, name string) int {
	n := 0
	for _, e := range errs {
		for _, rn := range e.RuleNames {
			if rn == name {
				n++
			}
		}
	}
	return n
}

// --- Options.Files ----------------------------------------------------------

func TestLintFilesViaOptions(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.md")
	bad := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(good, []byte("# Heading\n\nText.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Trailing-space line -> MD009.
	if err := os.WriteFile(bad, []byte("# Heading\n\nText   \n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Lint(context.Background(), Options{
		Config: ConfigFromMap(map[string]interface{}{"default": false, "MD009": true}),
		Files:  []string{good, bad},
	})
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if _, ok := res[good]; !ok {
		t.Fatalf("results not keyed by good path; keys=%v", keysOf(res))
	}
	if _, ok := res[bad]; !ok {
		t.Fatalf("results not keyed by bad path; keys=%v", keysOf(res))
	}
	if got := countRule(res[good], "MD009"); got != 0 {
		t.Fatalf("good file should have no MD009, got %d", got)
	}
	if got := countRule(res[bad], "MD009"); got != 1 {
		t.Fatalf("bad file should have 1 MD009, got %d: %+v", got, res[bad])
	}
}

func keysOf(r Results) []string {
	var ks []string
	for k := range r {
		ks = append(ks, k)
	}
	return ks
}

// LintFiles convenience wrapper.
func TestLintFilesWrapper(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(f, []byte("#NoSpace\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := LintFiles(context.Background(), []string{f}, ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}))
	if err != nil {
		t.Fatalf("LintFiles error: %v", err)
	}
	if countRule(res[f], "MD018") != 1 {
		t.Fatalf("expected MD018 in %s, got %+v", f, res[f])
	}
}

func TestLintFilesMissingFile(t *testing.T) {
	_, err := LintFiles(context.Background(), []string{filepath.Join(t.TempDir(), "does-not-exist.md")},
		ConfigFromMap(map[string]interface{}{"default": false, "MD009": true}))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- Options.Strings (multiple docs) ---------------------------------------

func TestLintMultipleStrings(t *testing.T) {
	res, err := Lint(context.Background(), Options{
		Config: ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings: map[string]string{
			"a.md": "#A\n",
			"b.md": "# B\n",
			"c.md": "#C\n",
		},
	})
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("expected 3 results, got %d", len(res))
	}
	if countRule(res["a.md"], "MD018") != 1 {
		t.Fatalf("a.md should have MD018: %+v", res["a.md"])
	}
	if countRule(res["b.md"], "MD018") != 0 {
		t.Fatalf("b.md should be clean: %+v", res["b.md"])
	}
	if countRule(res["c.md"], "MD018") != 1 {
		t.Fatalf("c.md should have MD018: %+v", res["c.md"])
	}
}

func TestLintStringWrapper(t *testing.T) {
	errs, err := LintString(context.Background(), "x.md", "#NoSpace\n",
		ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}))
	if err != nil {
		t.Fatalf("LintString error: %v", err)
	}
	if countRule(errs, "MD018") != 1 {
		t.Fatalf("expected MD018, got %+v", errs)
	}
}

// --- Options.Concurrency ----------------------------------------------------

func TestConcurrencySameResults(t *testing.T) {
	docs := map[string]string{}
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		docs[n+".md"] = "#NoSpace\n\nText   \n\n\n\nMore\n"
	}
	cfg := ConfigFromMap(map[string]interface{}{"default": true})

	run := func(conc int) Results {
		res, err := Lint(context.Background(), Options{
			Config:      cfg,
			Strings:     docs,
			Concurrency: conc,
		})
		if err != nil {
			t.Fatalf("lint error (conc=%d): %v", conc, err)
		}
		return res
	}

	r1 := run(1)
	r4 := run(4)
	if !reflect.DeepEqual(r1, r4) {
		t.Fatalf("concurrency 1 vs 4 differ:\n1=%+v\n4=%+v", r1, r4)
	}
}

// --- NoInlineConfig ---------------------------------------------------------

func TestNoInlineConfig(t *testing.T) {
	content := "#One\n\n<!-- markdownlint-disable MD018 -->\n#Two\n"
	cfg := ConfigFromMap(map[string]interface{}{"default": false, "MD018": true})

	// Inline config honored: the disable comment suppresses the 2nd violation.
	res, err := Lint(context.Background(), Options{
		Config:         cfg,
		Strings:        map[string]string{"a.md": content},
		NoInlineConfig: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRule(res["a.md"], "MD018"); got != 1 {
		t.Fatalf("NoInlineConfig=false: expected 1 MD018 (second disabled), got %d: %+v", got, res["a.md"])
	}

	// Inline config ignored: both violations reported.
	res2, err := Lint(context.Background(), Options{
		Config:         cfg,
		Strings:        map[string]string{"a.md": content},
		NoInlineConfig: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRule(res2["a.md"], "MD018"); got != 2 {
		t.Fatalf("NoInlineConfig=true: expected 2 MD018 (comment ignored), got %d: %+v", got, res2["a.md"])
	}
}

// --- Custom rules -----------------------------------------------------------

func TestCustomRulePanicHandled(t *testing.T) {
	panicRule := &Rule{
		Names:       []string{"custom-panic", "CP"},
		Description: "panic rule",
		Tags:        []string{"test"},
		Parser:      ParserNone,
		Fn: func(_ *RuleParams, _ OnError) {
			panic("boom")
		},
	}
	res, err := Lint(context.Background(), Options{
		Config:             ConfigFromMap(map[string]interface{}{"default": false, "custom-panic": true}),
		Strings:            map[string]string{"a.md": "# Title\n\nBody\n"},
		CustomRules:        []*Rule{panicRule},
		HandleRuleFailures: true,
	})
	if err != nil {
		t.Fatalf("HandleRuleFailures=true should swallow the panic, got err=%v", err)
	}
	errs := res["a.md"]
	if len(errs) != 1 {
		t.Fatalf("expected exactly one finding from the panicking rule, got %+v", errs)
	}
	e := errs[0]
	if e.LineNumber != 1 {
		t.Fatalf("panic finding should be at line 1, got %d", e.LineNumber)
	}
	if e.RuleNames[0] != "custom-panic" {
		t.Fatalf("unexpected rule names: %v", e.RuleNames)
	}
	if !strings.Contains(e.ErrorDetail, "threw an exception") {
		t.Fatalf("expected exception detail, got %q", e.ErrorDetail)
	}
}

func TestCustomRulePanicUnhandledReturnsError(t *testing.T) {
	// NOTE: with HandleRuleFailures=false a panicking rule surfaces as an error
	// from Lint.
	panicRule := &Rule{
		Names:       []string{"custom-panic2"},
		Description: "panic rule",
		Tags:        []string{"test"},
		Parser:      ParserNone,
		Fn:          func(_ *RuleParams, _ OnError) { panic("kaboom") },
	}
	_, err := Lint(context.Background(), Options{
		Config:      ConfigFromMap(map[string]interface{}{"default": false, "custom-panic2": true}),
		Strings:     map[string]string{"a.md": "# Title\n"},
		CustomRules: []*Rule{panicRule},
	})
	if err == nil {
		t.Fatal("expected an error when a rule panics and HandleRuleFailures=false")
	}
	if !strings.Contains(err.Error(), "threw an exception") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomRuleNormalError(t *testing.T) {
	normalRule := &Rule{
		Names:       []string{"my-custom-rule", "mcr"},
		Description: "My custom rule description",
		Tags:        []string{"custom"},
		Parser:      ParserNone,
		Fn: func(p *RuleParams, onError OnError) {
			for i, line := range p.Lines {
				if strings.Contains(line, "FORBIDDEN") {
					onError(ErrorInfo{
						LineNumber: i + 1,
						Detail:     "found forbidden token",
						Context:    "FORBIDDEN",
					})
				}
			}
		},
	}
	res, err := Lint(context.Background(), Options{
		Config:      ConfigFromMap(map[string]interface{}{"default": false, "my-custom-rule": true}),
		Strings:     map[string]string{"a.md": "# Title\n\nThis has FORBIDDEN content\n"},
		CustomRules: []*Rule{normalRule},
	})
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	errs := res["a.md"]
	if len(errs) != 1 {
		t.Fatalf("expected one finding, got %+v", errs)
	}
	e := errs[0]
	if e.LineNumber != 3 {
		t.Fatalf("expected finding at line 3, got %d", e.LineNumber)
	}
	if !reflect.DeepEqual(e.RuleNames, []string{"my-custom-rule", "mcr"}) {
		t.Fatalf("unexpected rule names: %v", e.RuleNames)
	}
	if e.RuleDescription != "My custom rule description" {
		t.Fatalf("unexpected description: %q", e.RuleDescription)
	}
	if e.ErrorDetail != "found forbidden token" {
		t.Fatalf("unexpected detail: %q", e.ErrorDetail)
	}
	if e.ErrorContext != "FORBIDDEN" {
		t.Fatalf("unexpected context: %q", e.ErrorContext)
	}
}

// --- Front matter -----------------------------------------------------------

func TestFrontMatterDefaultExcludesContent(t *testing.T) {
	// A heading-like line inside default YAML front matter is excluded.
	doc := "---\n#NoSpace\n---\n\nBody text.\n"
	res, err := Lint(context.Background(), Options{
		Config:  ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings: map[string]string{"a.md": doc},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRule(res["a.md"], "MD018"); got != 0 {
		t.Fatalf("default front matter should be excluded; got %d MD018: %+v", got, res["a.md"])
	}
}

func TestDisableFrontMatter(t *testing.T) {
	// With front matter disabled, the YAML block is treated as content, so the
	// "#NoSpace" line inside it triggers MD018 at line 2.
	doc := "---\n#NoSpace\n---\n\nBody text.\n"
	res, err := Lint(context.Background(), Options{
		Config:             ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings:            map[string]string{"a.md": doc},
		DisableFrontMatter: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	errs := res["a.md"]
	if countRule(errs, "MD018") != 1 {
		t.Fatalf("DisableFrontMatter: expected 1 MD018, got %+v", errs)
	}
	if errs[0].LineNumber != 2 {
		t.Fatalf("expected MD018 at line 2, got %d", errs[0].LineNumber)
	}
}

func TestCustomFrontMatterRegexp(t *testing.T) {
	// Custom front matter delimited by ;;; — the heading-like line inside is
	// excluded only when the custom regexp matches the block.
	doc := ";;;\n#NoSpace\n;;;\n\nBody text.\n"
	custom := regexp.MustCompile(`(?m)\A;;;$[\s\S]+?^;;;$`)

	res, err := Lint(context.Background(), Options{
		Config:      ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings:     map[string]string{"a.md": doc},
		FrontMatter: custom,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRule(res["a.md"], "MD018"); got != 0 {
		t.Fatalf("custom front matter should exclude the ;;; block; got %d: %+v", got, res["a.md"])
	}

	// Without the custom regexp the default does not recognise ;;;, so the
	// heading-like line is linted and MD018 fires.
	res2, err := Lint(context.Background(), Options{
		Config:  ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings: map[string]string{"a.md": doc},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRule(res2["a.md"], "MD018"); got != 1 {
		t.Fatalf("default front matter should not exclude ;;; block; got %d: %+v", got, res2["a.md"])
	}
}

// --- ApplyFix ---------------------------------------------------------------

func TestApplyFixVariants(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		fi         FixInfo
		lineEnding string
		want       string
		wantOK     bool
	}{
		{
			name:   "insert",
			line:   "abcd",
			fi:     FixInfo{EditColumn: 3, InsertText: "XY"},
			want:   "abXYcd",
			wantOK: true,
		},
		{
			name:   "delete",
			line:   "abcd",
			fi:     FixInfo{EditColumn: 2, DeleteCount: 2},
			want:   "ad",
			wantOK: true,
		},
		{
			name:   "replace",
			line:   "abcd",
			fi:     FixInfo{EditColumn: 2, DeleteCount: 2, InsertText: "ZZ"},
			want:   "aZZd",
			wantOK: true,
		},
		{
			name:   "whole-line-delete",
			line:   "abcd",
			fi:     FixInfo{DeleteCount: -1},
			want:   "",
			wantOK: false,
		},
		{
			name:       "insert-newline-uses-line-ending",
			line:       "abcd",
			fi:         FixInfo{EditColumn: 5, InsertText: "\nnew"},
			lineEnding: "\r\n",
			want:       "abcd\r\nnew",
			wantOK:     true,
		},
		{
			name:   "edit-column-past-end-clamps",
			line:   "ab",
			fi:     FixInfo{EditColumn: 99, InsertText: "X"},
			want:   "abX",
			wantOK: true,
		},
		{
			name:   "delete-count-past-end-clamps",
			line:   "ab",
			fi:     FixInfo{EditColumn: 1, DeleteCount: 99},
			want:   "",
			wantOK: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			le := tc.lineEnding
			if le == "" {
				le = "\n"
			}
			got, ok := ApplyFix(tc.line, tc.fi, le)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("ApplyFix(%q,%+v) = %q,%v; want %q,%v", tc.line, tc.fi, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// --- ApplyFixes round trip --------------------------------------------------

func TestApplyFixesRoundTrip(t *testing.T) {
	// MD009 (trailing spaces) and MD047 (trailing newline) are both fixable.
	content := "# Heading\n\nText with trailing spaces   \nAnother line"
	cfg := ConfigFromMap(map[string]interface{}{"default": false, "MD009": true, "MD047": true})

	errs, err := LintString(context.Background(), "a.md", content, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Fatal("expected fixable findings before applying fixes")
	}
	for _, e := range errs {
		if e.FixInfo == nil {
			t.Fatalf("expected FixInfo on %v", e.RuleNames)
		}
	}

	fixed := ApplyFixes(content, errs)
	if strings.Contains(fixed, "   \n") {
		t.Fatalf("trailing spaces not removed: %q", fixed)
	}
	if !strings.HasSuffix(fixed, "\n") {
		t.Fatalf("trailing newline not added: %q", fixed)
	}

	// Re-lint the fixed document: should have strictly fewer findings (ideally 0).
	errs2, err := LintString(context.Background(), "a.md", fixed, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs2) >= len(errs) {
		t.Fatalf("expected fewer findings after fix: before=%d after=%d", len(errs), len(errs2))
	}
	if len(errs2) != 0 {
		t.Fatalf("expected zero findings after fix, got %+v", errs2)
	}
}

// --- Rules() ----------------------------------------------------------------

func TestRulesMetadata(t *testing.T) {
	rs := Rules()
	if len(rs) != 53 {
		t.Fatalf("expected 53 built-in rules, got %d", len(rs))
	}

	byName := map[string]RuleInfo{}
	for _, r := range rs {
		if len(r.Names) == 0 {
			t.Fatalf("rule with no names: %+v", r)
		}
		if r.Description == "" {
			t.Fatalf("rule %v has empty description", r.Names)
		}
		byName[r.Names[0]] = r
	}

	md009, ok := byName["MD009"]
	if !ok {
		t.Fatal("MD009 not present")
	}
	if !md009.Fixable {
		t.Fatal("MD009 should be marked Fixable")
	}
	// Spot-check names/tags carry through.
	if !contains(md009.Names, "no-trailing-spaces") {
		t.Fatalf("MD009 names missing alias: %v", md009.Names)
	}
	if len(md009.Tags) == 0 {
		t.Fatalf("MD009 should have tags: %v", md009.Tags)
	}

	md001, ok := byName["MD001"]
	if !ok {
		t.Fatal("MD001 not present")
	}
	if md001.Fixable {
		t.Fatal("MD001 should not be marked Fixable (not in FixableRuleNames)")
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// --- Version ----------------------------------------------------------------

func TestVersionNonEmpty(t *testing.T) {
	if Version() == "" {
		t.Fatal("Version() should be non-empty")
	}
}

// --- ResultsString ----------------------------------------------------------

func TestResultsStringFormat(t *testing.T) {
	res, err := Lint(context.Background(), Options{
		Config:  ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings: map[string]string{"file.md": "#NoSpace\n"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := ResultsString(res)
	// Format: "file: line: RULE/alias description ...".
	if !strings.Contains(s, "file.md: 1: MD018/no-missing-space-atx No space after hash") {
		t.Fatalf("unexpected ResultsString output: %q", s)
	}
}

func TestResultsStringSortedByFile(t *testing.T) {
	res, err := Lint(context.Background(), Options{
		Config: ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}),
		Strings: map[string]string{
			"zeta.md":  "#X\n",
			"alpha.md": "#Y\n",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := ResultsString(res)
	ai := strings.Index(s, "alpha.md")
	zi := strings.Index(s, "zeta.md")
	if ai == -1 || zi == -1 || ai > zi {
		t.Fatalf("expected alpha.md before zeta.md in sorted output: %q", s)
	}
}

func TestResultsStringEmpty(t *testing.T) {
	if got := ResultsString(Results{}); got != "" {
		t.Fatalf("expected empty string for empty results, got %q", got)
	}
}
