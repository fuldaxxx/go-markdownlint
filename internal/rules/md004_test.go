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

func TestMD004(t *testing.T) {
	t.Run("positive_consistent_mixed", func(t *testing.T) {
		// First marker is asterisk; plus and dash are inconsistent.
		errs := lintRule(t, "MD004", "* a\n+ b\n- c\n")
		if !ruleFired(errs, "MD004") {
			t.Fatalf("expected MD004 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected first error on line 2, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: asterisk; Actual: plus" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_all_asterisk", func(t *testing.T) {
		errs := lintRule(t, "MD004", "* a\n* b\n* c\n")
		if ruleFired(errs, "MD004") {
			t.Fatalf("expected no MD004, got %v", errs)
		}
	})

	t.Run("config_style_dash_rejects_asterisk", func(t *testing.T) {
		errs := lintFull(t, "* a\n* b\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD004": map[string]interface{}{"style": "dash"}}))
		if !ruleFired(errs, "MD004") {
			t.Fatalf("expected MD004 with style=dash, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: dash; Actual: asterisk" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_style_sublist_alternation_clean", func(t *testing.T) {
		errs := lintFull(t, "* a\n+ b\n- c\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD004": map[string]interface{}{"style": "sublist"}}))
		// All at the same level with the "sublist" style; the first-seen marker
		// for a level governs, so differing same-level markers still violate.
		if !ruleFired(errs, "MD004") {
			t.Fatalf("expected MD004 with style=sublist, got %v", errs)
		}
	})
}
