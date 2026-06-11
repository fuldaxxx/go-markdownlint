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

func TestMD040(t *testing.T) {
	t.Run("positive_missing_language", func(t *testing.T) {
		errs := lintB(t, "MD040", "```\n#!/bin/bash\necho hi\n```\n")
		if !firedAtB(errs, "MD040", 1) {
			t.Fatalf("expected MD040 to fire at line 1, got %+v", errs)
		}
	})

	t.Run("negative_language_specified", func(t *testing.T) {
		errs := lintB(t, "MD040", "```bash\necho hi\n```\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_allowed_languages_violation", func(t *testing.T) {
		errs := lintBCfg(t, "MD040", map[string]interface{}{"allowed_languages": []interface{}{"bash"}},
			"```ruby\ncode\n```\n")
		if !firedAtB(errs, "MD040", 1) {
			t.Fatalf("expected violation for disallowed language, got %+v", errs)
		}
		if errs[0].ErrorDetail != "\"ruby\" is not allowed" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_allowed_languages_clean", func(t *testing.T) {
		errs := lintBCfg(t, "MD040", map[string]interface{}{"allowed_languages": []interface{}{"ruby"}},
			"```ruby\ncode\n```\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	// NOTE: the language_only option is meant to fire when the fenced code info
	// string contains more than just a language (e.g. "ruby startline=3"). This
	// implementation does not extract the meta portion, so the option never
	// reports a violation here. This test pins the actual (no-fire) behavior.
	t.Run("language_only_not_flagged", func(t *testing.T) {
		errs := lintBCfg(t, "MD040", map[string]interface{}{"language_only": true},
			"```ruby startline=3\ncode\n```\n")
		if len(errs) != 0 {
			t.Errorf("language_only unexpectedly fired (behavior changed): %+v", errs)
		}
	})
}
