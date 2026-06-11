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
	"reflect"
	"strings"
)

// Configuration is the typed markdownlint configuration object. It has explicit
// fields for the well-known keys ("default", "extends") and a typed config
// struct for each built-in rule, plus a Custom map for the irreducibly dynamic
// entries: custom-rule names and tag keys (which configure a group of rules at
// once).
//
// Build one directly with struct fields, or from an untyped map (as produced by
// the JSON/YAML/TOML parsers) with ConfigFromMap.
type Configuration struct {
	// Default toggles all rules on/off ("default" key). nil means unset.
	Default *bool
	// Extends lists parent configuration paths ("extends" key).
	Extends []string

	MD001 MD001Config `mdlint:"MD001,heading-increment"`
	MD003 MD003Config `mdlint:"MD003,heading-style"`
	MD004 MD004Config `mdlint:"MD004,ul-style"`
	MD005 RuleConfig  `mdlint:"MD005,list-indent"`
	MD007 MD007Config `mdlint:"MD007,ul-indent"`
	MD009 MD009Config `mdlint:"MD009,no-trailing-spaces"`
	MD010 MD010Config `mdlint:"MD010,no-hard-tabs"`
	MD011 RuleConfig  `mdlint:"MD011,no-reversed-links"`
	MD012 MD012Config `mdlint:"MD012,no-multiple-blanks"`
	MD013 MD013Config `mdlint:"MD013,line-length"`
	MD014 RuleConfig  `mdlint:"MD014,commands-show-output"`
	MD018 RuleConfig  `mdlint:"MD018,no-missing-space-atx"`
	MD019 RuleConfig  `mdlint:"MD019,no-multiple-space-atx"`
	MD020 RuleConfig  `mdlint:"MD020,no-missing-space-closed-atx"`
	MD021 RuleConfig  `mdlint:"MD021,no-multiple-space-closed-atx"`
	MD022 MD022Config `mdlint:"MD022,blanks-around-headings"`
	MD023 RuleConfig  `mdlint:"MD023,heading-start-left"`
	MD024 MD024Config `mdlint:"MD024,no-duplicate-heading"`
	MD025 MD025Config `mdlint:"MD025,single-title,single-h1"`
	MD026 MD026Config `mdlint:"MD026,no-trailing-punctuation"`
	MD027 MD027Config `mdlint:"MD027,no-multiple-space-blockquote"`
	MD028 RuleConfig  `mdlint:"MD028,no-blanks-blockquote"`
	MD029 MD029Config `mdlint:"MD029,ol-prefix"`
	MD030 MD030Config `mdlint:"MD030,list-marker-space"`
	MD031 MD031Config `mdlint:"MD031,blanks-around-fences"`
	MD032 RuleConfig  `mdlint:"MD032,blanks-around-lists"`
	MD033 MD033Config `mdlint:"MD033,no-inline-html"`
	MD034 RuleConfig  `mdlint:"MD034,no-bare-urls"`
	MD035 MD035Config `mdlint:"MD035,hr-style"`
	MD036 MD036Config `mdlint:"MD036,no-emphasis-as-heading"`
	MD037 RuleConfig  `mdlint:"MD037,no-space-in-emphasis"`
	MD038 RuleConfig  `mdlint:"MD038,no-space-in-code"`
	MD039 RuleConfig  `mdlint:"MD039,no-space-in-links"`
	MD040 MD040Config `mdlint:"MD040,fenced-code-language"`
	MD041 MD041Config `mdlint:"MD041,first-line-heading,first-line-h1"`
	MD042 RuleConfig  `mdlint:"MD042,no-empty-links"`
	MD043 MD043Config `mdlint:"MD043,required-headings"`
	MD044 MD044Config `mdlint:"MD044,proper-names"`
	MD045 RuleConfig  `mdlint:"MD045,no-alt-text"`
	MD046 MD046Config `mdlint:"MD046,code-block-style"`
	MD047 RuleConfig  `mdlint:"MD047,single-trailing-newline"`
	MD048 MD048Config `mdlint:"MD048,code-fence-style"`
	MD049 MD049Config `mdlint:"MD049,emphasis-style"`
	MD050 MD050Config `mdlint:"MD050,strong-style"`
	MD051 MD051Config `mdlint:"MD051,link-fragments"`
	MD052 MD052Config `mdlint:"MD052,reference-links-images"`
	MD053 MD053Config `mdlint:"MD053,link-image-reference-definitions"`
	MD054 MD054Config `mdlint:"MD054,link-image-style"`
	MD055 MD055Config `mdlint:"MD055,table-pipe-style"`
	MD056 RuleConfig  `mdlint:"MD056,table-column-count"`
	MD058 RuleConfig  `mdlint:"MD058,blanks-around-tables"`
	MD059 MD059Config `mdlint:"MD059,descriptive-link-text"`
	MD060 MD060Config `mdlint:"MD060,table-column-style"`

	// Custom holds entries that aren't built-in rule names/aliases: custom-rule
	// names and tag keys (which enable/disable a whole group of rules).
	Custom map[string]RuleConfig
}

