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
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ldmonster/go-markdownlint/configparse"
	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/rules"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

// --- test helpers ---------------------------------------------------------

// simpleRule builds a CustomRule with one or more names/tags and a Fn.
func simpleRule(names, tags []string, desc string, fn func(p *rule.RuleParams, onError types.OnError)) *rule.Rule {
	return &rule.Rule{
		Names:       names,
		Description: desc,
		Tags:        tags,
		Parser:      types.ParserNone,
		Fn:          fn,
	}
}

// alwaysFireRule fires once at line 1 on every document.
func alwaysFireRule(names, tags []string) *rule.Rule {
	return simpleRule(names, tags, "always fires", func(p *rule.RuleParams, onError types.OnError) {
		onError(types.ErrorInfo{LineNumber: 1, Detail: "boom"})
	})
}

// =====================================================================
// truthy
// =====================================================================

func TestTruthy(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want bool
	}{
		{"nil", nil, false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"empty string", "", false},
		{"non-empty string", "warning", true},
		{"zero int", 0, false},
		{"non-zero int", 5, true},
		{"zero float", 0.0, false},
		{"non-zero float", 1.5, true},
		{"object map", map[string]interface{}{"a": 1}, true},
		{"empty object map", map[string]interface{}{}, true},
		{"slice", []int{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truthy(tt.in); got != tt.want {
				t.Fatalf("truthy(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// =====================================================================
// mapAliasToRuleNames
// =====================================================================

func TestMapAliasToRuleNames(t *testing.T) {
	r1 := simpleRule([]string{"MD900", "custom-one"}, []string{"shared", "tagone"}, "one", func(*rule.RuleParams, types.OnError) {})
	r2 := simpleRule([]string{"MD901", "custom-two"}, []string{"shared", "tagtwo"}, "two", func(*rule.RuleParams, types.OnError) {})
	alias := mapAliasToRuleNames([]*rule.Rule{r1, r2})

	// Canonical name maps to itself.
	if got := alias["MD900"]; !reflect.DeepEqual(got, []string{"MD900"}) {
		t.Errorf("MD900 -> %v", got)
	}
	// Alias maps to canonical (upper-cased canonical = first name upper).
	if got := alias["CUSTOM-ONE"]; !reflect.DeepEqual(got, []string{"MD900"}) {
		t.Errorf("CUSTOM-ONE -> %v", got)
	}
	if got := alias["CUSTOM-TWO"]; !reflect.DeepEqual(got, []string{"MD901"}) {
		t.Errorf("CUSTOM-TWO -> %v", got)
	}
	// Per-rule unique tag maps to that one rule.
	if got := alias["TAGONE"]; !reflect.DeepEqual(got, []string{"MD900"}) {
		t.Errorf("TAGONE -> %v", got)
	}
	// Shared tag fans out to multiple rules in registration order.
	if got := alias["SHARED"]; !reflect.DeepEqual(got, []string{"MD900", "MD901"}) {
		t.Errorf("SHARED -> %v", got)
	}
}

func TestMapAliasToRuleNamesBuiltIn(t *testing.T) {
	alias := mapAliasToRuleNames(rules.BuiltIn())
	// MD009 alias is no-trailing-spaces.
	if got := alias["NO-TRAILING-SPACES"]; !reflect.DeepEqual(got, []string{"MD009"}) {
		t.Errorf("NO-TRAILING-SPACES -> %v", got)
	}
	if got := alias["MD009"]; !reflect.DeepEqual(got, []string{"MD009"}) {
		t.Errorf("MD009 -> %v", got)
	}
	// The "whitespace" tag should include MD009 among potentially others.
	ws := alias["WHITESPACE"]
	found := false
	for _, n := range ws {
		if n == "MD009" {
			found = true
		}
	}
	if !found {
		t.Errorf("WHITESPACE tag did not include MD009: %v", ws)
	}
}

// =====================================================================
// getEffectiveConfig
// =====================================================================

func TestGetEffectiveConfig(t *testing.T) {
	r1 := alwaysFireRule([]string{"MD900", "custom-one"}, []string{"shared"})
	r2 := alwaysFireRule([]string{"MD901", "custom-two"}, []string{"shared"})
	ruleList := []*rule.Rule{r1, r2}
	alias := mapAliasToRuleNames(ruleList)

	t.Run("default true", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{"default": true}, alias)
		if !ec.rulesEnabled["MD900"] || !ec.rulesEnabled["MD901"] {
			t.Fatalf("expected all enabled: %v", ec.rulesEnabled)
		}
		if ec.rulesSeverity["MD900"] != types.SeverityError {
			t.Fatalf("severity = %v", ec.rulesSeverity["MD900"])
		}
	})

	t.Run("default false", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{"default": false}, alias)
		if ec.rulesEnabled["MD900"] || ec.rulesEnabled["MD901"] {
			t.Fatalf("expected all disabled: %v", ec.rulesEnabled)
		}
	})

	t.Run("default warning", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{"default": "warning"}, alias)
		if !ec.rulesEnabled["MD900"] {
			t.Fatalf("warning default should be enabled")
		}
		if ec.rulesSeverity["MD900"] != types.SeverityWarning {
			t.Fatalf("severity = %v, want warning", ec.rulesSeverity["MD900"])
		}
	})

	t.Run("string warning per-rule", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{"default": false, "MD900": "warning"}, alias)
		if !ec.rulesEnabled["MD900"] {
			t.Fatalf("MD900 should be enabled")
		}
		if ec.rulesSeverity["MD900"] != types.SeverityWarning {
			t.Fatalf("MD900 severity = %v, want warning", ec.rulesSeverity["MD900"])
		}
		if ec.rulesEnabled["MD901"] {
			t.Fatalf("MD901 should be disabled")
		}
	})

	t.Run("object enabled false", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"MD900": map[string]interface{}{"enabled": false, "opt": 1},
		}, alias)
		if ec.rulesEnabled["MD900"] {
			t.Fatalf("MD900 should be disabled via object enabled:false")
		}
		// effectiveConfig strips enabled/severity.
		cfg := ec.effectiveConfig["MD900"]
		if _, has := cfg["enabled"]; has {
			t.Fatalf("effectiveConfig should not contain enabled: %v", cfg)
		}
		if cfg["opt"] != 1 {
			t.Fatalf("effectiveConfig opt = %v", cfg["opt"])
		}
	})

	t.Run("object severity warning", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"MD900": map[string]interface{}{"severity": "warning", "br_spaces": 2},
		}, alias)
		if !ec.rulesEnabled["MD900"] {
			t.Fatalf("object form without enabled defaults to enabled")
		}
		if ec.rulesSeverity["MD900"] != types.SeverityWarning {
			t.Fatalf("severity = %v want warning", ec.rulesSeverity["MD900"])
		}
		if ec.effectiveConfig["MD900"]["br_spaces"] != 2 {
			t.Fatalf("br_spaces stripped: %v", ec.effectiveConfig["MD900"])
		}
	})

	t.Run("types.Configuration object form", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"MD900": map[string]interface{}{"enabled": true, "severity": "warning", "x": "y"},
		}, alias)
		if !ec.rulesEnabled["MD900"] {
			t.Fatalf("should be enabled")
		}
		if ec.rulesSeverity["MD900"] != types.SeverityWarning {
			t.Fatalf("severity = %v", ec.rulesSeverity["MD900"])
		}
		if ec.effectiveConfig["MD900"]["x"] != "y" {
			t.Fatalf("x stripped: %v", ec.effectiveConfig["MD900"])
		}
	})

	t.Run("alias key application", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"default":    false,
			"custom-one": true,
		}, alias)
		if !ec.rulesEnabled["MD900"] {
			t.Fatalf("alias custom-one should enable MD900")
		}
		if ec.rulesEnabled["MD901"] {
			t.Fatalf("MD901 should remain disabled")
		}
	})

	t.Run("tag key fans out", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"default": false,
			"shared":  true,
		}, alias)
		if !ec.rulesEnabled["MD900"] || !ec.rulesEnabled["MD901"] {
			t.Fatalf("shared tag should enable both: %v", ec.rulesEnabled)
		}
	})

	t.Run("nil value disables", func(t *testing.T) {
		ec := getEffectiveConfig(ruleList, map[string]interface{}{
			"default": true,
			"MD900":   nil,
		}, alias)
		if ec.rulesEnabled["MD900"] {
			t.Fatalf("nil value should disable MD900")
		}
	})
}

