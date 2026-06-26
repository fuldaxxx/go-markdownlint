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

// Package rule defines the Rule descriptor and the RuleParams passed to a rule
// function. RuleParams exposes cache-backed token accessors so rules read the
// document consistently.
package rule

import (
	"net/url"

	"github.com/fuldaxxx/go-markdownlint/internal/cache"
	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

// Rule is a lint rule descriptor.
type Rule struct {
	Names        []string
	Description  string
	Tags         []string
	Parser       types.ParserType
	Information  *url.URL
	Asynchronous bool
	Fn           func(p *RuleParams, onError types.OnError)
}

// RuleParams is the per-document, per-rule parameter object.
type RuleParams struct {
	Name             string
	Version          string
	Lines            []string
	FrontMatterLines []string
	Config           types.Configuration // the resolved config; a rule reads its own typed field

	tokens []*mm.Token // micromark token tree (top-level)
	cache  *cache.Cache
}

// NewRuleParams builds a RuleParams. The cache is shared across all rules for
// one document.
func NewRuleParams(
	name, version string,
	lines, frontMatter []string,
	cfg types.Configuration,
	tokens []*mm.Token,
	c *cache.Cache,
) *RuleParams {
	return &RuleParams{
		Name:             name,
		Version:          version,
		Lines:            lines,
		FrontMatterLines: frontMatter,
		Config:           cfg,
		tokens:           tokens,
		cache:            c,
	}
}

// MicromarkTokens returns the top-level micromark tokens.
func (p *RuleParams) MicromarkTokens() []*mm.Token { return p.tokens }

// FilterByTypesCached returns the document's tokens of the given types (cached).
func (p *RuleParams) FilterByTypesCached(types []mm.TokenType, htmlFlow bool) []*mm.Token {
	return p.cache.FilterByTypes(types, htmlFlow)
}

// ReferenceLinkImageData returns the document's reference-link data (cached).
func (p *RuleParams) ReferenceLinkImageData() *helpers.ReferenceLinkImageData {
	return p.cache.ReferenceLinkImageData()
}
