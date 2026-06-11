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

package rules

import (
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

var md043 = rule.Rule{
	Names:       []string{"MD043", "required-headings"},
	Description: "Required heading structure",
	Tags:        []string{"headings"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		c := p.Config.MD043
		if c.Headings == nil {
			return
		}

		requiredHeadings, isArray := md043AsStringArray(c.Headings)
		if !isArray {
			return
		}

		matchCase := types.BoolOr(c.MatchCase, false)
		i := 0
		matchAny := false
		hasError := false
		anyHeadings := false
		getExpected := func() string {
			var s string
			if i < len(requiredHeadings) && requiredHeadings[i] != "" {
				s = requiredHeadings[i]
			} else {
				s = "[None]"
			}

			i++

			return s
		}
		handleCase := func(str string) string {
			if matchCase {
				return str
			}

			return strings.ToLower(str)
		}

		for _, heading := range p.FilterByTypesCached([]mm.TokenType{mm.TypeAtxHeading, mm.TypeSetextHeading}, false) {
			if hasError {
				continue
			}

			headingText := mdhelpers.GetHeadingText(heading)
			headingLevel := mdhelpers.GetHeadingLevel(heading)
			anyHeadings = true
			actual := strings.Repeat("#", headingLevel) + " " + headingText

			expected := getExpected()
			switch {
			case expected == "*":
				nextExpected := getExpected()
				if handleCase(nextExpected) != handleCase(actual) {
					matchAny = true
					i--
				}
			case expected == "+":
				matchAny = true
			case expected == "?":
				// Allow current, match next
			case handleCase(expected) == handleCase(actual):
				matchAny = false
			case matchAny:
				i--
			default:
				helpers.AddErrorDetailIf(
					onError,
					heading.StartLine,
					expected,
					actual,
					"",
					"",
					nil,
					nil,
				)

				hasError = true
			}
		}

		extraHeadings := len(requiredHeadings) - i
		allStar := true

		for _, h := range requiredHeadings {
			if h != "*" {
				allStar = false
				break
			}
		}

		nextIsStar := i < len(requiredHeadings) && requiredHeadings[i] == "*"
		if !hasError &&
			(extraHeadings > 1 || (extraHeadings == 1 && !nextIsStar)) &&
			(anyHeadings || !allStar) {
			ctx := ""
			if i < len(requiredHeadings) {
				ctx = requiredHeadings[i]
			}

			helpers.AddErrorContext(
				onError,
				len(p.Lines),
				ctx,
				false,
				false,
				nil,
				nil,
			)
		}
	},
}

func md043AsStringArray(v interface{}) ([]string, bool) {
	switch t := v.(type) {
	case []string:
		return t, true
	case []interface{}:
		out := make([]string, len(t))
		for idx, e := range t {
			out[idx] = helpers.Stringify(e)
		}

		return out, true
	default:
		return nil, false
	}
}

func init() { register(&md043) }
