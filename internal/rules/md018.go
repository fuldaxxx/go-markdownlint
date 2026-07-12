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
	"regexp"
	"strings"

	"github.com/ldmonster/go-markdownlint/internal/helpers"
	"github.com/ldmonster/go-markdownlint/internal/mdhelpers"
	mm "github.com/ldmonster/go-markdownlint/internal/micromark"
	"github.com/ldmonster/go-markdownlint/internal/rule"
	"github.com/ldmonster/go-markdownlint/internal/types"
)

func init() { register(&md018) }

var (
	md018MissingRe = regexp.MustCompile(`^#+[^# \t]`)
	md018EmptyRe   = regexp.MustCompile(`#\s*$`)
	md018LeadRe    = regexp.MustCompile(`^#+`)
)

var md018 = rule.Rule{
	Names:       []string{"MD018", "no-missing-space-atx"},
	Description: "No space after hash on atx style heading",
	Tags:        []string{"headings", "atx", "spaces"},
	Parser:      types.ParserMicromark,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		ignore := map[int]struct{}{}
		for _, b := range p.FilterByTypesCached([]mm.TokenType{mm.TypeCodeFenced, mm.TypeCodeIndented, mm.TypeHTMLFlow}, false) {
			mdhelpers.AddRangeToSet(ignore, b.StartLine, b.EndLine)
		}

		for lineIndex, line := range p.Lines {
			if _, skip := ignore[lineIndex+1]; skip {
				continue
			}

			if md018MissingRe.MatchString(line) && !md018EmptyRe.MatchString(line) &&
				!strings.HasPrefix(line, "#️⃣") {
				hashCount := len(md018LeadRe.FindString(line))
				helpers.AddErrorContext(
					onError,
					lineIndex+1,
					strings.TrimSpace(line),
					false,
					false,
					rng(1, hashCount+1),
					fixInsert(hashCount+1, " "),
				)
			}
		}
	},
}
