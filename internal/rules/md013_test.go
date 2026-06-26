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

import (
	"testing"

	markdownlint "github.com/fuldaxxx/go-markdownlint"
)

func TestMD013(t *testing.T) {
	longLine := "This is a very long line that exceeds eighty characters in total length yes indeed it does go on\n"

	t.Run("positive_long_line", func(t *testing.T) {
		errs := lintRule(t, "MD013", longLine)
		if !ruleFired(errs, "MD013") {
			t.Fatalf("expected MD013 to fire, got %v", errs)
		}
		if errs[0].LineNumber != 1 {
			t.Errorf("expected line 1, got %d", errs[0].LineNumber)
		}
		if errs[0].ErrorDetail != "Expected: 80; Actual: 96" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
		if errs[0].ErrorRange == nil || (*errs[0].ErrorRange)[0] != 81 {
			t.Errorf("unexpected range %v", errs[0].ErrorRange)
		}
	})

	t.Run("negative_short_line", func(t *testing.T) {
		errs := lintRule(t, "MD013", "Short line.\n")
		if ruleFired(errs, "MD013") {
			t.Fatalf("expected no MD013, got %v", errs)
		}
	})

	t.Run("config_line_length_200_clean", func(t *testing.T) {
		errs := lintFull(t, longLine,
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD013": map[string]interface{}{"line_length": 200}}))
		if ruleFired(errs, "MD013") {
			t.Fatalf("expected no MD013 with line_length=200, got %v", errs)
		}
	})

	t.Run("config_line_length_40_fires", func(t *testing.T) {
		errs := lintFull(t, longLine,
			markdownlint.ConfigFromMap(map[string]interface{}{"default": false, "MD013": map[string]interface{}{"line_length": 40}}))
		if !ruleFired(errs, "MD013") {
			t.Fatalf("expected MD013 with line_length=40, got %v", errs)
		}
		if errs[0].ErrorDetail != "Expected: 40; Actual: 96" {
			t.Errorf("unexpected detail %q", errs[0].ErrorDetail)
		}
	})
}
