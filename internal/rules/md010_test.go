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

func TestMD010(t *testing.T) {
	t.Run("positive_hard_tab", func(t *testing.T) {
		errs := lintRule(t, "MD010", "text\twith tab\n")
		if !ruleFired(errs, "MD010") {
			t.Fatalf("expected MD010 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Column: 5" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
		if errs[0].ErrorRange == nil || (*errs[0].ErrorRange)[0] != 5 {
			t.Errorf("unexpected range %v", errs[0].ErrorRange)
		}
	})

	t.Run("negative_spaces_only", func(t *testing.T) {
		errs := lintRule(t, "MD010", "text with spaces\n")
		if ruleFired(errs, "MD010") {
			t.Fatalf("expected no MD010, got %v", errs)
		}
	})
}
