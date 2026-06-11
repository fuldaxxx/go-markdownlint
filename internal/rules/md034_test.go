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

func TestMD034(t *testing.T) {
	t.Run("positive_bare_url", func(t *testing.T) {
		errs := lintB(t, "MD034", "Visit https://www.example.com/ now.\n")
		if !firedAtB(errs, "MD034", 1) {
			t.Fatalf("expected MD034 to fire, got %+v", errs)
		}
		if errs[0].ErrorContext != "https://www.example.com/" {
			t.Errorf("unexpected context %q", errs[0].ErrorContext)
		}
		if errs[0].ErrorRange == nil {
			t.Errorf("expected an errorRange")
		}
	})

	t.Run("negative_angle_bracketed", func(t *testing.T) {
		errs := lintB(t, "MD034", "Visit <https://www.example.com/> now.\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations, got %+v", errs)
		}
	})

	t.Run("negative_code_span", func(t *testing.T) {
		errs := lintB(t, "MD034", "Not a link: `https://www.example.com`\n")
		if len(errs) != 0 {
			t.Errorf("expected no violations for code span, got %+v", errs)
		}
	})
}
