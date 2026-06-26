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

package configparse

import (
	"reflect"
	"strings"
	"testing"

	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

// TestParsersValid exercises each individual parser with valid object content.
func TestParsersValid(t *testing.T) {
	tests := []struct {
		name    string
		parser  Parser
		content string
		want    Configuration
	}{
		{
			name:    "JSON object",
			parser:  JSON,
			content: `{"default": true, "MD013": false}`,
			want:    types.ConfigFromMap(map[string]interface{}{"default": true, "MD013": false}),
		},
		{
			name:    "JSONC with comments and trailing comma",
			parser:  JSONC,
			content: "{\n  // line comment\n  \"default\": true,\n  /* block */ \"MD013\": false,\n}",
			want:    types.ConfigFromMap(map[string]interface{}{"default": true, "MD013": false}),
		},
		{
			name:    "YAML object",
			parser:  YAML,
			content: "default: true\nMD013: false\n",
			want:    types.ConfigFromMap(map[string]interface{}{"default": true, "MD013": false}),
		},
		{
			name:    "TOML object",
			parser:  TOML,
			content: "default = true\nMD013 = false\n",
			want:    types.ConfigFromMap(map[string]interface{}{"default": true, "MD013": false}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestParsersNonObjectCoerce verifies that non-object content (valid syntax but
// not a map) coerces to an empty Configuration.
func TestParsersNonObjectCoerce(t *testing.T) {
	tests := []struct {
		name    string
		parser  Parser
		content string
	}{
		{"JSON array", JSON, `[1, 2, 3]`},
		{"JSON number", JSON, `42`},
		{"JSON string", JSON, `"hello"`},
		{"JSON null", JSON, `null`},
		{"JSON bool", JSON, `true`},
		{"JSONC array", JSONC, "[1, 2, 3,]"},
		{"YAML scalar", YAML, "hello"},
		{"YAML sequence", YAML, "- a\n- b\n"},
		{"YAML empty", YAML, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.ToMap()) != 0 {
				t.Errorf("expected empty Configuration, got %#v", got)
			}
		})
	}
}

// TestParsersInvalid verifies that syntactically invalid content returns an error.
func TestParsersInvalid(t *testing.T) {
	tests := []struct {
		name    string
		parser  Parser
		content string
	}{
		{"JSON garbage", JSON, `{not valid json`},
		{"JSONC garbage", JSONC, `{"a": }`},
		{"YAML bad indent", YAML, "a:\n  - b\n - c\n"},
		{"TOML garbage", TOML, "= this is not toml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser(tt.content)
			if err == nil {
				t.Fatalf("expected error, got nil (result %#v)", got)
			}
		})
	}
}

// TestYAMLNestedNormalize verifies normalizeYAML produces map[string]any for
// nested mappings (via the YAML parser, which exercises the nested path).
func TestYAMLNestedNormalize(t *testing.T) {
	content := "MD013:\n  line_length: 100\n  tables: false\nMD007:\n  indent: 4\n"
	got, err := YAML(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MD013.LineLength == nil || *got.MD013.LineLength != 100 {
		t.Errorf("line_length = %v, want 100", got.MD013.LineLength)
	}
	if got.MD013.Tables == nil || *got.MD013.Tables != false {
		t.Errorf("tables = %v, want false", got.MD013.Tables)
	}
}

// TestNormalizeYAML directly exercises normalizeYAML, including the map[any]any
// path that yaml.v3 does not normally produce but the function must handle.
func TestNormalizeYAML(t *testing.T) {
	t.Run("map[any]any to map[string]any recursive", func(t *testing.T) {
		in := map[interface{}]interface{}{
			"a": 1,
			2:   "two",
			"nested": map[interface{}]interface{}{
				"deep": map[interface{}]interface{}{"x": true},
			},
		}
		out := normalizeYAML(in)
		m, ok := out.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]any, got %T", out)
		}
		if m["a"] != 1 {
			t.Errorf("a = %#v, want 1", m["a"])
		}
		// non-string key fmt.Sprint'd to "2"
		if m["2"] != "two" {
			t.Errorf("key 2 = %#v, want \"two\"", m["2"])
		}
		nested, ok := m["nested"].(map[string]interface{})
		if !ok {
			t.Fatalf("nested should be map[string]any, got %T", m["nested"])
		}
		deep, ok := nested["deep"].(map[string]interface{})
		if !ok {
			t.Fatalf("deep should be map[string]any, got %T", nested["deep"])
		}
		if deep["x"] != true {
			t.Errorf("deep.x = %#v, want true", deep["x"])
		}
	})

	t.Run("slice elements normalized", func(t *testing.T) {
		in := []interface{}{map[interface{}]interface{}{"k": "v"}, 42}
		out := normalizeYAML(in)
		s, ok := out.([]interface{})
		if !ok {
			t.Fatalf("expected []any, got %T", out)
		}
		first, ok := s[0].(map[string]interface{})
		if !ok {
			t.Fatalf("s[0] should be map[string]any, got %T", s[0])
		}
		if first["k"] != "v" {
			t.Errorf("s[0].k = %#v, want \"v\"", first["k"])
		}
		if s[1] != 42 {
			t.Errorf("s[1] = %#v, want 42", s[1])
		}
	})

	t.Run("scalar passthrough", func(t *testing.T) {
		if got := normalizeYAML("plain"); got != "plain" {
			t.Errorf("got %#v, want \"plain\"", got)
		}
		if got := normalizeYAML(123); got != 123 {
			t.Errorf("got %#v, want 123", got)
		}
	})
}

// TestDefaultAndCommon verifies the parser-set helpers.
func TestDefaultAndCommon(t *testing.T) {
	if got := len(Default()); got != 1 {
		t.Errorf("Default() length = %d, want 1", got)
	}
	if got := len(Common()); got != 4 {
		t.Errorf("Common() length = %d, want 4", got)
	}
	// Default() must parse JSON but not YAML-only content (YAML scalar would
	// coerce, so use clearly non-JSON syntax to confirm JSON is the parser).
	if cfg, err := Default()[0](`{"MD013": false}`); err != nil || cfg.MD013.Enabled == nil || *cfg.MD013.Enabled {
		t.Errorf("Default()[0] should parse JSON: cfg=%#v err=%v", cfg, err)
	}
}

// TestParseFirstSuccessWins verifies Parse advances past a failing parser to a
// later successful one. JSON (parser 0) fails on the comment/trailing-comma;
// JSONC (parser 1) succeeds, so its non-empty result is returned.
func TestParseFirstSuccessWins(t *testing.T) {
	content := "{\n  // comment\n  \"default\": true,\n}"
	if _, jerr := JSON(content); jerr == nil {
		t.Fatal("precondition: strict JSON should reject this content")
	}
	got, err := Parse("config.jsonc", content, Common())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Default == nil || !*got.Default {
		t.Errorf("default = %v, want true", got.Default)
	}
}

// TestParseYAMLCoercesBeforeTOML is a characterization test of parser ordering.
// NOTE: TOML-style content like "default = true" is parsed by YAML (parser 2 in
// Common()) as a scalar STRING, which coerce() turns into an empty map. Because
// YAML succeeds before TOML is reached, Parse returns an empty Configuration
// rather than the TOML-decoded {"default": true}. This documents the current
// ordering behavior.
func TestParseYAMLCoercesBeforeTOML(t *testing.T) {
	// Confirm the standalone parsers behave as the ordering assumes.
	if y, err := YAML("default = true\n"); err != nil || len(y.ToMap()) != 0 {
		t.Fatalf("precondition YAML: cfg=%#v err=%v (want empty, no error)", y, err)
	}
	if tm, err := TOML("default = true\n"); err != nil || tm.Default == nil || !*tm.Default {
		t.Fatalf("precondition TOML: cfg=%#v err=%v (want {default:true})", tm, err)
	}
	got, err := Parse("config.toml", "default = true\n", Common())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.ToMap()) != 0 {
		t.Errorf("expected empty map (YAML coerced and won), got %#v", got)
	}
}

func TestParseJSONWins(t *testing.T) {
	got, err := Parse("config.json", `{"MD013": false}`, Common())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MD013.Enabled == nil || *got.MD013.Enabled {
		t.Errorf("MD013 = %v, want disabled", got.MD013.Enabled)
	}
}

// TestParseNilParsersDefaultsToJSON verifies nil/empty parsers use Default().
func TestParseNilParsersDefaultsToJSON(t *testing.T) {
	for _, name := range []string{"nil", "empty"} {
		t.Run(name, func(t *testing.T) {
			var parsers []Parser
			if name == "empty" {
				parsers = []Parser{}
			}
			got, err := Parse("c.json", `{"MD013": false}`, parsers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.MD013.Enabled == nil || *got.MD013.Enabled {
				t.Errorf("MD013 = %v, want disabled", got.MD013.Enabled)
			}
			// TOML-only content must FAIL because only JSON is tried.
			if _, err := Parse("c.toml", "x = 1\n", parsers); err == nil {
				t.Errorf("expected JSON-only default to fail on TOML content")
			}
		})
	}
}

// TestParseTotalFailure verifies the aggregated error message.
func TestParseTotalFailure(t *testing.T) {
	// Content invalid for all four common parsers.
	bad := "{this is : not parseable [ by anything"
	got, err := Parse("broken.cfg", bad, Common())
	if err == nil {
		t.Fatalf("expected error, got nil (result %#v)", got)
	}
	if len(got.ToMap()) != 0 {
		t.Errorf("expected empty Configuration on failure, got %#v", got)
	}
	if !strings.Contains(err.Error(), "Unable to parse") {
		t.Errorf("error should contain \"Unable to parse\", got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "broken.cfg") {
		t.Errorf("error should mention the name, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "Parser 0") {
		t.Errorf("error should enumerate parser failures, got %q", err.Error())
	}
}
