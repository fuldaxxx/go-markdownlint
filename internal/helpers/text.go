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

package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Stringify stringifies a value for string concatenation (used by
// AddErrorDetailIf). Integers print without a decimal point; booleans print as
// true/false.
func Stringify(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case bool:
		if t {
			return "true"
		}

		return "false"
	case float64:
		if t == math.Trunc(t) && !math.IsInf(t, 0) {
			return strconv.FormatInt(int64(t), 10)
		}

		return strconv.FormatFloat(t, 'g', -1, 64)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

// Ellipsify shortens text to at most 30 characters, keeping the start and/or
// end as requested. Operates on runes.
func Ellipsify(text string, start, end bool) string {
	r := []rune(text)
	if len(r) <= 30 {
		return text
	}

	switch {
	case start && end:
		return string(r[:15]) + "..." + string(r[len(r)-15:])
	case end:
		return "..." + string(r[len(r)-30:])
	default:
		return string(r[:30]) + "..."
	}
}

// IsString reports whether v is a string.
func IsString(v interface{}) bool { _, ok := v.(string); return ok }

// IsEmptyString reports whether s is empty.
func IsEmptyString(s string) bool { return len(s) == 0 }

var (
	notCrLfRe        = regexp.MustCompile(`[^\r\n]`)
	notSpaceCrLfRe   = regexp.MustCompile(`[^ \r\n]`)
	trailingSpaceRe  = regexp.MustCompile(` +[\r\n]`)
	startsWithPipeRe = regexp.MustCompile(`^ *\|`)
)

// IsBlankLine reports whether a line is blank: empty, whitespace-only, or
// containing only HTML comments and ">" characters. Ported from helpers.cjs.
func IsBlankLine(line string) bool {
	const (
		startComment = "<!--"
		endComment   = "-->"
	)

	removeComments := func(s string) string {
		for {
			start := strings.Index(s, startComment)

			end := strings.Index(s, endComment)
			switch {
			case end != -1 && (start == -1 || end < start):
				s = s[end+len(endComment):]
			case start != -1 && end != -1:
				s = s[:start] + s[end+len(endComment):]
			case start != -1 && end == -1:
				s = s[:start]
			default:
				return s
			}
		}
	}

	if line == "" || strings.TrimSpace(line) == "" {
		return true
	}

	cleared := strings.ReplaceAll(removeComments(line), ">", "")

	return strings.TrimSpace(cleared) == ""
}

// ClearHTMLCommentText replaces the content of well-formed CommonMark comments
// with "." while preserving line/column information. Ported from helpers.cjs.
func ClearHTMLCommentText(text string) string {
	const (
		begin = "<!--"
		end   = "-->"
		safe  = "."
	)

	i := 0
	for {
		idx := strings.Index(text[i:], begin)
		if idx == -1 {
			break
		}

		i += idx

		j := strings.Index(text[i+2:], end)
		if j == -1 {
			// Un-terminated comments are treated as text.
			break
		}

		j += i + 2
		if j > i+len(begin) {
			content := text[i+len(begin) : j]
			lastLf := strings.LastIndex(text[:i], "\n") + 1
			preText := text[lastLf:i]
			isBlock := strings.TrimSpace(preText) == ""
			couldBeTable := startsWithPipeRe.MatchString(preText)
			spansTableCells := couldBeTable && strings.Contains(content, "\n")

			isValid := isBlock ||
				(!spansTableCells && !strings.HasPrefix(content, ">") && !strings.HasPrefix(content, "->") && !strings.HasSuffix(content, "-") && !strings.Contains(content, "--"))
			if isValid {
				cleared := notSpaceCrLfRe.ReplaceAllString(content, safe)
				cleared = trailingSpaceRe.ReplaceAllStringFunc(cleared, func(s string) string {
					return notCrLfRe.ReplaceAllString(s, safe)
				})
				text = text[:i+len(begin)] + cleared + text[j:]
			}
		}

		i = j + len(end)
	}

	return text
}

var escapeForRegExpRe = regexp.MustCompile(`[-/\\^$*+?.()|\[\]{}]`)

// EscapeForRegExp escapes a string for literal use in a regular expression.
func EscapeForRegExp(s string) string {
	return escapeForRegExpRe.ReplaceAllString(s, `\$0`)
}

// GetHTMLAttributeRe returns a regexp matching the given HTML attribute.
func GetHTMLAttributeRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\s` + regexp.QuoteMeta(name) + `\s*=\s*['"]?([^'"\s>]*)`)
}

// FrontMatterHasTitle reports whether the front matter lines include a title,
// using the given pattern (empty pattern disables the check; a custom pattern
// overrides the default).
func FrontMatterHasTitle(frontMatterLines []string, pattern string, patternSet bool) bool {
	ignore := patternSet && pattern == ""
	if ignore {
		return false
	}

	p := pattern
	if p == "" {
		p = `^\s*"?title"?\s*[:=]`
	}

	re := regexp.MustCompile(`(?i)` + p)
	for _, line := range frontMatterLines {
		if re.MatchString(line) {
			return true
		}
	}

	return false
}

// FileRange is a start/end line/column range (1-based), a subset of a token.
type FileRange struct {
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

func positionLessThanOrEqual(lineA, colA, lineB, colB int) bool {
	return lineA < lineB || (lineA == lineB && colA <= colB)
}

// HasOverlap reports whether two ranges overlap anywhere.
func HasOverlap(a, b FileRange) bool {
	lte := positionLessThanOrEqual(a.StartLine, a.StartColumn, b.StartLine, b.StartColumn)

	first, second := a, b
	if !lte {
		first, second = b, a
	}

	return positionLessThanOrEqual(
		second.StartLine,
		second.StartColumn,
		first.EndLine,
		first.EndColumn,
	)
}

// ExpandTildePath expands a leading ~ to the given home directory.
func ExpandTildePath(file, homedir string) string {
	if homedir == "" {
		return file
	}

	if file == "~" {
		return homedir
	}

	if strings.HasPrefix(file, "~/") || strings.HasPrefix(file, `~\`) {
		return homedir + file[1:]
	}

	return file
}
