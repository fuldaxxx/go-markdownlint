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
	"testing"

	markdownlint "github.com/fuldaxxx/go-markdownlint"
)

func TestMD024(t *testing.T) {
	t.Run("positive_duplicate_heading", func(t *testing.T) {
		errs := lintRule(t, "MD024", "# Dup\n\n# Dup\n")
		if !ruleFired(errs, "MD024") {
			t.Fatalf("expected MD024 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "Dup" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("negative_unique_headings", func(t *testing.T) {
		errs := lintRule(t, "MD024", "# One\n\n# Two\n")
		if ruleFired(errs, "MD024") {
			t.Fatalf("expected no MD024, got %v", errs)
		}
	})

	t.Run("config_siblings_only_allows_different_parents", func(t *testing.T) {
		// With siblings_only, duplicate headings under different parents are OK.
		content := "# Top\n\n## A\n\n### dup\n\n## B\n\n### dup\n"
		errs := lintFull(t, content,
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD024": map[string]interface{}{"siblings_only": true}}))
		if ruleFired(errs, "MD024") {
			t.Fatalf("expected no MD024 with siblings_only and different parents, got %v", errs)
		}
	})

	t.Run("config_siblings_only_flags_same_parent", func(t *testing.T) {
		content := "# Top\n\n## A\n\n### dup\n\n### dup\n"
		errs := lintFull(t, content,
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD024": map[string]interface{}{"siblings_only": true}}))
		if !ruleFired(errs, "MD024") {
			t.Fatalf("expected MD024 with siblings_only and same parent, got %v", errs)
		}
	})
}
