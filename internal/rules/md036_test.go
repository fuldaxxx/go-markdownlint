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

func TestMD036(t *testing.T) {
	// NOTE: MD036 (emphasis used instead of a heading) is not implemented:
	// single-line paragraphs that consist entirely of emphasized text (e.g.
	// "**My document**") are not flagged. These tests pin the actual (no-fire)
	// behavior; if MD036 is implemented, they will start failing and should be
	// updated to positive assertions.
	t.Run("bold_paragraph_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "**My document**\n\nLorem ipsum dolor sit amet\n")
		if len(errs) != 0 {
			t.Errorf("MD036 unexpectedly fired (behavior changed): %+v", errs)
		}
	})

	t.Run("italic_paragraph_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "Text\n\n_Another section_\n\nMore text\n")
		if len(errs) != 0 {
			t.Errorf("MD036 unexpectedly fired (behavior changed): %+v", errs)
		}
	})

	t.Run("negative_real_heading", func(t *testing.T) {
		errs := lintB(t, "MD036", "# My document\n\nLorem ipsum\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
