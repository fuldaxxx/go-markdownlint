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

// GetPreferredLineEnding returns the most common line ending in input, falling
// back to defaultEOL (typically "\n") when there are none.
func GetPreferredLineEnding(input, defaultEOL string) string {
	var cr, lf, crlf int

	for _, m := range NewLineRe.FindAllString(input, -1) {
		switch m {
		case "\r":
			cr++
		case "\n":
			lf++
		case "\r\n":
			crlf++
		}
	}

	switch {
	case cr == 0 && lf == 0 && crlf == 0:
		if defaultEOL != "" {
			return defaultEOL
		}

		return "\n"
	case lf >= crlf && lf >= cr:
		return "\n"
	case crlf >= cr:
		return "\r\n"
	default:
		return "\r"
	}
}
