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

	"github.com/ldmonster/go-markdownlint/configparse"
	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

type enabledPerLineResult struct {
	effectiveConfig           map[string]map[string]interface{}
	enabledRulesPerLineNumber []map[string]bool
	enabledRuleList           []*rule.Rule
	rulesSeverity             map[string]types.Severity
}

func copyBoolMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}

	return out
}

// getEnabledRulesPerLineNumber computes the set of enabled rules per line number.
func getEnabledRulesPerLineNumber(
	ruleList []*rule.Rule,
	lines []string,
	frontMatterLines []string,
	noInlineConfig bool,
	cfg map[string]interface{},
	configParsers []types.ConfigParser,
	alias map[string][]string,
) enabledPerLineResult {
	var (
		enabledRules  map[string]bool
		capturedRules map[string]bool
	)

	// Prefilled with nil entries for line 0 and each front matter line; one entry
	// per content line is appended later, so indices line up with absolute line
	// numbers. The non-zero initial length is intentional.
	enabledRulesPerLineNumber := make([]map[string]bool, 1+len(frontMatterLines))

	var allRuleNames []string

	applyEnableDisable := func(action, parameter string, state map[string]bool) map[string]bool {
		state = copyBoolMap(state)
		enabled := strings.HasPrefix(action, "ENABLE")
		trimmed := strings.TrimSpace(parameter)

		var items []string
		if trimmed != "" {
			items = strings.Fields(strings.ToUpper(trimmed))
		} else {
			items = allRuleNames
		}

		for _, nameUpper := range items {
			for _, ruleName := range alias[nameUpper] {
				state[ruleName] = enabled
			}
		}

		return state
	}

	handleInlineConfig := func(input []string, forEachMatch func(action, parameter string, lineNumber int), forEachLine func()) {
		for lineIndex, line := range input {
			if !noInlineConfig {
				matches := helpers.InlineCommentStartRe.FindAllStringSubmatchIndex(line, -1)
				for _, m := range matches {
					// group 2 is the action keyword (disable/enable/...).
					action := strings.ToUpper(line[m[4]:m[5]])
					// group1 is the whole "<!-- markdownlint-xxx" portion (indices m[2]:m[3]).
					startIndex := m[2] + (m[3] - m[2])

					endIndex := strings.Index(line[startIndex:], "-->")
					if endIndex == -1 {
						break
					}

					endIndex += startIndex
					parameter := line[startIndex:endIndex]
					forEachMatch(action, parameter, lineIndex+1)
				}
			}

			if forEachLine != nil {
				forEachLine()
			}
		}
	}

	// Pass 0: CONFIGURE-FILE over the whole document.
	configureFile := func(action, parameter string, _ int) {
		if action == "CONFIGURE-FILE" {
			parsed, err := configparse.Parse("CONFIGURE-FILE", parameter, configParsers)
			if err == nil {
				merged := map[string]interface{}{}
				for k, v := range cfg {
					merged[k] = v
				}

				for k, v := range parsed.ToMap() {
					merged[k] = v
				}

				cfg = merged
			}
		}
	}
	handleInlineConfig([]string{strings.Join(lines, "\n")}, configureFile, nil)

	ec := getEffectiveConfig(ruleList, cfg, alias)
	allRuleNames = make([]string, 0, len(ec.rulesEnabled))

	for name := range ec.rulesEnabled {
		allRuleNames = append(allRuleNames, name)
	}

	enabledRules = copyBoolMap(ec.rulesEnabled)
	capturedRules = enabledRules

	// Pass 1: ENABLE-FILE / DISABLE-FILE.
	handleInlineConfig(lines, func(action, parameter string, _ int) {
		if action == "ENABLE-FILE" || action == "DISABLE-FILE" {
			enabledRules = applyEnableDisable(action, parameter, enabledRules)
		}
	}, nil)

	// Pass 2: CAPTURE / RESTORE / ENABLE / DISABLE, snapshot per line.
	handleInlineConfig(lines, func(action, parameter string, _ int) {
		switch action {
		case "CAPTURE":
			capturedRules = enabledRules
		case "RESTORE":
			enabledRules = capturedRules
		case "ENABLE", "DISABLE":
			enabledRules = applyEnableDisable(action, parameter, enabledRules)
		}
	}, func() {
		//nolint:makezero // intentional fixed prefix followed by per-line appends
		enabledRulesPerLineNumber = append(enabledRulesPerLineNumber, enabledRules)
	})

	// Pass 3: DISABLE-LINE / DISABLE-NEXT-LINE.
	handleInlineConfig(lines, func(action, parameter string, lineNumber int) {
		if action == "DISABLE-LINE" || action == "DISABLE-NEXT-LINE" {
			offset := 0
			if action == "DISABLE-NEXT-LINE" {
				offset = 1
			}

			idx := len(frontMatterLines) + lineNumber + offset
			if idx >= 0 && idx < len(enabledRulesPerLineNumber) {
				enabledRulesPerLineNumber[idx] = applyEnableDisable(
					action,
					parameter,
					enabledRulesPerLineNumber[idx],
				)
			}
		}
	}, nil)

	// Rules used at least once.
	var enabledRuleList []*rule.Rule

	for _, r := range ruleList {
		name := strings.ToUpper(r.Names[0])
		for _, m := range enabledRulesPerLineNumber {
			if m != nil && m[name] {
				enabledRuleList = append(enabledRuleList, r)
				break
			}
		}
	}

	return enabledPerLineResult{
		effectiveConfig:           ec.effectiveConfig,
		enabledRulesPerLineNumber: enabledRulesPerLineNumber,
		enabledRuleList:           enabledRuleList,
		rulesSeverity:             ec.rulesSeverity,
	}
}
