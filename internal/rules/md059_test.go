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

func TestMD059(t *testing.T) {
	t.Run("positive_click_here", func(t *testing.T) {
		errs := lintB(t, "MD059", "[click here](https://example.com)\n")
		if !firedAtB(errs, "MD059", 1) {
			t.Fatalf("expected MD059 to fire, got %+v", errs)
		}
	})

	t.Run("negative_descriptive_text", func(t *testing.T) {
		errs := lintB(t, "MD059", "[Download the budget document](https://example.com)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_custom_prohibited_texts", func(t *testing.T) {
		// "here" is no longer prohibited; "verboten" is.
		errs := lintBCfg(t, "MD059", map[string]interface{}{"prohibited_texts": []interface{}{"verboten"}},
			"[verboten](https://example.com)\n")
		if !firedAtB(errs, "MD059", 1) {
			t.Fatalf("expected MD059 to fire for custom prohibited text, got %+v", errs)
		}
	})
}
