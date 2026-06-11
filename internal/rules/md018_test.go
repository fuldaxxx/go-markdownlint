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

func TestMD018(t *testing.T) {
	t.Run("positive_no_space_atx", func(t *testing.T) {
		errs := lintRule(t, "MD018", "#Heading\n")
		if !ruleFired(errs, "MD018") {
			t.Fatalf("expected MD018 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorContext != "#Heading" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
	})

	t.Run("negative_proper_atx", func(t *testing.T) {
		errs := lintRule(t, "MD018", "# Heading\n")
		if ruleFired(errs, "MD018") {
			t.Fatalf("expected no MD018, got %v", errs)
		}
	})
}
