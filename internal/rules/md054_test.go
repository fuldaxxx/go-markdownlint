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

func TestMD054(t *testing.T) {
	t.Run("positive_inline_disabled", func(t *testing.T) {
		errs := lintBCfg(t, "MD054", map[string]interface{}{"inline": false}, "[link](https://example.com)\n")
		if !firedAtB(errs, "MD054", 1) {
			t.Fatalf("expected MD054 to fire when inline=false, got %+v", errs)
		}
	})

	t.Run("negative_default_allows_all", func(t *testing.T) {
		errs := lintB(t, "MD054", "[link](https://example.com)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations with default config, got %+v", errs)
		}
	})

	t.Run("positive_autolink_disabled", func(t *testing.T) {
		// NOTE: a standalone "<https://example.com>" line is parsed as an HTML
		// block (not an autolink) by this port, so the autolink must be inline
		// within surrounding text for MD054 to recognize it.
		errs := lintBCfg(t, "MD054", map[string]interface{}{"autolink": false}, "See <https://example.com> here.\n")
		if !firedAtB(errs, "MD054", 1) {
			t.Fatalf("expected MD054 to fire when autolink=false, got %+v", errs)
		}
	})
}
