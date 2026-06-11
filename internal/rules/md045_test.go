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

func TestMD045(t *testing.T) {
	t.Run("positive_missing_alt_text", func(t *testing.T) {
		errs := lintB(t, "MD045", "![](image.jpg)\n")
		if !firedAtB(errs, "MD045", 1) {
			t.Fatalf("expected MD045 to fire, got %+v", errs)
		}
	})

	t.Run("negative_has_alt_text", func(t *testing.T) {
		errs := lintB(t, "MD045", "![Alternate text](image.jpg)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("negative_html_with_alt", func(t *testing.T) {
		errs := lintB(t, "MD045", "<img src=\"image.jpg\" alt=\"Alternate text\" />\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for HTML img with alt, got %+v", errs)
		}
	})
}
