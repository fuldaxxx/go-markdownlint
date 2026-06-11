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

func TestMD046(t *testing.T) {
	t.Run("positive_mixed_styles", func(t *testing.T) {
		// First block is indented, the fenced block following it violates.
		errs := lintB(t, "MD046", "    indented code\n\n```\nfenced\n```\n")
		if !firedAtB(errs, "MD046", 0) {
			t.Fatalf("expected MD046 to fire, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: indented; Actual: fenced" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_consistent_fenced", func(t *testing.T) {
		errs := lintB(t, "MD046", "```\nfenced\n```\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_style_fenced_violation", func(t *testing.T) {
		errs := lintBCfg(t, "MD046", map[string]interface{}{"style": "fenced"}, "    indented code\n")
		if !firedAtB(errs, "MD046", 0) {
			t.Fatalf("expected MD046 to fire for indented when style=fenced, got %+v", errs)
		}
	})

	t.Run("config_style_indented_clean", func(t *testing.T) {
		errs := lintBCfg(t, "MD046", map[string]interface{}{"style": "indented"}, "    indented code\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
