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
	"net/url"

	"github.com/ldmonster/go-markdownlint/internal/types"
)

// AddError reports a generic error via the onError callback.
func AddError(
	onError types.OnError,
	lineNumber int,
	detail, context string,
	rng *[2]int,
	fixInfo *types.FixInfo,
) {
	onError(types.ErrorInfo{
		LineNumber: lineNumber,
		Detail:     detail,
		Context:    context,
		Range:      rng,
		FixInfo:    fixInfo,
	})
}

// AddErrorWithInfo reports an error including an information URL.
func AddErrorWithInfo(
	onError types.OnError,
	lineNumber int,
	detail, context string,
	info *url.URL,
	rng *[2]int,
	fixInfo *types.FixInfo,
) {
	onError(types.ErrorInfo{
		LineNumber:  lineNumber,
		Detail:      detail,
		Context:     context,
		Information: info,
		Range:       rng,
		FixInfo:     fixInfo,
	})
}

// AddErrorDetailIf reports an error with an "Expected/Actual" detail when
// expected != actual. Values are stringified for string concatenation
// (see Stringify).
func AddErrorDetailIf(
	onError types.OnError,
	lineNumber int,
	expected, actual interface{},
	detail, context string,
	rng *[2]int,
	fixInfo *types.FixInfo,
) {
	if Stringify(expected) == Stringify(actual) {
		return
	}

	d := "Expected: " + Stringify(expected) + "; Actual: " + Stringify(actual)
	if detail != "" {
		d += "; " + detail
	}

	AddError(onError, lineNumber, d, context, rng, fixInfo)
}

// AddErrorContext reports an error with an ellipsified context string.
func AddErrorContext(
	onError types.OnError,
	lineNumber int,
	context string,
	start, end bool,
	rng *[2]int,
	fixInfo *types.FixInfo,
) {
	context = Ellipsify(NewLineRe.ReplaceAllString(context, "\n"), start, end)
	AddError(onError, lineNumber, "", context, rng, fixInfo)
}