// =====================================================================
// removeFrontMatter
// =====================================================================

func TestRemoveFrontMatter(t *testing.T) {
	t.Run("nil regexp", func(t *testing.T) {
		content := "---\ntitle: x\n---\nbody\n"
		rest, fmLines := removeFrontMatter(content, nil)
		if rest != content {
			t.Fatalf("nil regexp must return content unchanged")
		}
		if fmLines != nil {
			t.Fatalf("nil regexp must return nil fm lines, got %v", fmLines)
		}
	})

	t.Run("yaml present", func(t *testing.T) {
		content := "---\ntitle: hello\nauthor: me\n---\n# Heading\n"
		rest, fmLines := removeFrontMatter(content, helpers.FrontMatterRe)
		if !strings.HasPrefix(rest, "# Heading") {
			t.Fatalf("rest = %q", rest)
		}
		// Front matter lines should be: ---, title:..., author:..., ---
		if len(fmLines) != 4 {
			t.Fatalf("fmLines = %v (len %d)", fmLines, len(fmLines))
		}
		if fmLines[0] != "---" || fmLines[len(fmLines)-1] != "---" {
			t.Fatalf("fence lines wrong: %v", fmLines)
		}
	})

	t.Run("toml present", func(t *testing.T) {
		content := "+++\ntitle = \"hi\"\n+++\nbody text\n"
		rest, fmLines := removeFrontMatter(content, helpers.FrontMatterRe)
		if !strings.HasPrefix(rest, "body text") {
			t.Fatalf("rest = %q", rest)
		}
		if len(fmLines) == 0 {
			t.Fatalf("expected toml front matter lines")
		}
		if fmLines[0] != "+++" {
			t.Fatalf("first fm line = %q", fmLines[0])
		}
	})

	t.Run("absent", func(t *testing.T) {
		content := "# No front matter\nbody\n"
		rest, fmLines := removeFrontMatter(content, helpers.FrontMatterRe)
		if rest != content {
			t.Fatalf("rest should equal content, got %q", rest)
		}
		if fmLines != nil {
			t.Fatalf("fmLines should be nil, got %v", fmLines)
		}
	})

	t.Run("not at index 0", func(t *testing.T) {
		content := "intro\n---\ntitle: x\n---\nbody\n"
		rest, fmLines := removeFrontMatter(content, helpers.FrontMatterRe)
		if rest != content {
			t.Fatalf("front matter not at start must be ignored, rest=%q", rest)
		}
		if fmLines != nil {
			t.Fatalf("fmLines should be nil, got %v", fmLines)
		}
	})
}

