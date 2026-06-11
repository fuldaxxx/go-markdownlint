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
	t.Run("positive_inconsistent_indent", func(t *testing.T) {
		// Third item has 3 spaces of indent while siblings have 0.
		errs := lintRule(t, "MD005", "* a\n* b\n   * c\n")
		if !ruleFired(errs, "MD005") {
			t.Fatalf("expected MD005 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
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
