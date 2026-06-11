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

func TestMD035(t *testing.T) {
	t.Run("positive_inconsistent_hr", func(t *testing.T) {
		errs := lintB(t, "MD035", "---\n\n***\n")
		if !firedAtB(errs, "MD035", 3) {
			t.Fatalf("expected MD035 to fire at line 3, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: ---; Actual: ***" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_consistent_hr", func(t *testing.T) {
		errs := lintB(t, "MD035", "---\n\n---\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_explicit_style", func(t *testing.T) {
		// style requires "---"; both "***" hrs should fire.
		errs := lintBCfg(t, "MD035", map[string]interface{}{"style": "---"}, "***\n\n***\n")
		if len(errs) != 2 {
			t.Errorf("expected 2 violations, got %d: %+v", len(errs), errs)
		}
	})

	t.Run("config_explicit_style_clean", func(t *testing.T) {
		errs := lintBCfg(t, "MD035", map[string]interface{}{"style": "***"}, "***\n\n***\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
