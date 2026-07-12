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

package styles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ldmonster/go-markdownlint/internal/types"
)

// styleDir is the canonical JSON source used by the embedded styles.
const styleDir = "/home/cnupt/work/github/markdownlint/style"

func loadStyleJSON(t *testing.T, name string) types.Configuration {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(styleDir, name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshaling %s: %v", name, err)
	}
	return types.ConfigFromMap(m)
}

// TestStylesNonEmpty verifies each style accessor returns a non-empty map.
func TestStylesNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		fn   func() types.Configuration
	}{
		{"All", All},
		{"Relaxed", Relaxed},
		{"Prettier", Prettier},
		{"Cirosantilli", Cirosantilli},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if len(got.ToMap()) == 0 {
				t.Fatal("got empty Configuration")
			}
		})
	}
}

// TestStylesMatchSourceJSON verifies each embedded style equals the canonical
// JSON file in the style/ directory.
func TestStylesMatchSourceJSON(t *testing.T) {
	tests := []struct {
		fn   func() types.Configuration
		file string
	}{
		{All, "all.json"},
		{Relaxed, "relaxed.json"},
		{Prettier, "prettier.json"},
		{Cirosantilli, "cirosantilli.json"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			want := loadStyleJSON(t, tt.file)
			got := tt.fn()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("style %s mismatch:\n got=%#v\nwant=%#v", tt.file, got, want)
			}
		})
	}
}

// disabledByConfig reports whether the rule entry is present and disabled
// (Enabled == false).
func disabledRule(enabled *bool) bool { return enabled != nil && !*enabled }

// TestAllStyle verifies the "all" style enables every rule by default.
func TestAllStyle(t *testing.T) {
	c := All()
	if c.Default == nil || !*c.Default {
		t.Errorf("All default = %v, want true", c.Default)
	}
	// The documentation "comment" key lands in Custom.
	if _, ok := c.Custom["comment"]; !ok {
		t.Errorf("All should carry a comment key, got %#v", c.Custom)
	}
}

// TestRelaxedStyle verifies relaxed enables default and disables several rules,
// including the "whitespace" tag key and alias-keyed rules.
func TestRelaxedStyle(t *testing.T) {
	c := Relaxed()
	if c.Default == nil || !*c.Default {
		t.Errorf("Relaxed default = %v, want true", c.Default)
	}
	// "whitespace" is a tag key -> Custom; the rest are alias-keyed rules.
	if rc, ok := c.Custom["whitespace"]; !ok || !disabledRule(rc.Enabled) {
		t.Errorf("Relaxed whitespace tag = %#v, want disabled", c.Custom["whitespace"])
	}
	if !disabledRule(c.MD007.Enabled) { // ul-indent
		t.Errorf("Relaxed MD007 (ul-indent) should be disabled")
	}
	if !disabledRule(c.MD033.Enabled) { // no-inline-html
		t.Errorf("Relaxed MD033 (no-inline-html) should be disabled")
	}
	if !disabledRule(c.MD041.Enabled) { // first-line-h1
		t.Errorf("Relaxed MD041 (first-line-h1) should be disabled")
	}
}

// TestPrettierStyle verifies prettier disables conflicting rules (no default key).
func TestPrettierStyle(t *testing.T) {
	c := Prettier()
	if c.Default != nil {
		t.Errorf("Prettier should not set \"default\", got %v", c.Default)
	}
	if !disabledRule(c.MD013.Enabled) { // line-length
		t.Errorf("Prettier MD013 (line-length) should be disabled")
	}
	if !disabledRule(c.MD010.Enabled) { // no-hard-tabs
		t.Errorf("Prettier MD010 (no-hard-tabs) should be disabled")
	}
	if !disabledRule(c.MD007.Enabled) { // ul-indent
		t.Errorf("Prettier MD007 (ul-indent) should be disabled")
	}
}

// TestCirosantilliStyle verifies cirosantilli enables default and sets options.
func TestCirosantilliStyle(t *testing.T) {
	c := Cirosantilli()
	if c.Default == nil || !*c.Default {
		t.Errorf("Cirosantilli default = %v, want true", c.Default)
	}
	if !disabledRule(c.MD033.Enabled) {
		t.Errorf("Cirosantilli MD033 should be disabled")
	}
	if c.MD003.Style == nil || *c.MD003.Style != "atx" {
		t.Errorf("MD003.style = %v, want \"atx\"", c.MD003.Style)
	}
	if c.MD030.UlMulti == nil || *c.MD030.UlMulti != 3 {
		t.Errorf("MD030.ul_multi = %v, want 3", c.MD030.UlMulti)
	}
	if c.MD030.OlMulti == nil || *c.MD030.OlMulti != 2 {
		t.Errorf("MD030.ol_multi = %v, want 2", c.MD030.OlMulti)
	}
}
