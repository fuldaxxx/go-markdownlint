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

package rules_test

import (
	"context"
	"testing"

	markdownlint "github.com/ldmonster/go-markdownlint"
)

// lintRule lints content with only the named rule enabled.
func lintRule(t *testing.T, rule, content string) []markdownlint.Error {
	t.Helper()
	errs, err := markdownlint.LintString(context.Background(), "t.md", content,
		markdownlint.ConfigFromMap(map[string]interface{}{"default": false, rule: true}))
	if err != nil {
		t.Fatalf("LintString error: %v", err)
	}
	return errs
}

// lintFull lints content with a complete custom configuration.
func lintFull(t *testing.T, content string, cfg markdownlint.Configuration) []markdownlint.Error {
	t.Helper()
	errs, err := markdownlint.LintString(context.Background(), "t.md", content, cfg)
	if err != nil {
		t.Fatalf("LintString error: %v", err)
	}
	return errs
}

// ruleFired reports whether any error names the given rule.
func ruleFired(errs []markdownlint.Error, rule string) bool {
	for _, e := range errs {
		for _, n := range e.RuleNames {
			if n == rule {
				return true
			}
		}
	}
	return false
}

func TestMD001(t *testing.T) {
	t.Run("positive_skips_level", func(t *testing.T) {
		errs := lintRule(t, "MD001", "# H1\n\n### H3 skips\n")
		if !ruleFired(errs, "MD001") {
			t.Fatalf("expected MD001 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: h2; Actual: h3" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_clean_increment", func(t *testing.T) {
		errs := lintRule(t, "MD001", "# H1\n\n## H2\n\n### H3\n")
		if ruleFired(errs, "MD001") {
			t.Fatalf("expected no MD001, got %v", errs)
		}
	})
}
