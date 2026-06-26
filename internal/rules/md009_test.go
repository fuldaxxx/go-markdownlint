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

	markdownlint "github.com/fuldaxxx/go-markdownlint"
)

func TestMD009(t *testing.T) {
	t.Run("positive_single_trailing_space", func(t *testing.T) {
		// Line 2 has one trailing space (1 is neither 0 nor a valid 2-space break).
		errs := lintRule(t, "MD009", "text  \nmore \n")
		if !ruleFired(errs, "MD009") {
			t.Fatalf("expected MD009 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected line 2, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 0 or 2; Actual: 1" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
		if errs[0].ErrorRange == nil || (*errs[0].ErrorRange)[0] != 5 {
			t.Errorf("unexpected range %v", errs[0].ErrorRange)
		}
	})

	t.Run("negative_no_trailing_space", func(t *testing.T) {
		errs := lintRule(t, "MD009", "text\nmore\n")
		if ruleFired(errs, "MD009") {
			t.Fatalf("expected no MD009, got %v", errs)
		}
	})

	t.Run("config_br_spaces_0_flags_two_spaces", func(t *testing.T) {
		// With br_spaces < 2, even 2 trailing spaces are not allowed.
		errs := lintFull(t, "text  \n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD009": map[string]interface{}{"br_spaces": 0}}))
		if !ruleFired(errs, "MD009") {
			t.Fatalf("expected MD009 with br_spaces=0, got %v", errs)
		}
	})
}
