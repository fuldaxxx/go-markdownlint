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

// Package helpers ports helpers/helpers.cjs and helpers/shared.cjs: the
// rule-facing utilities, compiled regexes, and the reference-link data scan.
package helpers

import "regexp"

// NewLineRe matches newline characters (\r\n, \r, or \n). Mirrors
// shared.cjs newLineRe (/\r\n?|\n/g).
var NewLineRe = regexp.MustCompile(`\r\n?|\n`)

// NextLinesRe matches from the first newline to end of string.
var NextLinesRe = regexp.MustCompile(`[\r\n][\s\S]*$`)

// FrontMatterRe matches common YAML/TOML/JSON front matter at the start of a
// document. Ported from helpers.cjs (the `m` multiline flag is applied via the
// (?m) prefix; RE2 supports this).
var FrontMatterRe = regexp.MustCompile(
	`(?m)\A((^---[^\S\r\n\x{2028}\x{2029}]*$[\s\S]+?^---\s*)|(^\+\+\+[^\S\r\n\x{2028}\x{2029}]*$[\s\S]+?^(\+\+\+|\.\.\.)\s*)|(^\{[^\S\r\n\x{2028}\x{2029}]*$[\s\S]+?^\}\s*))(\r\n|\r|\n|$)`,
)

// InlineCommentStartRe matches the start of an inline markdownlint config
// comment. Case-insensitive; the trailing group ensures a separator follows.
var InlineCommentStartRe = regexp.MustCompile(
	`(?i)(<!--\s*markdownlint-(disable|enable|capture|restore|disable-file|enable-file|disable-line|disable-next-line|configure-file))(?:\s|-->)`,
)

// EndOfLineHTMLEntityRe matches an HTML entity at the end of a line.
var EndOfLineHTMLEntityRe = regexp.MustCompile(
	`&(?:#\d+|#[xX][\da-fA-F]+|[a-zA-Z]{2,31}|blk\d{2}|emsp1[34]|frac\d{2}|sup\d|there4);$`,
)

// EndOfLineGemojiCodeRe matches a GitHub emoji code at the end of a line.
var EndOfLineGemojiCodeRe = regexp.MustCompile(
	`:(?:[abmovx]|[-+]1|100|1234|(?:1st|2nd|3rd)_place_medal|8ball|clock\d{1,4}|e-mail|non-potable_water|o2|t-rex|u5272|u5408|u55b6|u6307|u6708|u6709|u6e80|u7121|u7533|u7981|u7a7a|[a-z]{2,15}2?|[a-z]{1,14}(?:_[a-z\d]{1,16})+):$`,
)

// AllPunctuation is every punctuation character (normal and full-width).
const AllPunctuation = ".,;:!?。，；：！？"

// AllPunctuationNoQuestion is AllPunctuation without question marks.
const AllPunctuationNoQuestion = ".,;:!。，；：！"

// SplitLines splits content into lines on any newline sequence (a trailing
// newline yields a trailing empty element).
func SplitLines(content string) []string {
	return NewLineRe.Split(content, -1)
}
