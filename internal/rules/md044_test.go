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

func TestMD044(t *testing.T) {
	t.Run("positive_wrong_capitalization", func(t *testing.T) {
		errs := lintBCfg(t, "MD044", map[string]interface{}{"names": []interface{}{"JavaScript"}}, "Use javascript here.\n")
		if !firedAtB(errs, "MD044", 1) {
			t.Fatalf("expected MD044 to fire, got %+v", errs)
		}
		if errs[0].ErrorDetail != "Expected: JavaScript; Actual: javascript" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_correct_capitalization", func(t *testing.T) {
		errs := lintBCfg(t, "MD044", map[string]interface{}{"names": []interface{}{"JavaScript"}}, "Use JavaScript here.\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_code_blocks_false", func(t *testing.T) {
		// With code_blocks=false, an occurrence inside a code span is ignored.
		errs := lintBCfg(t, "MD044", map[string]interface{}{"names": []interface{}{"JavaScript"}, "code_blocks": false},
			"Inline `javascript` code.\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations when code_blocks=false, got %+v", errs)
		}
	})
}
