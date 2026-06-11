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

func TestMD041(t *testing.T) {
	t.Run("positive_no_top_level_heading", func(t *testing.T) {
		errs := lintB(t, "MD041", "This is a document without a heading\n")
		if !firedAtB(errs, "MD041", 1) {
			t.Fatalf("expected MD041 to fire at line 1, got %+v", errs)
		}
	})

	t.Run("negative_starts_with_h1", func(t *testing.T) {
		errs := lintB(t, "MD041", "# Document Heading\n\nText\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_allow_preamble", func(t *testing.T) {
		errs := lintBCfg(t, "MD041", map[string]interface{}{"allow_preamble": true},
			"preamble text\n\n# Heading\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations with allow_preamble, got %+v", errs)
		}
	})

	t.Run("config_level_2", func(t *testing.T) {
		// With level=2, a leading H2 is acceptable.
		errs := lintBCfg(t, "MD041", map[string]interface{}{"level": 2}, "## H2 heading\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations with level=2 and leading H2, got %+v", errs)
		}
	})

	t.Run("config_level_2_violation", func(t *testing.T) {
		// With level=2, a leading H1 should fire.
		errs := lintBCfg(t, "MD041", map[string]interface{}{"level": 2}, "# H1 heading\n")
		if !firedAtB(errs, "MD041", 1) {
			t.Fatalf("expected MD041 to fire for H1 when level=2, got %+v", errs)
		}
	})
}
