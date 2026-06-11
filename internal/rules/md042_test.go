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

func TestMD042(t *testing.T) {
	t.Run("positive_empty_link", func(t *testing.T) {
		errs := lintB(t, "MD042", "[an empty link]()\n")
		if !firedAtB(errs, "MD042", 1) {
			t.Fatalf("expected MD042 to fire, got %+v", errs)
		}
	})

	t.Run("positive_empty_fragment", func(t *testing.T) {
		errs := lintB(t, "MD042", "[an empty fragment](#)\n")
		if !firedAtB(errs, "MD042", 1) {
			t.Fatalf("expected MD042 to fire for #, got %+v", errs)
		}
	})

	t.Run("negative_valid_link", func(t *testing.T) {
		errs := lintB(t, "MD042", "[a valid link](https://example.com/)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("negative_non_empty_fragment", func(t *testing.T) {
		errs := lintB(t, "MD042", "[a valid fragment](#fragment)\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for #fragment, got %+v", errs)
		}
	})
}
