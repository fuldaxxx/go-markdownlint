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
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/ldmonster/go-markdownlint/internal/cache"
	"github.com/ldmonster/go-markdownlint/internal/helpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/rules"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

// Options is the subset of options needed by the engine.
type Options struct {
	Config             types.Configuration
	ConfigParsers      []types.ConfigParser
	CustomRules        []*rule.Rule
	Files              []string
	Strings            map[string]string
	FrontMatter        *regexp.Regexp
	DisableFrontMatter bool
	HandleRuleFailures bool
	NoInlineConfig     bool
	Concurrency        int
	ReadFile           func(path string) (string, error)
}

func removeFrontMatter(content string, fm *regexp.Regexp) (string, []string) {
	if fm == nil {
		return content, nil
	}

	loc := fm.FindStringIndex(content)
	if loc == nil || loc[0] != 0 {
		return content, nil
	}

	matched := content[loc[0]:loc[1]]
	rest := content[loc[1]:]

	fmLines := helpers.SplitLines(matched)
	if len(fmLines) > 0 && fmLines[len(fmLines)-1] == "" {
		fmLines = fmLines[:len(fmLines)-1]
	}

	return rest, fmLines
}

// lintContent lints a single document and returns sorted errors.
func lintContent(
	ruleList []*rule.Rule,
	alias map[string][]string,
	name, content string,
	cfg map[string]interface{},
	configParsers []types.ConfigParser,
	fm *regexp.Regexp,
	handleRuleFailures, noInlineConfig bool,
) ([]types.Error, error) {
	content = strings.TrimPrefix(content, "\uFEFF")
	content, frontMatterLines := removeFrontMatter(content, fm)

	enabled := getEnabledRulesPerLineNumber(
		ruleList,
		helpers.SplitLines(content),
		frontMatterLines,
		noInlineConfig,
		cfg,
		configParsers,
		alias,
	)

	needMicromark := false

	for _, r := range enabled.enabledRuleList {
		if r.Parser == types.ParserMicromark || r.Parser == types.ParserMarkdownIt {
			needMicromark = true
		}
	}

	var (
		tokens []*mm.Token
		flat   []*mm.Token
	)

	if needMicromark {
		doc := mm.Parse(content)
		tokens = doc.Children
		flat = doc.Flat
	}

	preCleared := content
	_ = preCleared
	content = helpers.ClearHTMLCommentText(content)
	lines := helpers.SplitLines(content)

	c := cache.New(flat)

	// Build the fully-resolved typed configuration once; every rule reads its own
	// typed field from it.
	ecMap := make(map[string]interface{}, len(enabled.effectiveConfig))
	for name, opts := range enabled.effectiveConfig {
		ecMap[name] = opts
	}

	resolvedCfg := types.ConfigFromMap(ecMap)

	var results []types.Error

	for _, r := range enabled.enabledRuleList {
		ruleName := strings.ToUpper(r.Names[0])
		params := rule.NewRuleParams(
			name,
			rules.Version,
			lines,
			frontMatterLines,
			resolvedCfg,
			tokens,
			c,
		)

		onError := func(info types.ErrorInfo) {
			if info.LineNumber < 1 || info.LineNumber > len(lines) {
				return
			}

			absLine := info.LineNumber + len(frontMatterLines)
			if absLine < 0 || absLine >= len(enabled.enabledRulesPerLineNumber) {
				return
			}

			perLine := enabled.enabledRulesPerLineNumber[absLine]
			if perLine == nil || !perLine[ruleName] {
				return
			}

			var fixInfo *types.FixInfo

			if info.FixInfo != nil {
				fi := *info.FixInfo
				if fi.LineNumber != 0 {
					fi.LineNumber += len(frontMatterLines)
				}

				fixInfo = &fi
			}

			ruleInfo := ""
			if info.Information != nil {
				ruleInfo = info.Information.String()
			} else if r.Information != nil {
				ruleInfo = r.Information.String()
			}

			var rng *[2]int

			if info.Range != nil {
				v := *info.Range
				rng = &v
			}

			results = append(results, types.Error{
				LineNumber:      absLine,
				RuleNames:       append([]string(nil), r.Names...),
				RuleDescription: r.Description,
				RuleInformation: ruleInfo,
				ErrorDetail:     helpers.NewLineRe.ReplaceAllString(info.Detail, " "),
				ErrorContext:    helpers.NewLineRe.ReplaceAllString(info.Context, " "),
				ErrorRange:      rng,
				FixInfo:         fixInfo,
				Severity:        enabled.rulesSeverity[ruleName],
			})
		}

		var panicErr error

		func() {
			defer func() {
				if rec := recover(); rec != nil {
					if handleRuleFailures {
						onError(
							types.ErrorInfo{
								LineNumber: 1,
								Detail:     "This rule threw an exception: " + recoverMessage(rec),
							},
						)
					} else {
						panicErr = fmt.Errorf(
							"rule %s threw an exception: %s",
							ruleName,
							recoverMessage(rec),
						)
					}
				}
			}()

			r.Fn(params, onError)
		}()

		if panicErr != nil {
			return nil, panicErr
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RuleNames[0] != results[j].RuleNames[0] {
			return results[i].RuleNames[0] < results[j].RuleNames[0]
		}

		return results[i].LineNumber < results[j].LineNumber
	})

	return results, nil
}

func recoverMessage(rec interface{}) string {
	if err, ok := rec.(error); ok {
		return err.Error()
	}

	if s, ok := rec.(string); ok {
		return s
	}

	return "unknown error"
}

// Run lints all inputs in opts and returns the results.
func Run(ctx context.Context, opts Options) (types.Results, error) {
	ruleList := append([]*rule.Rule(nil), rules.BuiltIn()...)

	ruleList = append(ruleList, opts.CustomRules...)
	if err := validateRuleList(ruleList); err != nil {
		return nil, err
	}

	alias := mapAliasToRuleNames(ruleList)

	// Resolve config to its untyped map form once for the per-document linting
	// loop. An empty configuration leaves all rules enabled by default.
	cfg := opts.Config.ToMap()

	fm := opts.FrontMatter
	if fm == nil && !opts.DisableFrontMatter {
		fm = helpers.FrontMatterRe
	}

	readFile := opts.ReadFile
	if readFile == nil {
		readFile = func(p string) (string, error) {
			b, err := os.ReadFile(p)
			return string(b), err
		}
	}

	type job struct {
		name    string
		content string
		isFile  bool
	}

	var jobs []job
	for _, f := range opts.Files {
		jobs = append(jobs, job{name: f, isFile: true})
	}

	for name, content := range opts.Strings {
		jobs = append(jobs, job{name: name, content: content})
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 8
	}

	results := types.Results{}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	jobCh := make(chan job)
	errCh := make(chan error, 1)

	worker := func() {
		defer wg.Done()

		for j := range jobCh {
			if ctx.Err() != nil {
				return
			}

			content := j.content
			if j.isFile {
				c, err := readFile(j.name)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}

					return
				}

				content = c
			}

			errs, err := lintContent(
				ruleList,
				alias,
				j.name,
				content,
				cfg,
				opts.ConfigParsers,
				fm,
				opts.HandleRuleFailures,
				opts.NoInlineConfig,
			)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}

				return
			}

			mu.Lock()
			results[j.name] = errs
			mu.Unlock()
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go worker()
	}

	go func() {
		for _, j := range jobs {
			select {
			case <-ctx.Done():
			case jobCh <- j:
			}
		}

		close(jobCh)
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return results, nil
}
