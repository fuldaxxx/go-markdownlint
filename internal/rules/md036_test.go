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
	t.Run("bold_paragraph_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "**My document**\n\nLorem ipsum dolor sit amet\n")
		if !firedAtB(errs, "MD036", 1) {
			t.Errorf("expected MD036 at line 1, got %+v", errs)
		}
	})

	t.Run("italic_paragraph_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "Text\n\n_Another section_\n\nMore text\n")
		if !firedAtB(errs, "MD036", 3) {
			t.Errorf("expected MD036 at line 3, got %+v", errs)
		}
	})

	t.Run("trailing_punctuation_not_flagged", func(t *testing.T) {
		// A single emphasized line ending with punctuation reads as a sentence,
		// not a heading, so it must not be flagged.
		errs := lintB(t, "MD036", "**Is this a heading?**\n\nBody\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for punctuated emphasis, got %+v", errs)
		}
	})

	t.Run("emphasis_inside_blockquote_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "> **Quoted heading**\n>\n> text\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations inside block quote, got %+v", errs)
		}
	})

	t.Run("emphasis_inside_list_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD036", "- **Item heading**\n- next\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations inside list, got %+v", errs)
		}
	})

	t.Run("negative_real_heading", func(t *testing.T) {
		errs := lintB(t, "MD036", "# My document\n\nLorem ipsum\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
