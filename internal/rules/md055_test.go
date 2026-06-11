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

func TestMD055(t *testing.T) {
	t.Run("positive_inconsistent_pipes", func(t *testing.T) {
		// Header row has leading+trailing pipes; delimiter row drops trailing.
		errs := lintB(t, "MD055", "| Header | Header |\n| ------ | ------\n| Cell   | Cell   |\n")
		if !firedAtB(errs, "MD055", 0) {
			t.Fatalf("expected MD055 to fire, got %+v", errs)
		}
	})

	t.Run("negative_consistent_pipes", func(t *testing.T) {
		errs := lintB(t, "MD055", "| Header | Header |\n| ------ | ------ |\n| Cell   | Cell   |\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("config_style_no_leading_or_trailing", func(t *testing.T) {
		errs := lintBCfg(t, "MD055", map[string]interface{}{"style": "no_leading_or_trailing"},
			"| Header | Header |\n| ------ | ------ |\n| Cell   | Cell   |\n")
		if !firedAtB(errs, "MD055", 0) {
			t.Fatalf("expected MD055 to fire for leading/trailing pipes, got %+v", errs)
		}
	})
}