// ruleMeta is embedded in every rule config; its exported fields are promoted
// into the rule's JSON object ("enabled"/"severity").
type ruleMeta struct {
	Enabled  *bool  `json:"enabled,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// RuleConfig is the generic rule entry used for rules without typed options
// (and for Custom entries). Options carries any remaining option keys.
type RuleConfig struct {
	ruleMeta
	Options map[string]interface{} `json:"-"`
}

// Per-rule typed config structs. Pointer fields distinguish "unset" (nil, use
// the rule's default) from an explicit value; slice/interface fields use nil for
// unset. JSON tags match the option keys used in configuration files.

type MD001Config struct {
	ruleMeta
	FrontMatterTitle *string `json:"front_matter_title,omitempty"`
}

type MD003Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD004Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD007Config struct {
	ruleMeta
	Indent        *int  `json:"indent,omitempty"`
	StartIndented *bool `json:"start_indented,omitempty"`
	StartIndent   *int  `json:"start_indent,omitempty"`
}

type MD009Config struct {
	ruleMeta
	BrSpaces           *int  `json:"br_spaces,omitempty"`
	CodeBlocks         *bool `json:"code_blocks,omitempty"`
	ListItemEmptyLines *bool `json:"list_item_empty_lines,omitempty"`
	Strict             *bool `json:"strict,omitempty"`
}

type MD010Config struct {
	ruleMeta
	CodeBlocks          *bool    `json:"code_blocks,omitempty"`
	SpacesPerTab        *int     `json:"spaces_per_tab,omitempty"`
	IgnoreCodeLanguages []string `json:"ignore_code_languages,omitempty"`
}

type MD012Config struct {
	ruleMeta
	Maximum *int `json:"maximum,omitempty"`
}

type MD013Config struct {
	ruleMeta
	LineLength          *int  `json:"line_length,omitempty"`
	HeadingLineLength   *int  `json:"heading_line_length,omitempty"`
	CodeBlockLineLength *int  `json:"code_block_line_length,omitempty"`
	Strict              *bool `json:"strict,omitempty"`
	Stern               *bool `json:"stern,omitempty"`
	CodeBlocks          *bool `json:"code_blocks,omitempty"`
	Tables              *bool `json:"tables,omitempty"`
	Headings            *bool `json:"headings,omitempty"`
}

type MD022Config struct {
	ruleMeta
	LinesAbove interface{} `json:"lines_above,omitempty"`
	LinesBelow interface{} `json:"lines_below,omitempty"`
}

type MD024Config struct {
	ruleMeta
	SiblingsOnly *bool `json:"siblings_only,omitempty"`
}

type MD025Config struct {
	ruleMeta
	Level            *int    `json:"level,omitempty"`
	FrontMatterTitle *string `json:"front_matter_title,omitempty"`
}

type MD026Config struct {
	ruleMeta
	Punctuation interface{} `json:"punctuation,omitempty"`
}

type MD027Config struct {
	ruleMeta
	ListItems *bool `json:"list_items,omitempty"`
}

type MD029Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD030Config struct {
	ruleMeta
	UlSingle *int `json:"ul_single,omitempty"`
	OlSingle *int `json:"ol_single,omitempty"`
	UlMulti  *int `json:"ul_multi,omitempty"`
	OlMulti  *int `json:"ol_multi,omitempty"`
}

type MD031Config struct {
	ruleMeta
	ListItems *bool `json:"list_items,omitempty"`
}

type MD033Config struct {
	ruleMeta
	AllowedElements      []string  `json:"allowed_elements,omitempty"`
	TableAllowedElements *[]string `json:"table_allowed_elements,omitempty"`
}

type MD035Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD036Config struct {
	ruleMeta
	Punctuation interface{} `json:"punctuation,omitempty"`
}

type MD040Config struct {
	ruleMeta
	AllowedLanguages []string `json:"allowed_languages,omitempty"`
	LanguageOnly     *bool    `json:"language_only,omitempty"`
}

type MD041Config struct {
	ruleMeta
	AllowPreamble    *bool   `json:"allow_preamble,omitempty"`
	Level            *int    `json:"level,omitempty"`
	FrontMatterTitle *string `json:"front_matter_title,omitempty"`
}

type MD043Config struct {
	ruleMeta
	Headings  interface{} `json:"headings,omitempty"`
	MatchCase *bool       `json:"match_case,omitempty"`
}

type MD044Config struct {
	ruleMeta
	Names        []string `json:"names,omitempty"`
	CodeBlocks   *bool    `json:"code_blocks,omitempty"`
	HTMLElements *bool    `json:"html_elements,omitempty"`
}

type MD046Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD048Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD049Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD050Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD051Config struct {
	ruleMeta
	IgnoreCase     *bool   `json:"ignore_case,omitempty"`
	IgnoredPattern *string `json:"ignored_pattern,omitempty"`
}

type MD052Config struct {
	ruleMeta
	ShortcutSyntax *bool    `json:"shortcut_syntax,omitempty"`
	IgnoredLabels  []string `json:"ignored_labels,omitempty"`
}

type MD053Config struct {
	ruleMeta
	IgnoredDefinitions []string `json:"ignored_definitions,omitempty"`
}

type MD054Config struct {
	ruleMeta
	Autolink  *bool `json:"autolink,omitempty"`
	Inline    *bool `json:"inline,omitempty"`
	Full      *bool `json:"full,omitempty"`
	Collapsed *bool `json:"collapsed,omitempty"`
	Shortcut  *bool `json:"shortcut,omitempty"`
	URLInline *bool `json:"url_inline,omitempty"`
}

type MD055Config struct {
	ruleMeta
	Style *string `json:"style,omitempty"`
}

type MD059Config struct {
	ruleMeta
	ProhibitedTexts *[]string `json:"prohibited_texts,omitempty"`
}

type MD060Config struct {
	ruleMeta
	Style            *string `json:"style,omitempty"`
	AlignedDelimiter *bool   `json:"aligned_delimiter,omitempty"`
}

// --- Conversion between the untyped map form and the typed struct ---

// aliasField maps an upper-cased rule name or alias to the index of its struct
// field in Configuration. Built once from the `mdlint` struct tags.
var aliasField = func() map[string]int {
	m := map[string]int{}

	t := reflect.TypeOf(Configuration{})
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("mdlint")
		if tag == "" {
			continue
		}

		for _, name := range strings.Split(tag, ",") {
			m[strings.ToUpper(name)] = i
		}
	}

	return m
}()

// canonicalName returns the canonical rule name (the first `mdlint` tag token)
// for the Configuration field at index i, or "".
func canonicalName(i int) string {
	tag := reflect.TypeOf(Configuration{}).Field(i).Tag.Get("mdlint")
	if tag == "" {
		return ""
	}

	return strings.SplitN(tag, ",", 2)[0]
}

// ConfigFromMap builds a Configuration from an untyped map as produced by the
// configuration parsers. Keys are matched case-insensitively. Built-in rule
// names/aliases populate the corresponding typed field; "default"/"extends" set
// those fields; anything else (custom-rule names, tag keys) goes into Custom.
func ConfigFromMap(m map[string]interface{}) Configuration {
	var cfg Configuration

	v := reflect.ValueOf(&cfg).Elem()

	for key, value := range m {
		switch strings.ToUpper(key) {
		case "DEFAULT":
			b := truthyValue(value)
			cfg.Default = &b

			continue
		case "EXTENDS":
			cfg.Extends = toStringSlice(value)
			continue
		}

		if idx, ok := aliasField[strings.ToUpper(key)]; ok {
			field := v.Field(idx)
			// Generic RuleConfig fields keep their raw options; typed fields are
			// populated via a JSON round-trip using their option tags.
			if rc, isRC := field.Addr().Interface().(*RuleConfig); isRC {
				*rc = ruleConfigFromValue(value)
			} else {
				assignTypedField(field, value)
			}

			continue
		}

		if cfg.Custom == nil {
			cfg.Custom = map[string]RuleConfig{}
		}

		cfg.Custom[key] = ruleConfigFromValue(value)
	}

	return cfg
}

// assignTypedField populates a typed rule-config struct field from a config
// value, which may be a bool (enable/disable) or an options object.
func assignTypedField(field reflect.Value, value interface{}) {
	obj := normalizeRuleValue(value)

	b, err := json.Marshal(obj)
	if err != nil {
		return
	}

	_ = json.Unmarshal(b, field.Addr().Interface())
}

// normalizeRuleValue turns a config value into an options object: a bool becomes
// {"enabled": bool}; a map is returned (with interface keys coerced to strings);
// anything else becomes {"enabled": truthy}.
func normalizeRuleValue(value interface{}) map[string]interface{} {
	switch t := value.(type) {
	case map[string]interface{}:
		return t
	case Configuration:
		return t.ToMap()
	case bool:
		return map[string]interface{}{"enabled": t}
	default:
		return map[string]interface{}{"enabled": truthyValue(value)}
	}
}

// ruleConfigFromValue builds a generic RuleConfig (used for Custom entries and
// untyped rules) from a config value.
func ruleConfigFromValue(value interface{}) RuleConfig {
	var rc RuleConfig

	obj := normalizeRuleValue(value)
	rc.Options = map[string]interface{}{}

	for k, val := range obj {
		switch k {
		case "enabled":
			b := truthyValue(val)
			rc.Enabled = &b
		case "severity":
			if s, ok := val.(string); ok {
				rc.Severity = s
			}
		default:
			rc.Options[k] = val
		}
	}

	if len(rc.Options) == 0 {
		rc.Options = nil
	}

	return rc
}

// toMap converts a Configuration back into the untyped map form consumed by the
// engine's resolution logic (which still operates on maps to preserve exact
// tag/alias/severity semantics).
func (c Configuration) ToMap() map[string]interface{} {
	out := map[string]interface{}{}
	if c.Default != nil {
		out["default"] = *c.Default
	}

	if len(c.Extends) > 0 {
		out["extends"] = c.Extends
	}

	v := reflect.ValueOf(c)
	for i := 0; i < v.NumField(); i++ {
		name := canonicalName(i)
		if name == "" {
			continue
		}

		field := v.Field(i)
		if rc, ok := field.Interface().(RuleConfig); ok {
			if m := rc.toValue(); m != nil {
				out[name] = m
			}

			continue
		}

		b, err := json.Marshal(field.Interface())
		if err != nil {
			continue
		}

		obj := map[string]interface{}{}
		if err := json.Unmarshal(b, &obj); err != nil {
			continue
		}

		if len(obj) > 0 {
			out[name] = obj
		}
	}

	for name, rc := range c.Custom {
		if m := rc.toValue(); m != nil {
			out[name] = m
		}
	}

	return out
}

// toValue renders a RuleConfig back to its map form, or nil when it carries no
// information (so the key is omitted entirely).
func (rc RuleConfig) toValue() map[string]interface{} {
	if rc.Enabled == nil && rc.Severity == "" && len(rc.Options) == 0 {
		return nil
	}

	m := map[string]interface{}{}
	if rc.Enabled != nil {
		m["enabled"] = *rc.Enabled
	}

	if rc.Severity != "" {
		m["severity"] = rc.Severity
	}

	for k, v := range rc.Options {
		m[k] = v
	}

	return m
}

func truthyValue(v interface{}) bool {
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

func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}

		return []string{t}
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}

		return out
	default:
		return nil
	}
}

// --- Value-or-default helpers used by rule implementations ---

// BoolOr returns *p when p is non-nil, otherwise def.
func BoolOr(p *bool, def bool) bool {
	if p != nil {
		return *p
	}

	return def
}

// IntOr returns *p when p is non-nil, otherwise def.
func IntOr(p *int, def int) int {
	if p != nil {
		return *p
	}

	return def
}

// StringOr returns *p when p is non-nil, otherwise def.
func StringOr(p *string, def string) string {
	if p != nil {
		return *p
	}

	return def
}
