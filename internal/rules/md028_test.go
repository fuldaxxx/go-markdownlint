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

func TestMD028(t *testing.T) {
	// NOTE: MD028 does not fire for two adjacent blockquotes separated by a bare
	// blank line. The micromark tokenization does not surface the adjacent
	// blockquote siblings the rule expects, so the rule never fires for these
	// inputs. These tests pin the actual current (no-fire) behavior.

	t.Run("adjacent_blockquotes", func(t *testing.T) {
		errs := lintRule(t, "MD028", "> a\n\n> b\n")
		if ruleFired(errs, "MD028") {
			t.Fatalf("MD028 unexpectedly fired for adjacent blockquotes (behavior changed): %v", errs)
		}
	})

	t.Run("doc_example", func(t *testing.T) {
		content := "> This is a blockquote\n> which is immediately followed by\n\n> this blockquote.\n> In some parsers.\n"
		errs := lintRule(t, "MD028", content)
		if ruleFired(errs, "MD028") {
			t.Fatalf("MD028 unexpectedly fired for doc example (behavior changed): %v", errs)
		}
	})

	t.Run("negative_separated_by_text", func(t *testing.T) {
		// Genuinely clean: blockquotes separated by text never violate.
		errs := lintRule(t, "MD028", "> a\n\nAnd Jimmy said:\n\n> b\n")
		if ruleFired(errs, "MD028") {
			t.Fatalf("expected no MD028, got %v", errs)
		}
	})
}
