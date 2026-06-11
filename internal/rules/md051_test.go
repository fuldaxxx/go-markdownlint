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

func TestMD051(t *testing.T) {
	// NOTE: MD051 (link fragments should be valid) is not implemented: a link
	// fragment (e.g. [x](#nonexistent)) that does not correspond to any heading
	// anchor or named anchor in the document is not flagged. These tests pin the
	// actual (no-fire) behavior; if MD051 is implemented, update to positive
	// assertions.
	t.Run("invalid_fragment_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD051", "# My Heading\n\n[link](#nonexistent)\n")
		if len(errs) != 0 {
			t.Errorf("MD051 unexpectedly fired (behavior changed): %+v", errs)
		}
	})

	t.Run("negative_valid_fragment", func(t *testing.T) {
		// A fragment matching the heading anchor must not fire.
		errs := lintB(t, "MD051", "# My Heading\n\n[link](#my-heading)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for valid fragment, got %+v", errs)
		}
	})
}
