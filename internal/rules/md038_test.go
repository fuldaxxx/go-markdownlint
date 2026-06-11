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

func TestMD038(t *testing.T) {
	t.Run("positive_trailing_space", func(t *testing.T) {
		errs := lintB(t, "MD038", "`some text `\n")
		if !firedAtB(errs, "MD038", 1) {
			t.Fatalf("expected MD038 to fire, got %+v", errs)
		}
		if errs[0].ErrorContext != "`some text `" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("positive_leading_space", func(t *testing.T) {
		errs := lintB(t, "MD038", "` some text`\n")
		if !firedAtB(errs, "MD038", 1) {
			t.Fatalf("expected MD038 to fire, got %+v", errs)
		}
	})

	t.Run("negative_clean_code_span", func(t *testing.T) {
		errs := lintB(t, "MD038", "`some text`\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("negative_backtick_padding_allowed", func(t *testing.T) {
		errs := lintB(t, "MD038", "`` `backticks` ``\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for spec-required padding, got %+v", errs)
		}
	})
}
