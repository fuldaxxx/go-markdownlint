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

// Package hujson provides a minimal standardizer that converts JSON with
// comments and trailing commas (a.k.a. JWCC/HuJSON) into standard JSON that the
// encoding/json package can parse. It implements only the Standardize entry
// point used by the configuration parsers; it is intentionally self-contained
// so the module can target older Go toolchains without external dependencies.
package hujson

// Standardize strips the HuJSON-specific extensions from b — line comments
// (// ... ), block comments (/* ... */), and trailing commas before a closing
// brace or bracket — returning standard JSON. Comment bytes are replaced with
// spaces so byte offsets and line numbers are preserved. The result is not
// validated as JSON; callers should unmarshal it to detect syntax errors.
func Standardize(b []byte) ([]byte, error) {
	out := stripComments(b)
	removeTrailingCommas(out)

	return out, nil
}

// stripComments returns a copy of b with // and /* */ comments replaced by
// spaces. Content inside JSON string literals is preserved verbatim.
func stripComments(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)

	inString := false

	for i := 0; i < len(out); i++ {
		c := out[i]
		if inString {
			if c == '\\' {
				i++ // skip the escaped byte
				continue
			}

			if c == '"' {
				inString = false
			}

			continue
		}

		switch c {
		case '"':
			inString = true
		case '/':
			if i+1 < len(out) && out[i+1] == '/' {
				// Line comment: blank out until end of line.
				for i < len(out) && out[i] != '\n' {
					out[i] = ' '
					i++
				}

				i-- // let the loop re-examine the newline (or end)
			} else if i+1 < len(out) && out[i+1] == '*' {
				// Block comment: blank out until the closing */, keeping newlines.
				out[i] = ' '
				out[i+1] = ' '

				i += 2
				for i < len(out) {
					if out[i] == '*' && i+1 < len(out) && out[i+1] == '/' {
						out[i] = ' '
						out[i+1] = ' '
						i++

						break
					}

					if out[i] != '\n' {
						out[i] = ' '
					}

					i++
				}
			}
		}
	}

	return out
}

// removeTrailingCommas rewrites b in place, replacing any comma that is
// followed (after whitespace) by a closing '}' or ']' with a space. It is
// string-aware so commas inside string literals are left untouched. b is
// expected to be comment-free (run stripComments first).
func removeTrailingCommas(b []byte) {
	inString := false

	for i := 0; i < len(b); i++ {
		c := b[i]
		if inString {
			if c == '\\' {
				i++
				continue
			}

			if c == '"' {
				inString = false
			}

			continue
		}

		switch c {
		case '"':
			inString = true
		case ',':
			j := i + 1
			for j < len(b) && isJSONSpace(b[j]) {
				j++
			}

			if j < len(b) && (b[j] == '}' || b[j] == ']') {
				b[i] = ' '
			}
		}
	}
}

func isJSONSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
