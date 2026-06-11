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

func TestMD043(t *testing.T) {
	t.Run("positive_wrong_heading", func(t *testing.T) {
		errs := lintBCfg(t, "MD043", map[string]interface{}{"headings": []interface{}{"# Heading"}}, "# Wrong\n")
		if !firedAtB(errs, "MD043", 1) {
			t.Fatalf("expected MD043 to fire, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: # Heading; Actual: # Wrong" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_matching_heading", func(t *testing.T) {
		errs := lintBCfg(t, "MD043", map[string]interface{}{"headings": []interface{}{"# Heading"}}, "# Heading\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_match_case_default_insensitive", func(t *testing.T) {
		// Default match_case=false: differing case is OK.
		errs := lintBCfg(t, "MD043", map[string]interface{}{"headings": []interface{}{"# Heading"}}, "# heading\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations (case-insensitive default), got %+v", errs)
		}
	})

	t.Run("config_match_case_true", func(t *testing.T) {
		errs := lintBCfg(t, "MD043", map[string]interface{}{"headings": []interface{}{"# Heading"}, "match_case": true}, "# heading\n")
		if !firedAtB(errs, "MD043", 1) {
			t.Fatalf("expected MD043 to fire with match_case=true, got %+v", errs)
		}
	})
}
