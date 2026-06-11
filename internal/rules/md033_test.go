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

func TestMD033(t *testing.T) {
	t.Run("positive_inline_html", func(t *testing.T) {
		errs := lintB(t, "MD033", "text with <span>html</span> here\n")
		if !firedAtB(errs, "MD033", 1) {
			t.Fatalf("expected MD033 to fire at line 1, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Element: span" {
			t.Errorf("expected detail 'Element: span', got %q", errs[0].ErrorDetail)
		}
	})

	t.Run("positive_br_tag", func(t *testing.T) {
		errs := lintB(t, "MD033", "A line<br>break\n")
		if !firedAtB(errs, "MD033", 1) {
			t.Fatalf("expected MD033 to fire for <br>, got %+v", errs)
		}
	})

	// NOTE: MD033 does not flag block-level HTML such as <div>...</div> or
	// <img ...> on their own line; this test pins that no-fire behavior. Update
	// it if block-level HTML detection is added.
	t.Run("block_level_html_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD033", "<div>block</div>\n")
		if len(errs) != 0 {
			t.Errorf("block-level HTML unexpectedly fired (behavior changed): %+v", errs)
		}
	})

	t.Run("negative_pure_markdown", func(t *testing.T) {
		errs := lintB(t, "MD033", "# Markdown heading\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_allowed_elements", func(t *testing.T) {
		errs := lintBCfg(t, "MD033", map[string]interface{}{"allowed_elements": []interface{}{"span"}},
			"text with <span>html</span> here\n")
		if len(errs) != 0 {
			t.Errorf("expected span to be allowed, got %+v", errs)
		}
	})
}
