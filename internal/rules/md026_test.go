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

func TestMD026(t *testing.T) {
	t.Run("positive_trailing_period", func(t *testing.T) {
		errs := lintRule(t, "MD026", "# Heading.\n")
		if !ruleFired(errs, "MD026") {
			t.Fatalf("expected MD026 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Punctuation: '.'" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_no_trailing_punctuation", func(t *testing.T) {
		errs := lintRule(t, "MD026", "# Heading\n")
		if ruleFired(errs, "MD026") {
			t.Fatalf("expected no MD026, got %v", errs)
		}
	})

	t.Run("config_custom_punctuation_semicolon", func(t *testing.T) {
		errs := lintFull(t, "# Heading;\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD026": map[string]interface{}{"punctuation": ".,;:!?"}}))
		if !ruleFired(errs, "MD026") {
			t.Fatalf("expected MD026 with custom punctuation, got %v", errs)
		}
		if errs[0].ErrorDetail != "Punctuation: ';'" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_empty_punctuation_disables", func(t *testing.T) {
		errs := lintFull(t, "# Heading.\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD026": map[string]interface{}{"punctuation": ""}}))
		if ruleFired(errs, "MD026") {
			t.Fatalf("expected no MD026 with empty punctuation set, got %v", errs)
		}
	})
}
