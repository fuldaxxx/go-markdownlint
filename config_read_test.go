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

package markdownlint

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadConfigExtendsMerge verifies ReadConfig resolves an "extends" parent
// reference and merges parent-then-child (child overrides parent).
func TestReadConfigExtendsMerge(t *testing.T) {
	dir := t.TempDir()

	parent := filepath.Join(dir, "parent.json")
	parentContent := `{
		"default": true,
		"MD013": false,
		"MD007": {"indent": 2}
	}`
	if err := os.WriteFile(parent, []byte(parentContent), 0o600); err != nil {
		t.Fatal(err)
	}

	child := filepath.Join(dir, "child.json")
	childContent := `{
		"extends": "parent.json",
		"MD013": true,
		"MD009": false
	}`
	if err := os.WriteFile(child, []byte(childContent), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadConfig(child, nil)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}

	// Inherited unchanged from parent.
	if cfg.Default == nil || *cfg.Default != true {
		t.Errorf("default = %#v, want true (from parent)", cfg.Default)
	}
	if cfg.MD007.Indent == nil {
		t.Fatalf("MD007.indent = nil, want 2 (from parent)")
	}
	if *cfg.MD007.Indent != 2 {
		t.Errorf("MD007.indent = %#v, want 2 (from parent)", *cfg.MD007.Indent)
	}

	// Child overrides parent.
	if cfg.MD013.Enabled == nil || *cfg.MD013.Enabled != true {
		t.Errorf("MD013 = %#v, want true (child overrides parent's false)", cfg.MD013.Enabled)
	}

	// Child-only key present.
	if cfg.MD009.Enabled == nil || *cfg.MD009.Enabled != false {
		t.Errorf("MD009 = %#v, want false (child only)", cfg.MD009.Enabled)
	}

	// "extends" must be stripped from the result.
	if len(cfg.Extends) != 0 {
		t.Errorf("extends should be removed from merged config, got %#v", cfg.Extends)
	}
}

// TestReadConfigNoExtends verifies a plain config with no extends is returned as-is.
func TestReadConfigNoExtends(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.json")
	if err := os.WriteFile(file, []byte(`{"default": false, "MD001": true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := ReadConfig(file, nil)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if cfg.Default == nil || *cfg.Default != false {
		t.Errorf("default = %#v, want false", cfg.Default)
	}
	if cfg.MD001.Enabled == nil || *cfg.MD001.Enabled != true {
		t.Errorf("MD001 = %#v, want true", cfg.MD001.Enabled)
	}
}

// TestReadConfigMissingFile verifies a missing file returns an error.
func TestReadConfigMissingFile(t *testing.T) {
	if _, err := ReadConfig(filepath.Join(t.TempDir(), "nope.json"), nil); err == nil {
		t.Error("expected error for missing config file")
	}
}
