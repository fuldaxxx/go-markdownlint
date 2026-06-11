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

// Package types holds the lowest-level shared types used across the linter:
// the error-reporting callback, fix information, lint results, and the
// configuration model. It has no dependencies on other internal packages so it
// can be imported everywhere without creating cycles.
package types

import "net/url"

// Severity is the severity of a lint error.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// FixInfo describes an automatic fix for a single line. A zero value means the
// corresponding field is unset, the semantics that normalizeFixInfo applies.
// DeleteCount == -1 deletes the whole line.
type FixInfo struct {
	LineNumber  int    // 1-based; 0 means "use the error's line number"
	EditColumn  int    // 1-based; 0 normalizes to 1
	DeleteCount int    // 0 normalizes to 0; -1 deletes the line
	InsertText  string // text to insert after deleting
}

// ErrorInfo is the argument passed to an OnError callback by a rule.
type ErrorInfo struct {
	LineNumber  int      // 1-based, relative to content (front matter excluded)
	Detail      string   // "" if absent
	Context     string   // "" if absent
	Information *url.URL // nil if absent
	Range       *[2]int  // [column, length], 1-based; nil if absent
	FixInfo     *FixInfo // nil if absent
}

// OnError is the error-reporting callback handed to each rule.
type OnError func(ErrorInfo)

// Error is a single lint finding (resultVersion 3 shape).
type Error struct {
	LineNumber      int      `json:"lineNumber"`
	RuleNames       []string `json:"ruleNames"`
	RuleDescription string   `json:"ruleDescription"`
	RuleInformation string   `json:"ruleInformation"`
	ErrorDetail     string   `json:"errorDetail"`
	ErrorContext    string   `json:"errorContext"`
	ErrorRange      *[2]int  `json:"errorRange"`
	FixInfo         *FixInfo `json:"fixInfo"`
	Severity        Severity `json:"severity"`
}

// Configuration is the typed markdownlint configuration object; see config.go.

// Results maps a file/string name to its lint errors.
type Results map[string][]Error

// ConfigParser parses configuration file content into a Configuration.
type ConfigParser func(content string) (Configuration, error)

// ParserType identifies which parser a rule consumes.
type ParserType int

const (
	ParserMicromark ParserType = iota
	ParserMarkdownIt
	ParserNone
)
