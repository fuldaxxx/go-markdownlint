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
	"strings"
	"testing"
)

func lintOne(t *testing.T, content string, cfg Configuration) []Error {
	t.Helper()
	errs, err := LintString(context.Background(), "test.md", content, cfg)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	return errs
}

func hasRule(errs []Error, name string) bool {
	for _, e := range errs {
		for _, n := range e.RuleNames {
			if n == name {
				return true
			}
		}
	}
	return false
}

func TestMD001HeadingIncrement(t *testing.T) {
	errs := lintOne(t, "# H1\n\n### H3 skips H2\n", ConfigFromMap(map[string]interface{}{"default": false, "MD001": true}))
	if !hasRule(errs, "MD001") {
		t.Fatalf("expected MD001, got %+v", errs)
	}
}

func TestMD001NoViolation(t *testing.T) {
	errs := lintOne(t, "# H1\n\n## H2\n\n### H3\n", ConfigFromMap(map[string]interface{}{"default": false, "MD001": true}))
	if hasRule(errs, "MD001") {
		t.Fatalf("unexpected MD001: %+v", errs)
	}
}

func TestMD047TrailingNewline(t *testing.T) {
	errs := lintOne(t, "# Heading\n\nText with no trailing newline", ConfigFromMap(map[string]interface{}{"default": false, "MD047": true}))
	if !hasRule(errs, "MD047") {
		t.Fatalf("expected MD047, got %+v", errs)
	}
}

func TestMD009TrailingSpaces(t *testing.T) {
	errs := lintOne(t, "# Heading\n\nText with trailing space   \n", ConfigFromMap(map[string]interface{}{"default": false, "MD009": true}))
	if !hasRule(errs, "MD009") {
		t.Fatalf("expected MD009, got %+v", errs)
	}
}

func TestMD012MultipleBlanks(t *testing.T) {
	errs := lintOne(t, "# H\n\n\n\nText\n", ConfigFromMap(map[string]interface{}{"default": false, "MD012": true}))
	if !hasRule(errs, "MD012") {
		t.Fatalf("expected MD012, got %+v", errs)
	}
}

func TestMD018NoSpaceAtx(t *testing.T) {
	errs := lintOne(t, "#Heading\n", ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}))
	if !hasRule(errs, "MD018") {
		t.Fatalf("expected MD018, got %+v", errs)
	}
}

func TestInlineDisable(t *testing.T) {
	content := "#Heading\n\n<!-- markdownlint-disable MD018 -->\n#Another\n"
	errs := lintOne(t, content, ConfigFromMap(map[string]interface{}{"default": false, "MD018": true}))
	count := 0
	for _, e := range errs {
		if e.RuleNames[0] == "MD018" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 MD018 (second disabled), got %d: %+v", count, errs)
	}
}

func TestApplyFixesTrailingNewline(t *testing.T) {
	content := "# Heading\n\nText"
	errs := lintOne(t, content, ConfigFromMap(map[string]interface{}{"default": false, "MD047": true}))
	fixed := ApplyFixes(content, errs)
	if !strings.HasSuffix(fixed, "Text\n") {
		t.Fatalf("expected trailing newline added, got %q", fixed)
	}
}

func TestConfigDefaultTrue(t *testing.T) {
	errs := lintOne(t, "#Heading\n", Configuration{})
	if len(errs) == 0 {
		t.Fatalf("expected default rules to produce errors")
	}
}

func TestResultsString(t *testing.T) {
	res, err := Lint(context.Background(), Options{
		Config:  ConfigFromMap(map[string]interface{}{"default": false, "MD047": true}),
		Strings: map[string]string{"a.md": "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := ResultsString(res)
	if !strings.Contains(s, "MD047") {
		t.Fatalf("expected MD047 in output: %q", s)
	}
}
