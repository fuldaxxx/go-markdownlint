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

func TestMD014(t *testing.T) {
	t.Run("positive_dollar_no_output", func(t *testing.T) {
		errs := lintRule(t, "MD014", "```\n$ ls\n$ cd foo\n```\n")
		if !ruleFired(errs, "MD014") {
			t.Fatalf("expected MD014 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected first error on line 2, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "$ ls" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("negative_dollar_with_output", func(t *testing.T) {
		// When output is shown (a line without $), the rule does not fire.
		errs := lintRule(t, "MD014", "```\n$ ls\nfile.txt\n```\n")
		if ruleFired(errs, "MD014") {
			t.Fatalf("expected no MD014, got %v", errs)
		}
	})

	t.Run("negative_no_dollar", func(t *testing.T) {
		errs := lintRule(t, "MD014", "```\nls\ncd foo\n```\n")
		if ruleFired(errs, "MD014") {
			t.Fatalf("expected no MD014, got %v", errs)
		}
	})
}
