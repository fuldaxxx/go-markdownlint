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

import "testing"

func TestMD025(t *testing.T) {
	t.Run("positive_multiple_h1", func(t *testing.T) {
		errs := lintRule(t, "MD025", "# One\n\n# Two\n")
		if !ruleFired(errs, "MD025") {
			t.Fatalf("expected MD025 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "Two" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("negative_single_h1", func(t *testing.T) {
		errs := lintRule(t, "MD025", "# One\n\n## Two\n\n## Three\n")
		if ruleFired(errs, "MD025") {
			t.Fatalf("expected no MD025, got %v", errs)
		}
	})

	t.Run("positive_front_matter_title_plus_h1", func(t *testing.T) {
		// Front matter title plus a top-level H1 violates the single-title rule.
		errs := lintRule(t, "MD025", "---\ntitle: My Title\n---\n\n# Heading\n")
		if !ruleFired(errs, "MD025") {
			t.Fatalf("expected MD025 with front matter title, got %v", errs)
		}
		if errs[0].LineNumber != 5 {
			t.Errorf("expected line 5, got %d", errs[0].LineNumber)
		}
	})
}
