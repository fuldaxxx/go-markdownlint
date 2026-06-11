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

func TestMD050(t *testing.T) {
	t.Run("positive_inconsistent_bold", func(t *testing.T) {
		errs := lintB(t, "MD050", "**bold** and __bold__\n")
		if !firedAtB(errs, "MD050", 1) {
			t.Fatalf("expected MD050 to fire, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: asterisk; Actual: underscore" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_consistent_bold", func(t *testing.T) {
		errs := lintB(t, "MD050", "**bold** and **bold**\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_style_underscore_violation", func(t *testing.T) {
		errs := lintBCfg(t, "MD050", map[string]interface{}{"style": "underscore"}, "**bold**\n")
		if !firedAtB(errs, "MD050", 1) {
			t.Fatalf("expected MD050 to fire for asterisk when style=underscore, got %+v", errs)
		}
	})
}
