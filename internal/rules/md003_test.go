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

func TestMD003(t *testing.T) {
	t.Run("positive_consistent_mixed", func(t *testing.T) {
		// First heading is ATX, so setext is inconsistent.
		errs := lintRule(t, "MD003", "# ATX\n\nSetext\n======\n")
		if !ruleFired(errs, "MD003") {
			t.Fatalf("expected MD003 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 3 {
			t.Errorf("expected line 3, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: atx; Actual: setext" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("positive_atx_closed_mixed", func(t *testing.T) {
		errs := lintRule(t, "MD003", "# ATX heading\n\n## Closed ##\n")
		if !ruleFired(errs, "MD003") {
			t.Fatalf("expected MD003 to fire, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: atx; Actual: atx_closed" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_all_atx", func(t *testing.T) {
		errs := lintRule(t, "MD003", "# H1\n\n## H2\n\n### H3\n")
		if ruleFired(errs, "MD003") {
			t.Fatalf("expected no MD003, got %v", errs)
		}
	})

	t.Run("config_style_atx_rejects_setext", func(t *testing.T) {
		errs := lintFull(t, "Setext\n======\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD003": map[string]interface{}{"style": "atx"}}))
		if !ruleFired(errs, "MD003") {
			t.Fatalf("expected MD003 with style=atx, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: atx; Actual: setext" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("config_style_setext_accepts_setext", func(t *testing.T) {
		errs := lintFull(t, "Setext\n======\n",
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD003": map[string]interface{}{"style": "setext"}}))
		if ruleFired(errs, "MD003") {
			t.Fatalf("expected no MD003 with style=setext, got %v", errs)
		}
	})
}