// =====================================================================
// validateRuleList
// =====================================================================

func TestValidateRuleList(t *testing.T) {
	noop := func(*rule.RuleParams, types.OnError) {}

	t.Run("built-in valid", func(t *testing.T) {
		if err := validateRuleList(rules.BuiltIn()); err != nil {
			t.Fatalf("built-in rules should validate: %v", err)
		}
	})

	t.Run("valid custom rule", func(t *testing.T) {
		r := simpleRule([]string{"MD900", "my-rule"}, []string{"mytag"}, "desc", noop)
		if err := validateRuleList([]*rule.Rule{r}); err != nil {
			t.Fatalf("valid custom rule failed: %v", err)
		}
	})

	tests := []struct {
		name    string
		rule    *rule.Rule
		wantSub string
	}{
		{
			"no names",
			&rule.Rule{Names: nil, Tags: []string{"t"}, Description: "d", Fn: noop},
			"no names",
		},
		{
			"empty name",
			&rule.Rule{Names: []string{""}, Tags: []string{"t"}, Description: "d", Fn: noop},
			"empty name",
		},
		{
			"no tags",
			&rule.Rule{Names: []string{"MD900"}, Tags: nil, Description: "d", Fn: noop},
			"no tags",
		},
		{
			"no description",
			&rule.Rule{Names: []string{"MD900"}, Tags: []string{"t"}, Description: "", Fn: noop},
			"no description",
		},
		{
			"no function",
			&rule.Rule{Names: []string{"MD900"}, Tags: []string{"t"}, Description: "d", Fn: nil},
			"no function",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuleList([]*rule.Rule{tt.rule})
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantSub)
			}
		})
	}

	t.Run("name collision between rules", func(t *testing.T) {
		r1 := simpleRule([]string{"MD900"}, []string{"t1"}, "d", noop)
		r2 := simpleRule([]string{"MD900"}, []string{"t2"}, "d", noop)
		err := validateRuleList([]*rule.Rule{r1, r2})
		if err == nil || !strings.Contains(err.Error(), "already used") {
			t.Fatalf("expected name collision error, got %v", err)
		}
	})

	t.Run("name collides with prior tag", func(t *testing.T) {
		r1 := simpleRule([]string{"MD900"}, []string{"shared"}, "d", noop)
		r2 := simpleRule([]string{"SHARED"}, []string{"t2"}, "d", noop)
		err := validateRuleList([]*rule.Rule{r1, r2})
		if err == nil || !strings.Contains(err.Error(), "already used") {
			t.Fatalf("expected name/tag collision error, got %v", err)
		}
	})

	t.Run("tag collides with prior name", func(t *testing.T) {
		r1 := simpleRule([]string{"MD900"}, []string{"t1"}, "d", noop)
		r2 := simpleRule([]string{"MD901"}, []string{"MD900"}, "d", noop)
		err := validateRuleList([]*rule.Rule{r1, r2})
		if err == nil || !strings.Contains(err.Error(), "already used as a name") {
			t.Fatalf("expected tag/name collision error, got %v", err)
		}
	})

	t.Run("two rules sharing a tag is allowed", func(t *testing.T) {
		r1 := simpleRule([]string{"MD900"}, []string{"shared"}, "d", noop)
		r2 := simpleRule([]string{"MD901"}, []string{"shared"}, "d", noop)
		if err := validateRuleList([]*rule.Rule{r1, r2}); err != nil {
			t.Fatalf("shared tag between rules should be allowed: %v", err)
		}
	})
}

