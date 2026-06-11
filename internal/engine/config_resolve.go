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
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

// mapAliasToRuleNames maps every rule name/alias/tag (upper-cased) to the
// canonical rule name(s) it refers to.
func mapAliasToRuleNames(ruleList []*rule.Rule) map[string][]string {
	out := map[string][]string{}

	for _, r := range ruleList {
		canonical := strings.ToUpper(r.Names[0])
		for _, name := range r.Names {
			out[strings.ToUpper(name)] = []string{canonical}
		}

		for _, tag := range r.Tags {
			tu := strings.ToUpper(tag)
			out[tu] = append(out[tu], canonical)
		}
	}

	return out
}

type effectiveConfigResult struct {
	effectiveConfig map[string]map[string]interface{}
	rulesEnabled    map[string]bool
	rulesSeverity   map[string]types.Severity
}

// getEffectiveConfig applies and normalizes the configuration for ruleList.
func getEffectiveConfig(
	ruleList []*rule.Rule,
	cfg map[string]interface{},
	alias map[string][]string,
) effectiveConfigResult {
	ruleDefaultEnable := true
	ruleDefaultSeverity := types.SeverityError

	for key, value := range cfg {
		if strings.ToUpper(key) == "DEFAULT" {
			ruleDefaultEnable = truthy(value)
			if value == "warning" {
				ruleDefaultSeverity = types.SeverityWarning
			}

			break
		}
	}

	res := effectiveConfigResult{
		effectiveConfig: map[string]map[string]interface{}{},
		rulesEnabled:    map[string]bool{},
		rulesSeverity:   map[string]types.Severity{},
	}

	for _, r := range ruleList {
		name := strings.ToUpper(r.Names[0])
		res.effectiveConfig[name] = map[string]interface{}{}
		res.rulesEnabled[name] = ruleDefaultEnable
		res.rulesSeverity[name] = ruleDefaultSeverity
	}

	for key, value := range cfg {
		keyUpper := strings.ToUpper(key)
		enabled := false
		severity := types.SeverityError
		effectiveValue := map[string]interface{}{}

		if truthy(value) {
			if obj, ok := value.(map[string]interface{}); ok {
				enabled = true
				if v, has := obj["enabled"]; has {
					enabled = truthy(v)
				}

				if obj["severity"] == "warning" {
					severity = types.SeverityWarning
				}

				for k, v := range obj {
					if k != "enabled" && k != "severity" {
						effectiveValue[k] = v
					}
				}
			} else {
				enabled = true

				if value == "warning" {
					severity = types.SeverityWarning
				}
			}
		}

		for _, name := range alias[keyUpper] {
			res.effectiveConfig[name] = effectiveValue
			res.rulesEnabled[name] = enabled
			res.rulesSeverity[name] = severity
		}
	}

	return res
}

// truthy reports whether a config value is truthy.
func truthy(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int:
		return t != 0
	case float64:
		return t != 0
	default:
		return true
	}
}
