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

func TestMD048(t *testing.T) {
	t.Run("positive_mixed_fence_styles", func(t *testing.T) {
		errs := lintB(t, "MD048", "```\nx\n```\n\n~~~\ny\n~~~\n")
		if !firedAtB(errs, "MD048", 5) {
			t.Fatalf("expected MD048 to fire at line 5, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: backtick; Actual: tilde" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_consistent_backtick", func(t *testing.T) {
		errs := lintB(t, "MD048", "```\nx\n```\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_style_tilde_violation", func(t *testing.T) {
		errs := lintBCfg(t, "MD048", map[string]interface{}{"style": "tilde"}, "```\nx\n```\n")
		if !firedAtB(errs, "MD048", 0) {
			t.Fatalf("expected MD048 to fire for backtick when style=tilde, got %+v", errs)
		}
	})
}
