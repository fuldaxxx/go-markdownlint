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

func TestMD049(t *testing.T) {
	t.Run("positive_inconsistent_italic", func(t *testing.T) {
		// First emphasis uses asterisk, the underscore one violates (consistent).
		errs := lintB(t, "MD049", "*italic* and _italic_\n")
		if !firedAtB(errs, "MD049", 1) {
			t.Fatalf("expected MD049 to fire, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: asterisk; Actual: underscore" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_consistent_italic", func(t *testing.T) {
		errs := lintB(t, "MD049", "*italic* and *italic*\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_style_underscore_violation", func(t *testing.T) {
		errs := lintBCfg(t, "MD049", map[string]interface{}{"style": "underscore"}, "*italic*\n")
		if !firedAtB(errs, "MD049", 1) {
			t.Fatalf("expected MD049 to fire for asterisk when style=underscore, got %+v", errs)
		}
	})
}
