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

func TestMD029(t *testing.T) {
	t.Run("positive_inconsistent_prefix", func(t *testing.T) {
		// Style inferred as "one" (1/1/...) but the third item uses 3.
		errs := lintRule(t, "MD029", "1. one\n1. two\n3. three\n")
		if !ruleFired(errs, "MD029") {
			t.Fatalf("expected MD029 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 1; Actual: 3; Style: 1/1/1" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_ordered_sequence", func(t *testing.T) {
		errs := lintRule(t, "MD029", "1. one\n2. two\n3. three\n")
		if ruleFired(errs, "MD029") {
			t.Fatalf("expected no MD029, got %v", errs)
		}
	})

	t.Run("config_style_ordered_rejects_all_ones", func(t *testing.T) {
		errs := lintFull(t, "1. a\n1. b\n1. c\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD029": map[string]interface{}{"style": "ordered"}}))
		if !ruleFired(errs, "MD029") {
			t.Fatalf("expected MD029 with style=ordered, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: 2; Actual: 1; Style: 1/2/3" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_style_one_rejects_increment", func(t *testing.T) {
		errs := lintFull(t, "1. a\n2. b\n3. c\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD029": map[string]interface{}{"style": "one"}}))
		if !ruleFired(errs, "MD029") {
			t.Fatalf("expected MD029 with style=one, got %v", errs)
		}
	})
}
