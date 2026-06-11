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
	"context"
	"testing"

	markdownlint "github.com/ldmonster/go-markdownlint"
)

// lintB lints content with only the given rule enabled (boolean true).
// Helper namespace is suffixed "B" to avoid collisions with helpers defined by
// other test files in this package (MD001..MD031 batch).
func lintB(t *testing.T, rule, content string) []markdownlint.Error {
	t.Helper()
	errs, err := markdownlint.LintString(context.Background(), "t.md", content,
		markdownlint.ConfigFromMap(map[string]interface{}{"default": false, rule: true}))
	if err != nil {
		t.Fatalf("LintString(%s) error: %v", rule, err)
	}
	return errs
}

// lintBCfg lints content with the given rule enabled using a config map.
func lintBCfg(t *testing.T, rule string, opts map[string]interface{}, content string) []markdownlint.Error {
	t.Helper()
	errs, err := markdownlint.LintString(context.Background(), "t.md", content,
		markdownlint.ConfigFromMap(map[string]interface{}{"default": false, rule: opts}))
	if err != nil {
		t.Fatalf("LintString(%s) error: %v", rule, err)
	}
	return errs
}

// firedAtB reports whether any error fired for the rule at the given line
// (line==0 matches any line).
func firedAtB(errs []markdownlint.Error, rule string, line int) bool {
	for _, e := range errs {
		if line != 0 && e.LineNumber != line {
			continue
		}
		for _, n := range e.RuleNames {
			if n == rule {
				return true
			}
		}
	}
	return false
}
