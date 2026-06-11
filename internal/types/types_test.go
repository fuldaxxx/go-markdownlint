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

package types

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestSeverityConstants(t *testing.T) {
	if SeverityError != "error" {
		t.Errorf("SeverityError = %q", SeverityError)
	}
	if SeverityWarning != "warning" {
		t.Errorf("SeverityWarning = %q", SeverityWarning)
	}
}

func TestParserTypeConstants(t *testing.T) {
	// iota ordering.
	if ParserMicromark != 0 {
		t.Errorf("ParserMicromark = %d, want 0", ParserMicromark)
	}
	if ParserMarkdownIt != 1 {
		t.Errorf("ParserMarkdownIt = %d, want 1", ParserMarkdownIt)
	}
	if ParserNone != 2 {
		t.Errorf("ParserNone = %d, want 2", ParserNone)
	}
}

func TestFixInfoConstruction(t *testing.T) {
	fi := FixInfo{LineNumber: 5, EditColumn: 2, DeleteCount: -1, InsertText: "x"}
	if fi.LineNumber != 5 || fi.EditColumn != 2 || fi.DeleteCount != -1 || fi.InsertText != "x" {
		t.Errorf("FixInfo = %#v", fi)
	}
	// Zero value.
	var z FixInfo
	if z.LineNumber != 0 || z.EditColumn != 0 || z.DeleteCount != 0 || z.InsertText != "" {
		t.Errorf("zero FixInfo = %#v", z)
	}
}

func TestErrorInfoConstruction(t *testing.T) {
	info, _ := url.Parse("https://example.com")
	rng := &[2]int{3, 4}
	fi := &FixInfo{LineNumber: 1}
	ei := ErrorInfo{
		LineNumber:  7,
		Detail:      "detail",
		Context:     "context",
		Information: info,
		Range:       rng,
		FixInfo:     fi,
	}
	if ei.LineNumber != 7 || ei.Detail != "detail" || ei.Context != "context" {
		t.Errorf("ErrorInfo = %#v", ei)
	}
	if ei.Information.Host != "example.com" {
		t.Errorf("Information = %v", ei.Information)
	}
	if ei.Range[0] != 3 || ei.Range[1] != 4 {
		t.Errorf("Range = %v", ei.Range)
	}
	if ei.FixInfo.LineNumber != 1 {
		t.Errorf("FixInfo = %#v", ei.FixInfo)
	}
}

func TestErrorConstruction(t *testing.T) {
	e := Error{
		LineNumber:      10,
		RuleNames:       []string{"MD001", "heading-increment"},
		RuleDescription: "desc",
		RuleInformation: "https://example.com",
		ErrorDetail:     "edetail",
		ErrorContext:    "ectx",
		ErrorRange:      &[2]int{1, 2},
		FixInfo:         &FixInfo{LineNumber: 10, DeleteCount: 1},
		Severity:        SeverityError,
	}
	if e.LineNumber != 10 || e.RuleNames[0] != "MD001" || e.Severity != SeverityError {
		t.Errorf("Error = %#v", e)
	}
}

func TestErrorJSONTags(t *testing.T) {
	e := Error{
		LineNumber:      2,
		RuleNames:       []string{"MD009"},
		RuleDescription: "Trailing spaces",
		ErrorRange:      &[2]int{4, 1},
		Severity:        SeverityWarning,
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"lineNumber", "ruleNames", "ruleDescription", "errorRange", "severity"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing json key %q in %s", key, b)
		}
	}
	if m["severity"] != "warning" {
		t.Errorf("severity json = %v", m["severity"])
	}
}

func TestConfigurationMap(t *testing.T) {
	cfg := ConfigFromMap(map[string]interface{}{
		"default": false,
		"MD009":   true,
		"MD013":   map[string]interface{}{"line_length": 120},
		"extends": "base.json",
	})
	if cfg.Default == nil || *cfg.Default != false {
		t.Errorf("default = %v", cfg.Default)
	}
	if cfg.MD009.Enabled == nil || *cfg.MD009.Enabled != true {
		t.Errorf("MD009 = %v", cfg.MD009.Enabled)
	}
	if cfg.MD013.LineLength == nil || *cfg.MD013.LineLength != 120 {
		t.Errorf("MD013 line_length = %v", cfg.MD013.LineLength)
	}
	if len(cfg.Extends) != 1 || cfg.Extends[0] != "base.json" {
		t.Errorf("extends = %v", cfg.Extends)
	}
}

func TestResultsMap(t *testing.T) {
	r := Results{
		"a.md": {{LineNumber: 1, RuleNames: []string{"MD001"}}},
		"b.md": nil,
	}
	if len(r["a.md"]) != 1 {
		t.Errorf("a.md errors = %d", len(r["a.md"]))
	}
	if r["b.md"] != nil {
		t.Errorf("b.md should be nil")
	}
}

func TestOnErrorCallback(t *testing.T) {
	var got ErrorInfo
	var cb OnError = func(e ErrorInfo) { got = e }
	cb(ErrorInfo{LineNumber: 42, Detail: "d"})
	if got.LineNumber != 42 || got.Detail != "d" {
		t.Errorf("got %#v", got)
	}
}

func TestConfigParser(t *testing.T) {
	var p ConfigParser = func(content string) (Configuration, error) {
		return ConfigFromMap(map[string]interface{}{content: true}), nil
	}
	cfg, err := p("my-rule")
	if err != nil {
		t.Fatal(err)
	}
	if rc, ok := cfg.Custom["my-rule"]; !ok || rc.Enabled == nil || !*rc.Enabled {
		t.Errorf("cfg = %#v", cfg)
	}
}