// =====================================================================
// getEnabledRulesPerLineNumber
// =====================================================================

// enabledAt returns whether ruleName is enabled at 1-based document line.
func enabledAt(res enabledPerLineResult, fmLen, line int, ruleName string) bool {
	idx := fmLen + line
	if idx < 0 || idx >= len(res.enabledRulesPerLineNumber) {
		return false
	}
	m := res.enabledRulesPerLineNumber[idx]
	return m != nil && m[ruleName]
}

func TestGetEnabledRulesPerLineNumber(t *testing.T) {
	r1 := alwaysFireRule([]string{"MD900", "rule-one"}, []string{"shared"})
	r2 := alwaysFireRule([]string{"MD901", "rule-two"}, []string{"shared"})
	ruleList := []*rule.Rule{r1, r2}
	alias := mapAliasToRuleNames(ruleList)
	baseCfg := map[string]interface{}{"default": true}

	run := func(lines []string, fmLines []string, noInline bool, cfg map[string]interface{}, parsers []types.ConfigParser) enabledPerLineResult {
		return getEnabledRulesPerLineNumber(ruleList, lines, fmLines, noInline, cfg, parsers, alias)
	}

	t.Run("all enabled by default", func(t *testing.T) {
		lines := []string{"a", "b", "c"}
		res := run(lines, nil, false, baseCfg, nil)
		for i := 1; i <= 3; i++ {
			if !enabledAt(res, 0, i, "MD900") {
				t.Fatalf("MD900 should be enabled at line %d", i)
			}
		}
		// enabledRuleList contains both rules.
		if len(res.enabledRuleList) != 2 {
			t.Fatalf("enabledRuleList len = %d, want 2", len(res.enabledRuleList))
		}
	})

	t.Run("disable then enable inline", func(t *testing.T) {
		lines := []string{
			"line 1",
			"<!-- markdownlint-disable MD900 -->",
			"line 3",
			"<!-- markdownlint-enable MD900 -->",
			"line 5",
		}
		res := run(lines, nil, false, baseCfg, nil)
		// The disable comment on line 2 takes effect for line 2 onward.
		if enabledAt(res, 0, 2, "MD900") {
			t.Fatalf("MD900 should be disabled at line 2")
		}
		if enabledAt(res, 0, 3, "MD900") {
			t.Fatalf("MD900 should be disabled at line 3")
		}
		// Re-enabled at line 4 onward.
		if !enabledAt(res, 0, 4, "MD900") {
			t.Fatalf("MD900 should be re-enabled at line 4")
		}
		if !enabledAt(res, 0, 5, "MD900") {
			t.Fatalf("MD900 should be enabled at line 5")
		}
		// MD901 unaffected.
		if !enabledAt(res, 0, 3, "MD901") {
			t.Fatalf("MD901 should remain enabled at line 3")
		}
	})

	t.Run("disable all (no params)", func(t *testing.T) {
		lines := []string{
			"a",
			"<!-- markdownlint-disable -->",
			"b",
		}
		res := run(lines, nil, false, baseCfg, nil)
		if enabledAt(res, 0, 3, "MD900") || enabledAt(res, 0, 3, "MD901") {
			t.Fatalf("blanket disable should turn off all rules")
		}
	})

	t.Run("disable-file / enable-file", func(t *testing.T) {
		lines := []string{
			"top",
			"<!-- markdownlint-disable-file MD900 -->",
			"mid",
		}
		res := run(lines, nil, false, baseCfg, nil)
		// disable-file affects the whole file including the first line.
		if enabledAt(res, 0, 1, "MD900") {
			t.Fatalf("disable-file should disable MD900 on line 1")
		}
		if enabledAt(res, 0, 3, "MD900") {
			t.Fatalf("disable-file should disable MD900 on line 3")
		}
		if !enabledAt(res, 0, 1, "MD901") {
			t.Fatalf("MD901 should be unaffected")
		}
	})

	t.Run("capture and restore", func(t *testing.T) {
		lines := []string{
			"l1",
			"<!-- markdownlint-capture -->",
			"<!-- markdownlint-disable MD900 -->",
			"l4",
			"<!-- markdownlint-restore -->",
			"l6",
		}
		res := run(lines, nil, false, baseCfg, nil)
		if enabledAt(res, 0, 4, "MD900") {
			t.Fatalf("MD900 should be disabled at line 4 after disable")
		}
		// After restore, MD900 returns to captured (enabled) state.
		if !enabledAt(res, 0, 6, "MD900") {
			t.Fatalf("MD900 should be restored (enabled) at line 6")
		}
	})

	t.Run("disable-line", func(t *testing.T) {
		lines := []string{
			"clean",
			"bad <!-- markdownlint-disable-line MD900 -->",
			"clean again",
		}
		res := run(lines, nil, false, baseCfg, nil)
		if enabledAt(res, 0, 2, "MD900") {
			t.Fatalf("disable-line should disable MD900 only on line 2")
		}
		// neighbors unaffected
		if !enabledAt(res, 0, 1, "MD900") || !enabledAt(res, 0, 3, "MD900") {
			t.Fatalf("disable-line should not affect other lines")
		}
	})

	t.Run("disable-next-line", func(t *testing.T) {
		lines := []string{
			"<!-- markdownlint-disable-next-line MD900 -->",
			"target",
			"after",
		}
		res := run(lines, nil, false, baseCfg, nil)
		if enabledAt(res, 0, 2, "MD900") {
			t.Fatalf("disable-next-line should disable MD900 on line 2")
		}
		if !enabledAt(res, 0, 1, "MD900") {
			t.Fatalf("the comment line itself stays enabled")
		}
		if !enabledAt(res, 0, 3, "MD900") {
			t.Fatalf("line 3 should remain enabled")
		}
	})

	t.Run("noInlineConfig ignores comments", func(t *testing.T) {
		lines := []string{
			"a",
			"<!-- markdownlint-disable MD900 -->",
			"b",
		}
		res := run(lines, nil, true, baseCfg, nil)
		if !enabledAt(res, 0, 3, "MD900") {
			t.Fatalf("with noInlineConfig the disable comment must be ignored")
		}
	})

	t.Run("configure-file merge via JSON parser", func(t *testing.T) {
		lines := []string{
			"<!-- markdownlint-configure-file { \"MD900\": false } -->",
			"body",
		}
		parsers := []types.ConfigParser{configparse.JSON}
		res := run(lines, nil, false, baseCfg, parsers)
		if enabledAt(res, 0, 2, "MD900") {
			t.Fatalf("configure-file should disable MD900")
		}
		if !enabledAt(res, 0, 2, "MD901") {
			t.Fatalf("MD901 should stay enabled after configure-file")
		}
	})

	t.Run("front matter offsets per-line entries", func(t *testing.T) {
		fmLines := []string{"---", "title: x", "---"}
		lines := []string{
			"body 1",
			"bad <!-- markdownlint-disable-line MD900 -->",
			"body 3",
		}
		res := run(lines, fmLines, false, baseCfg, nil)
		// total length = 1 + len(fm) + len(lines)
		wantLen := 1 + len(fmLines) + len(lines)
		if len(res.enabledRulesPerLineNumber) != wantLen {
			t.Fatalf("perLine len = %d, want %d", len(res.enabledRulesPerLineNumber), wantLen)
		}
		// disable-line on content line 2 -> absolute index fmLen+2.
		if enabledAt(res, len(fmLines), 2, "MD900") {
			t.Fatalf("disable-line should be offset by front matter")
		}
		if !enabledAt(res, len(fmLines), 1, "MD900") {
			t.Fatalf("line 1 should remain enabled")
		}
	})
}

