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

func TestMD030(t *testing.T) {
	t.Run("positive_multiple_space_after_marker", func(t *testing.T) {
		errs := lintRule(t, "MD030", "*  item\n")
		if !ruleFired(errs, "MD030") {
			t.Fatalf("expected MD030 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 1; Actual: 2" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_single_space", func(t *testing.T) {
		errs := lintRule(t, "MD030", "* item\n* item2\n")
		if ruleFired(errs, "MD030") {
			t.Fatalf("expected no MD030, got %v", errs)
		}
	})

	t.Run("config_ul_single_2_allows_two_spaces", func(t *testing.T) {
		errs := lintFull(t, "*  item\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD030": map[string]interface{}{"ul_single": 2}}))
		if ruleFired(errs, "MD030") {
			t.Fatalf("expected no MD030 with ul_single=2, got %v", errs)
		}
	})
}
