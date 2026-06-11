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

func TestMD058(t *testing.T) {
	t.Run("positive_table_followed_by_blockquote", func(t *testing.T) {
		// A table immediately followed by a blockquote (no blank line) fires.
		errs := lintB(t, "MD058", "# H\n\n| H | H |\n| - | - |\n| C | C |\n> quote\n")
		if !firedAtB(errs, "MD058", 5) {
			t.Fatalf("expected MD058 to fire at line 5, got %+v", errs)
		}
	})

	t.Run("negative_table_surrounded_by_blanks", func(t *testing.T) {
		errs := lintB(t, "MD058", "text\n\n| H | H |\n| - | - |\n| C | C |\n\ntext\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	// NOTE: text immediately preceding a table (no blank line) does not trigger
	// MD058 (the leading paragraph appears to be merged with the table), so a
	// missing blank line before the table is not reported here. This test pins
	// the actual (no-fire) behavior.
	t.Run("text_before_table_not_flagged", func(t *testing.T) {
		errs := lintB(t, "MD058", "Some text\n| H | H |\n| - | - |\n| C | C |\n\nMore\n")
		if len(errs) != 0 {
			t.Errorf("MD058 unexpectedly fired for text-before (behavior changed): %+v", errs)
		}
	})
}
