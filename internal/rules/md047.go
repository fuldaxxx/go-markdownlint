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
	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/rule"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func init() { register(&md047) }

var md047 = rule.Rule{
	Names:       []string{"MD047", "single-trailing-newline"},
	Description: "Files should end with a single newline character",
	Tags:        []string{"blank_lines"},
	Parser:      types.ParserNone,
	Fn: func(p *rule.RuleParams, onError types.OnError) {
		if len(p.Lines) == 0 {
			return
		}

		lastLineNumber := len(p.Lines)

		lastLine := p.Lines[lastLineNumber-1]
		if !helpers.IsBlankLine(lastLine) {
			n := runeLen(lastLine)
			helpers.AddError(onError, lastLineNumber, "", "", rng(n, 1), fixInsert(n+1, "\n"))
		}
	},
}
