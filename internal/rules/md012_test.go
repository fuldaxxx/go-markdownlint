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

import (
	"testing"

	markdownlint "github.com/ldmonster/go-markdownlint"
)

func TestMD012(t *testing.T) {
	t.Run("positive_multiple_blanks", func(t *testing.T) {
		errs := lintRule(t, "MD012", "a\n\n\nb\n")
		if !ruleFired(errs, "MD012") {
			t.Fatalf("expected MD012 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 1; Actual: 2" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_single_blank", func(t *testing.T) {
		errs := lintRule(t, "MD012", "a\n\nb\n")
		if ruleFired(errs, "MD012") {
			t.Fatalf("expected no MD012, got %v", errs)
		}
	})

	t.Run("config_maximum_2", func(t *testing.T) {
		// maximum=2 allows two blank lines, three violate.
		errs := lintFull(t, "a\n\n\n\nb\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD012": map[string]interface{}{"maximum": 2}}))
		if !ruleFired(errs, "MD012") {
			t.Fatalf("expected MD012 with maximum=2, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: 2; Actual: 3" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_maximum_2_clean", func(t *testing.T) {
		errs := lintFull(t, "a\n\n\nb\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD012": map[string]interface{}{"maximum": 2}}))
		if ruleFired(errs, "MD012") {
			t.Fatalf("expected no MD012 with maximum=2 and 2 blanks, got %v", errs)
		}
	})
}
