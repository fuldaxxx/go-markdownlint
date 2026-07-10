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

// Package configparse provides configuration parsers (JSON, JSONC, YAML, TOML)
// and the Parse driver that tries parsers in order.
package configparse

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/fuldaxxx/go-markdownlint/internal/hujson"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

// Configuration is re-exported for convenience.
type Configuration = types.Configuration

// Parser parses configuration content into a Configuration.
type Parser = types.ConfigParser

// JSON parses strict JSON.
func JSON(content string) (Configuration, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return Configuration{}, err
	}

	return coerce(v), nil
}

// JSONC parses JSON with comments and trailing commas.
func JSONC(content string) (Configuration, error) {
	std, err := hujson.Standardize([]byte(content))
	if err != nil {
		return Configuration{}, err
	}

	return JSON(string(std))
}

// YAML parses YAML.
func YAML(content string) (Configuration, error) {
	var v interface{}
	if err := yaml.Unmarshal([]byte(content), &v); err != nil {
		return Configuration{}, err
	}

	return coerce(normalizeYAML(v)), nil
}

// TOML parses TOML.
func TOML(content string) (Configuration, error) {
	var v map[string]interface{}
	if _, err := toml.Decode(content, &v); err != nil {
		return Configuration{}, err
	}

	return coerce(v), nil
}

// Default returns the default parser list (JSON only).
func Default() []Parser { return []Parser{JSON} }

// Common returns JSON, JSONC, YAML, and TOML parsers (the set tools usually use).
func Common() []Parser { return []Parser{JSON, JSONC, YAML, TOML} }

// Parse tries each parser in order; the first success wins. On total failure it
// returns an aggregated error.
func Parse(name, content string, parsers []Parser) (Configuration, error) {
	if len(parsers) == 0 {
		parsers = Default()
	}

	var errs []string

	for i, parser := range parsers {
		cfg, err := parser(content)
		if err == nil {
			return cfg, nil
		}

		errs = append(errs, fmt.Sprintf("Parser %d: %s", i, err.Error()))
	}

	msg := append([]string{fmt.Sprintf("Unable to parse '%s'", name)}, errs...)

	return Configuration{}, fmt.Errorf("%s", strings.Join(msg, "; "))
}

// coerce maps a parsed value into a Configuration; a non-object result yields an
// empty configuration.
func coerce(v interface{}) Configuration {
	if m, ok := v.(map[string]interface{}); ok {
		return types.ConfigFromMap(m)
	}

	return Configuration{}
}

// normalizeYAML converts map[interface{}]interface{} (older yaml shapes) to
// map[string]any recursively.
func normalizeYAML(v interface{}) interface{} {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, val := range t {
			m[fmt.Sprint(k)] = normalizeYAML(val)
		}

		return m
	case map[string]interface{}:
		for k, val := range t {
			t[k] = normalizeYAML(val)
		}

		return t
	case []interface{}:
		for i, val := range t {
			t[i] = normalizeYAML(val)
		}

		return t
	default:
		return v
	}
}
