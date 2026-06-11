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

package engine

import (
	"fmt"
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/rule"
)

// validateRuleList checks custom rules for structural validity and name/tag
// collisions. Built-in rules are assumed valid.
func validateRuleList(ruleList []*rule.Rule) error {
	allIDs := map[string]bool{} // value: true if a name, false if a tag

	for _, r := range ruleList {
		if len(r.Names) == 0 {
			return fmt.Errorf("rule has no names")
		}

		for _, n := range r.Names {
			if n == "" {
				return fmt.Errorf("rule %v has an empty name", r.Names)
			}
		}

		if len(r.Tags) == 0 {
			return fmt.Errorf("rule %v has no tags", r.Names)
		}

		if r.Description == "" {
			return fmt.Errorf("rule %v has no description", r.Names)
		}

		if r.Fn == nil {
			return fmt.Errorf("rule %v has no function", r.Names)
		}

		for _, name := range r.Names {
			u := strings.ToUpper(name)
			if _, exists := allIDs[u]; exists {
				return fmt.Errorf(
					"name '%s' of rule %v is already used as a name or tag",
					name,
					r.Names,
				)
			}

			allIDs[u] = true
		}

		for _, tag := range r.Tags {
			u := strings.ToUpper(tag)
			if v, exists := allIDs[u]; exists && v {
				return fmt.Errorf("tag '%s' of rule %v is already used as a name", tag, r.Names)
			}

			allIDs[u] = false
		}
	}

	return nil
}
