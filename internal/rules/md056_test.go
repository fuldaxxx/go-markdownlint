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

func TestMD056(t *testing.T) {
	t.Run("positive_too_few_cells", func(t *testing.T) {
		errs := lintB(t, "MD056", "| H | H |\n| - | - |\n| C |\n")
		if !firedAtB(errs, "MD056", 3) {
			t.Fatalf("expected MD056 to fire at line 3, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: 2; Actual: 1; Too few cells, row will be missing data" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("positive_too_many_cells", func(t *testing.T) {
		errs := lintB(t, "MD056", "| H | H |\n| - | - |\n| C | C | C |\n")
		if !firedAtB(errs, "MD056", 3) {
			t.Fatalf("expected MD056 to fire at line 3, got %+v", errs)
		}
	})

	t.Run("negative_consistent_columns", func(t *testing.T) {
		errs := lintB(t, "MD056", "| H | H |\n| - | - |\n| C | C |\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})
}
