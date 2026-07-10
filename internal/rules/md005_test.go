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

func TestMD005(t *testing.T) {
	t.Run("positive_ordered_misindent", func(t *testing.T) {
		// The middle ordered item is indented one space past its siblings.
		errs := lintRule(t, "MD005", "1. one\n 1. two\n1. three\n")
		if !ruleFired(errs, "MD005") {
			t.Fatalf("expected MD005 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected line 2, got %d", errs[0].LineNumber)
		}
	})

	t.Run("negative_nested_under_ordered", func(t *testing.T) {
		// Regression: bullets aligned under an ordered item's text share a
		// consistent indent and must not be flagged. This nested-list pattern
		// previously produced a flood of false positives because the rule
		// measured indentation from the list container instead of the marker.
		errs := lintRule(t, "MD005", "1. one\n   * a\n   * b\n1. two\n")
		if ruleFired(errs, "MD005") {
			t.Fatalf("expected no MD005 for aligned nested bullets, got %v", errs)
		}
	})

	t.Run("negative_consistent_indent", func(t *testing.T) {
		errs := lintRule(t, "MD005", "* a\n* b\n* c\n")
		if ruleFired(errs, "MD005") {
			t.Fatalf("expected no MD005, got %v", errs)
		}
	})

	t.Run("negative_consistent_nested", func(t *testing.T) {
		errs := lintRule(t, "MD005", "* a\n  * b\n  * c\n* d\n")
		if ruleFired(errs, "MD005") {
			t.Fatalf("expected no MD005, got %v", errs)
		}
	})
}
