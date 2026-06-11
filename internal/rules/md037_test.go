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

func TestMD037(t *testing.T) {
	t.Run("positive_spaces_in_bold", func(t *testing.T) {
		errs := lintB(t, "MD037", "Here is some ** bold ** text.\n")
		if !firedAtB(errs, "MD037", 1) {
			t.Fatalf("expected MD037 to fire, got %+v", errs)
		}
	})

	t.Run("positive_spaces_in_italic", func(t *testing.T) {
		errs := lintB(t, "MD037", "Here is some * italic * text.\n")
		if !firedAtB(errs, "MD037", 1) {
			t.Fatalf("expected MD037 to fire, got %+v", errs)
		}
	})

	t.Run("negative_clean_emphasis", func(t *testing.T) {
		errs := lintB(t, "MD037", "Here is some **bold** and *italic* text.\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
