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

func TestMD053(t *testing.T) {
	t.Run("positive_unused_definition", func(t *testing.T) {
		errs := lintB(t, "MD053", "text\n\n[unused]: https://example.com\n")
		if !firedAtB(errs, "MD053", 3) {
			t.Fatalf("expected MD053 to fire at line 3, got %+v", errs)
		}
	})

	t.Run("negative_used_definition", func(t *testing.T) {
		errs := lintB(t, "MD053", "[used]\n\n[used]: https://example.com\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_ignored_definitions", func(t *testing.T) {
		// The default ignored_definitions includes "//"; override to ignore "unused".
		errs := lintBCfg(t, "MD053", map[string]interface{}{"ignored_definitions": []interface{}{"unused"}},
			"text\n\n[unused]: https://example.com\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations when definition ignored, got %+v", errs)
		}
	})
}
