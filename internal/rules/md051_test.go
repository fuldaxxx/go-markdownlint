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
	t.Run("invalid_fragment_flagged", func(t *testing.T) {
		errs := lintB(t, "MD051", "# My Heading\n\n[link](#nonexistent)\n")
		if !firedAtB(errs, "MD051", 3) {
			t.Errorf("expected MD051 at line 3, got %+v", errs)
		}
	})

	t.Run("negative_valid_fragment", func(t *testing.T) {
		// A fragment matching the heading anchor must not fire.
		errs := lintB(t, "MD051", "# My Heading\n\n[link](#my-heading)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for valid fragment, got %+v", errs)
		}
	})

	t.Run("negative_top_fragment", func(t *testing.T) {
		// "#top" is always a valid fragment.
		errs := lintB(t, "MD051", "# My Heading\n\n[back to top](#top)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for #top, got %+v", errs)
		}
	})

	t.Run("negative_html_flow_id_anchor", func(t *testing.T) {
		// A block-level `<a id="...">` on its own line lands in htmlFlow (this
		// port classifies a leading "<" as an HTML block). Its id anchor must
		// still count as a valid fragment target, matching upstream markdownlint.
		errs := lintB(t, "MD051", "# Title\n\n[go to rule](#rule-1)\n\n<a id=\"rule-1\"></a>\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for htmlFlow id anchor, got %+v", errs)
		}
	})

	t.Run("negative_html_flow_name_anchor", func(t *testing.T) {
		// The `name` attribute is honored only on <a> tags, matching upstream.
		errs := lintB(t, "MD051", "# Title\n\n[jump](#sec)\n\n<a name=\"sec\"></a>\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for htmlFlow name anchor, got %+v", errs)
		}
	})
}
