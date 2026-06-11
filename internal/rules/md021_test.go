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

func TestMD021(t *testing.T) {
	t.Run("positive_multiple_space_closed_atx", func(t *testing.T) {
		errs := lintRule(t, "MD021", "#  Heading  #\n")
		if !ruleFired(errs, "MD021") {
			t.Fatalf("expected MD021 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "#  Heading  #" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
		// Two violations: leading and trailing multiple spaces.
		if len(errs) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errs))
		}
	})

	t.Run("negative_single_space_closed_atx", func(t *testing.T) {
		errs := lintRule(t, "MD021", "# Heading #\n")
		if ruleFired(errs, "MD021") {
			t.Fatalf("expected no MD021, got %v", errs)
		}
	})
}