// =====================================================================
// Run (exported)
// =====================================================================

func TestRunStrings(t *testing.T) {
	cr := alwaysFireRule([]string{"MD900", "always-fire"}, []string{"custom"})
	res, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Strings:     map[string]string{"doc.md": "hello world\nsecond line\n"},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	errs, ok := res["doc.md"]
	if !ok {
		t.Fatalf("no results for doc.md: %v", res)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].LineNumber != 1 || errs[0].RuleNames[0] != "MD900" {
		t.Fatalf("unexpected error: %+v", errs[0])
	}
}

func TestRunNilConfigDefaultsTrue(t *testing.T) {
	// With nil config, default is true so a built-in rule fires.
	res, err := Run(context.Background(), Options{
		Strings: map[string]string{"t.md": "# Heading \n"}, // trailing space -> MD009
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	errs := res["t.md"]
	found := false
	for _, e := range errs {
		if e.RuleNames[0] == "MD009" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected MD009 (trailing space) to fire with default config; got %v", errs)
	}
}

func TestRunFilesTempDir(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(p, []byte("trailing   \nok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Run(context.Background(), Options{
		Config: types.ConfigFromMap(map[string]interface{}{"default": false, "MD009": true}),
		Files:  []string{p},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	errs := res[p]
	if len(errs) == 0 || errs[0].RuleNames[0] != "MD009" {
		t.Fatalf("expected MD009 in %s: %v", p, errs)
	}
}

func TestRunCustomReadFile(t *testing.T) {
	cr := alwaysFireRule([]string{"MD900", "always-fire"}, []string{"custom"})
	res, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Files:       []string{"virtual.md"},
		ReadFile: func(path string) (string, error) {
			if path != "virtual.md" {
				t.Errorf("unexpected path %q", path)
			}
			return "synthetic content\n", nil
		},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res["virtual.md"]) != 1 {
		t.Fatalf("expected 1 error from virtual file: %v", res["virtual.md"])
	}
}

func TestRunReadFileError(t *testing.T) {
	_, err := Run(context.Background(), Options{
		Files: []string{"nope.md"},
		ReadFile: func(path string) (string, error) {
			return "", os.ErrNotExist
		},
	})
	if err == nil {
		t.Fatalf("expected error when ReadFile fails")
	}
}

func TestRunConcurrency(t *testing.T) {
	cr := alwaysFireRule([]string{"MD900", "always-fire"}, []string{"custom"})
	strs := map[string]string{}
	for i := 0; i < 50; i++ {
		strs["doc"+string(rune('A'+i%26))+string(rune('0'+i/26))] = "content line\n"
	}
	res, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Strings:     strs,
		Concurrency: 4,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res) != len(strs) {
		t.Fatalf("expected %d results, got %d", len(strs), len(res))
	}
	for name, errs := range res {
		if len(errs) != 1 {
			t.Fatalf("doc %s: expected 1 error, got %d", name, len(errs))
		}
	}
}

func TestRunConcurrencyDefaultWhenZero(t *testing.T) {
	cr := alwaysFireRule([]string{"MD900", "always-fire"}, []string{"custom"})
	res, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Strings:     map[string]string{"a.md": "x\n", "b.md": "y\n"},
		Concurrency: 0, // -> defaults to 8
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
}

func TestRunHandleRuleFailures(t *testing.T) {
	panicRule := simpleRule([]string{"MD900", "panic-rule"}, []string{"custom"}, "panics", func(p *rule.RuleParams, onError types.OnError) {
		panic("kaboom")
	})
	res, err := Run(context.Background(), Options{
		Config:             types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules:        []*rule.Rule{panicRule},
		Strings:            map[string]string{"doc.md": "anything\nhere\n"},
		HandleRuleFailures: true,
	})
	if err != nil {
		t.Fatalf("HandleRuleFailures should swallow panic, got err: %v", err)
	}
	errs := res["doc.md"]
	if len(errs) != 1 {
		t.Fatalf("expected 1 synthetic error, got %d: %v", len(errs), errs)
	}
	if errs[0].LineNumber != 1 {
		t.Fatalf("panic error should be reported at line 1, got %d", errs[0].LineNumber)
	}
	if !strings.Contains(errs[0].ErrorDetail, "threw an exception") {
		t.Fatalf("expected exception detail, got %q", errs[0].ErrorDetail)
	}
	if !strings.Contains(errs[0].ErrorDetail, "kaboom") {
		t.Fatalf("expected panic message in detail, got %q", errs[0].ErrorDetail)
	}
}

func TestRunPanicWithoutHandleRuleFailures(t *testing.T) {
	panicRule := simpleRule([]string{"MD900", "panic-rule"}, []string{"custom"}, "panics", func(p *rule.RuleParams, onError types.OnError) {
		panic("kaboom")
	})
	_, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{panicRule},
		Strings:     map[string]string{"doc.md": "anything\n"},
	})
	if err == nil {
		t.Fatalf("expected error propagated from panicking rule")
	}
	if !strings.Contains(err.Error(), "threw an exception") {
		t.Fatalf("error = %q, expected exception message", err.Error())
	}
	if !strings.Contains(err.Error(), "MD900") {
		t.Fatalf("error should name the rule: %q", err.Error())
	}
}

