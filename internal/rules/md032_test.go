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

func TestMD032(t *testing.T) {
	t.Run("positive_no_blank_before_list", func(t *testing.T) {
		errs := lintB(t, "MD032", "Some text\n* List item\n* List item\n")
		if len(errs) == 0 {
			t.Fatal("expected MD032 to fire")
		}
		if !firedAtB(errs, "MD032", 2) {
			t.Errorf("expected violation at line 2, got %+v", errs)
		}
	})

	t.Run("negative_blank_lines_around_list", func(t *testing.T) {
		errs := lintB(t, "MD032", "Some text\n\n* List item\n* List item\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("negative_list_at_document_start", func(t *testing.T) {
		errs := lintB(t, "MD032", "* List item\n* List item\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for list at start, got %+v", errs)
		}
	})
}
