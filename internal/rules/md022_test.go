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

func TestMD022(t *testing.T) {
	t.Run("positive_no_blank_below", func(t *testing.T) {
		errs := lintRule(t, "MD022", "# H1\ntext\n")
		if !ruleFired(errs, "MD022") {
			t.Fatalf("expected MD022 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 1; Actual: 0; Below" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})

	t.Run("negative_blanks_around", func(t *testing.T) {
		errs := lintRule(t, "MD022", "# H1\n\ntext\n\n## H2\n\nmore\n")
		if ruleFired(errs, "MD022") {
			t.Fatalf("expected no MD022, got %v", errs)
		}
	})
}
