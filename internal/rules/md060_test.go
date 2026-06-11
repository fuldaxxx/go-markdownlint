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

func TestMD060(t *testing.T) {
	t.Run("positive_aligned_style_misaligned", func(t *testing.T) {
		// Pipes do not align vertically; style=aligned fires.
		errs := lintBCfg(t, "MD060", map[string]interface{}{"style": "aligned"},
			"| A | B |\n| - | - |\n| Y | Yes |\n")
		if !firedAtB(errs, "MD060", 0) {
			t.Fatalf("expected MD060 to fire for aligned style, got %+v", errs)
		}
	})

	t.Run("negative_default_any", func(t *testing.T) {
		errs := lintB(t, "MD060", "| A | B |\n| - | - |\n| Y | Yes |\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations with default 'any' style, got %+v", errs)
		}
	})

	t.Run("negative_aligned_style_aligned_table", func(t *testing.T) {
		errs := lintBCfg(t, "MD060", map[string]interface{}{"style": "aligned"},
			"| Character | Meaning |\n| --------- | ------- |\n| Y         | Yes     |\n| N         | No      |\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for properly aligned table, got %+v", errs)
		}
	})
}
