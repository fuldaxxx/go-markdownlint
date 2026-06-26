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
		// Second item has 1 space of indent while the first has 0; both are
		// siblings in the same (loose) list, so MD005 fires.
		errs := lintRule(t, "MD005", "* Alpha\n * Bravo\n")
		if !ruleFired(errs, "MD005") {
			t.Fatalf("expected MD005 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected line 2, got %d", errs[0].LineNumber)
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

	t.Run("negative_indented_top_level_list", func(t *testing.T) {
		// A list whose items share the same (non-zero) leading indentation must
		// not fire MD005. Previously the listItemPrefix token reported its
		// startColumn as the line column (before the indentation) while the
		// list token reported the column after the indentation, producing false
		// "Expected: N; Actual: 0" errors. Upstream micromark sets
		// listItemPrefix.startColumn to the marker column (after indentation).
		errs := lintRule(t, "MD005", "  - `type`\n  - `address`\n")
		if ruleFired(errs, "MD005") {
			t.Fatalf("expected no MD005, got %v", errs)
		}
	})

	t.Run("negative_sublist_after_lazy_continuation", func(t *testing.T) {
		// Real-world deckhouse docs pattern: a list item whose paragraph wraps
		// onto un-indented continuation lines followed by an indented sublist.
		// The sublist items are consistently indented and must not fire MD005.
		errs := lintRule(t, "MD005",
			"* `addresses` - is a `MachineAddresses` which represents host names,\n"+
				"external DNS names, and/or internal DNS names. `MachineAddress` is\n"+
				"defined as:\n"+
				"  - `type` (string): one of `Hostname`, `ExternalIP`\n"+
				"  - `address` (string)\n")
		if ruleFired(errs, "MD005") {
			t.Fatalf("expected no MD005, got %v", errs)
		}
	})
}
