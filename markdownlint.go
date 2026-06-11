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

// Package markdownlint is a Go port of the markdownlint library: a static
// analysis tool for Markdown/CommonMark files. It exposes a linting API, the
// built-in rule library, a fix engine, and configuration handling.
package markdownlint

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/engine"
	"github.com/ldmonster/go-markdownlint/internal/fix"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/rules"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

// Re-exported types.
type (
	// Configuration is the markdownlint configuration object.
	Configuration = types.Configuration
	// Error is a single lint finding.
	Error = types.Error
	// Results maps an input name to its lint errors.
	Results = types.Results
	// FixInfo describes an automatic fix.
	FixInfo = types.FixInfo
	// ErrorInfo is passed to a rule's OnError callback.
	ErrorInfo = types.ErrorInfo
	// OnError is the rule error-reporting callback.
	OnError = types.OnError
	// ConfigParser parses configuration content.
	ConfigParser = types.ConfigParser
	// Severity is an error severity.
	Severity = types.Severity
	// Rule is a lint rule descriptor.
	Rule = rule.Rule
	// RuleParams is the per-document parameter object passed to a rule.
	RuleParams = rule.RuleParams
	// ParserType identifies which parser a rule consumes.
	ParserType = types.ParserType
)

// Parser type constants.
const (
	ParserMicromark  = types.ParserMicromark
	ParserMarkdownIt = types.ParserMarkdownIt
	ParserNone       = types.ParserNone
)

// Severity constants.
const (
	SeverityError   = types.SeverityError
	SeverityWarning = types.SeverityWarning
)

// Options configures a Lint call.
type Options struct {
	Config             Configuration
	ConfigParsers      []ConfigParser
	CustomRules        []*Rule
	Files              []string
	Strings            map[string]string
	FrontMatter        *regexp.Regexp
	DisableFrontMatter bool
	HandleRuleFailures bool
	NoInlineConfig     bool
	Concurrency        int
	// ReadFile reads a file's content; defaults to os.ReadFile.
	ReadFile func(path string) (string, error)
}

// Lint runs the configured rules over the inputs in opts.
func Lint(ctx context.Context, opts Options) (Results, error) {
	return engine.Run(ctx, engine.Options{
		Config:             opts.Config,
		ConfigParsers:      opts.ConfigParsers,
		CustomRules:        opts.CustomRules,
		Files:              opts.Files,
		Strings:            opts.Strings,
		FrontMatter:        opts.FrontMatter,
		DisableFrontMatter: opts.DisableFrontMatter,
		HandleRuleFailures: opts.HandleRuleFailures,
		NoInlineConfig:     opts.NoInlineConfig,
		Concurrency:        opts.Concurrency,
		ReadFile:           opts.ReadFile,
	})
}

// LintString lints a single in-memory document.
func LintString(ctx context.Context, name, content string, cfg Configuration) ([]Error, error) {
	res, err := Lint(ctx, Options{Config: cfg, Strings: map[string]string{name: content}})
	if err != nil {
		return nil, err
	}

	return res[name], nil
}

// LintFiles lints the given files.
func LintFiles(ctx context.Context, files []string, cfg Configuration) (Results, error) {
	return Lint(ctx, Options{Config: cfg, Files: files})
}

// Version returns the library version.
func Version() string { return rules.Version }

// ConfigFromMap builds a Configuration from an untyped map (rule names/aliases,
// "default", "extends", custom-rule and tag keys), matching the shape accepted
// by markdownlint config files.
func ConfigFromMap(m map[string]interface{}) Configuration { return types.ConfigFromMap(m) }

// RuleInfo describes a built-in rule.
type RuleInfo struct {
	Names       []string
	Description string
	Tags        []string
	Fixable     bool
}

// Rules returns metadata for the built-in rules.
func Rules() []RuleInfo {
	fixable := map[string]bool{}
	for _, n := range rules.FixableRuleNames {
		fixable[n] = true
	}

	builtIn := rules.BuiltIn()

	out := make([]RuleInfo, 0, len(builtIn))
	for _, r := range builtIn {
		out = append(out, RuleInfo{
			Names:       r.Names,
			Description: r.Description,
			Tags:        r.Tags,
			Fixable:     fixable[strings.ToUpper(r.Names[0])],
		})
	}

	return out
}

// ApplyFix applies a single fix to a line. ok is false when the line is deleted.
func ApplyFix(line string, fi FixInfo, lineEnding string) (string, bool) {
	return fix.ApplyFix(line, fi, lineEnding)
}

// ApplyFixes applies as many of the given fixes as possible to input.
func ApplyFixes(input string, errors []Error) string {
	return fix.ApplyFixes(input, errors)
}

// String renders results in the default human format (sorted by name).
func ResultsString(r Results) string {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var b strings.Builder

	for _, file := range keys {
		for _, e := range r[file] {
			b.WriteString(file)
			b.WriteString(": ")
			b.WriteString(strconv.Itoa(e.LineNumber))
			b.WriteString(": ")
			b.WriteString(strings.Join(e.RuleNames, "/"))
			b.WriteString(" ")
			b.WriteString(e.RuleDescription)

			if e.ErrorDetail != "" {
				b.WriteString(" [" + e.ErrorDetail + "]")
			}

			if e.ErrorContext != "" {
				b.WriteString(" [Context: \"" + e.ErrorContext + "\"]")
			}

			b.WriteString("\n")
		}
	}

	return b.String()
}
