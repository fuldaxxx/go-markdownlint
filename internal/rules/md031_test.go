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

func TestMD031(t *testing.T) {
	t.Run("positive_no_blank_above_fence", func(t *testing.T) {
		errs := lintRule(t, "MD031", "text\n```\ncode\n```\n")
		if !ruleFired(errs, "MD031") {
			t.Fatalf("expected MD031 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 2 {
			t.Errorf("expected line 2, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "```" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("negative_blanks_around_fence", func(t *testing.T) {
		errs := lintRule(t, "MD031", "text\n\n```\ncode\n```\n\nmore\n")
		if ruleFired(errs, "MD031") {
			t.Fatalf("expected no MD031, got %v", errs)
		}
	})
}
