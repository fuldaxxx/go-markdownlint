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
	"fmt"
	"os"
	"path/filepath"

	"github.com/fuldaxxx/go-markdownlint/configparse"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

// ReadConfig reads and parses a configuration file, resolving any "extends"
// references relative to the file's directory. If parsers is nil, the common
// parser set (JSON, JSONC, YAML, TOML) is used. The merge order is
// parent-then-child (child overrides).
func ReadConfig(file string, parsers []ConfigParser) (Configuration, error) {
	return readConfig(file, parsers, map[string]bool{})
}

func readConfig(file string, parsers []ConfigParser, seen map[string]bool) (Configuration, error) {
	if parsers == nil {
		parsers = configparse.Common()
	}

	abs, err := filepath.Abs(file)
	if err != nil {
		abs = file
	}

	if seen[abs] {
		return Configuration{}, fmt.Errorf("circular configuration extends at %q", file)
	}

	seen[abs] = true

	content, err := os.ReadFile(file)
	if err != nil {
		return Configuration{}, err
	}

	cfg, err := configparse.Parse(file, string(content), parsers)
	if err != nil {
		return Configuration{}, err
	}

	exts := cfg.Extends
	cfg.Extends = nil

	child := cfg.ToMap()
	delete(child, "extends")

	if len(exts) == 0 {
		return types.ConfigFromMap(child), nil
	}

	// Merge each parent in order, then overlay the child (child overrides).
	merged := map[string]interface{}{}

	for _, ext := range exts {
		resolved := ext
		if !filepath.IsAbs(ext) {
			resolved = filepath.Join(filepath.Dir(file), ext)
		}

		parent, err := readConfig(resolved, parsers, seen)
		if err != nil {
			return Configuration{}, err
		}

		for k, v := range parent.ToMap() {
			merged[k] = v
		}
	}

	for k, v := range child {
		merged[k] = v
	}

	return types.ConfigFromMap(merged), nil
}
