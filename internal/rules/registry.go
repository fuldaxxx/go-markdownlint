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

// Package rules contains the built-in markdownlint rules and a registry that
// exposes them in a deterministic order with documentation URLs attached.
package rules

import (
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/ldmonster/go-markdownlint/internal/rule"
)

// Version is the library version.
const Version = "0.40.0"

// Homepage is the project homepage used to build rule documentation URLs.
const Homepage = "https://github.com/DavidAnson/markdownlint"

// FixableRuleNames lists the rules that can emit fixes.
var FixableRuleNames = []string{
	"MD004", "MD005", "MD007", "MD009", "MD010", "MD011",
	"MD012", "MD014", "MD018", "MD019", "MD020", "MD021",
	"MD022", "MD023", "MD026", "MD027", "MD029", "MD030",
	"MD031", "MD032", "MD034", "MD037", "MD038", "MD039",
	"MD044", "MD047", "MD049", "MD050", "MD051", "MD053",
	"MD054", "MD058",
}

var (
	registry []*rule.Rule
	once     sync.Once
)

// register adds a rule to the built-in registry. Called from rule file init().
func register(r *rule.Rule) {
	registry = append(registry, r)
}

// BuiltIn returns the built-in rules, sorted by primary name, each with its
// documentation Information URL set.
func BuiltIn() []*rule.Rule {
	once.Do(func() {
		sort.SliceStable(registry, func(i, j int) bool {
			return registry[i].Names[0] < registry[j].Names[0]
		})

		for _, r := range registry {
			name := strings.ToLower(r.Names[0])
			u, _ := url.Parse(Homepage + "/blob/v" + Version + "/doc/" + name + ".md")
			r.Information = u
		}
	})

	return registry
}