func TestRunPanicWithErrorValue(t *testing.T) {
	// recoverMessage handles error-typed panics.
	panicRule := simpleRule([]string{"MD900", "panic-rule"}, []string{"custom"}, "panics", func(p *rule.RuleParams, onError types.OnError) {
		panic(os.ErrPermission)
	})
	res, err := Run(context.Background(), Options{
		Config:             types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules:        []*rule.Rule{panicRule},
		Strings:            map[string]string{"doc.md": "x\n"},
		HandleRuleFailures: true,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(res["doc.md"][0].ErrorDetail, os.ErrPermission.Error()) {
		t.Fatalf("expected wrapped error message, got %q", res["doc.md"][0].ErrorDetail)
	}
}

func TestRunValidationError(t *testing.T) {
	// A custom rule with no tags must fail validation before linting.
	bad := &rule.Rule{Names: []string{"MD900"}, Description: "d", Fn: func(*rule.RuleParams, types.OnError) {}}
	_, err := Run(context.Background(), Options{
		CustomRules: []*rule.Rule{bad},
		Strings:     map[string]string{"a.md": "x"},
	})
	if err == nil || !strings.Contains(err.Error(), "no tags") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRunDisableFrontMatter(t *testing.T) {
	// When DisableFrontMatter is set, the YAML block is linted as content.
	cr := simpleRule([]string{"MD900", "count-lines"}, []string{"custom"}, "reports line count", func(p *rule.RuleParams, onError types.OnError) {
		onError(types.ErrorInfo{LineNumber: 1, Detail: lineCountDetail(len(p.Lines))})
	})
	content := "---\ntitle: x\n---\nbody\n"

	withFM, err := Run(context.Background(), Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Strings:     map[string]string{"a.md": content},
	})
	if err != nil {
		t.Fatal(err)
	}
	withoutFM, err := Run(context.Background(), Options{
		Config:             types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules:        []*rule.Rule{cr},
		Strings:            map[string]string{"a.md": content},
		DisableFrontMatter: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	d1 := withFM["a.md"][0].ErrorDetail
	d2 := withoutFM["a.md"][0].ErrorDetail
	if d1 == d2 {
		t.Fatalf("expected different line counts with/without front matter: %q vs %q", d1, d2)
	}
}

func TestRunContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cr := alwaysFireRule([]string{"MD900", "always-fire"}, []string{"custom"})
	res, err := Run(ctx, Options{
		Config:      types.ConfigFromMap(map[string]interface{}{"default": false, "MD900": true}),
		CustomRules: []*rule.Rule{cr},
		Strings:     map[string]string{"a.md": "x\n"},
	})
	if err == nil {
		// Cancellation is best-effort; if it slipped through, results must be sane.
		if res == nil {
			t.Fatalf("nil results without error")
		}
		return
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func lineCountDetail(n int) string {
	return "lines=" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// stable sort helper used to make some assertions deterministic if needed.
var _ = sort.Strings
