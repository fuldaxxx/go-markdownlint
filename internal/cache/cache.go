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

// Package cache provides per-document memoization of token filtering and
// reference-link data. A Cache is request-scoped: one per document, used by a
// single goroutine, so no locking is required.
package cache

import (
	"strconv"
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
)

// Cache memoizes derived data for one parsed document.
type Cache struct {
	flat    []*mm.Token
	byTypes map[string][]*mm.Token
	refData *helpers.ReferenceLinkImageData
}

// New creates a cache over the document's flat token list.
func New(flat []*mm.Token) *Cache {
	return &Cache{flat: flat, byTypes: map[string][]*mm.Token{}}
}

// Tokens returns the flat token list.
func (c *Cache) Tokens() []*mm.Token { return c.flat }

func key(types []mm.TokenType, htmlFlow bool) string {
	return strings.Join(types, ",") + "|" + strconv.FormatBool(htmlFlow)
}

// FilterByTypes returns (and memoizes) the flat tokens of the given types.
func (c *Cache) FilterByTypes(types []mm.TokenType, htmlFlow bool) []*mm.Token {
	k := key(types, htmlFlow)
	if v, ok := c.byTypes[k]; ok {
		return v
	}

	v := mdhelpers.FilterFlat(c.flat, types, htmlFlow)
	c.byTypes[k] = v

	return v
}

// ReferenceLinkImageData returns (and memoizes) the reference-link data.
func (c *Cache) ReferenceLinkImageData() *helpers.ReferenceLinkImageData {
	if c.refData == nil {
		c.refData = helpers.GetReferenceLinkImageData(c.flat)
	}

	return c.refData
}
