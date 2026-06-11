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

import "testing"

func TestGetPreferredLineEnding(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultEOL string
		want       string
	}{
		{"none uses default", "no newlines", "\r\n", "\r\n"},
		{"none empty default falls back to lf", "no newlines", "", "\n"},
		{"pure lf", "a\nb\nc", "", "\n"},
		{"pure crlf", "a\r\nb\r\nc", "", "\r\n"},
		{"pure cr", "a\rb\rc", "", "\r"},
		{"mixed lf majority", "a\nb\nc\r\n", "", "\n"},
		{"mixed crlf majority", "a\r\nb\r\nc\n", "", "\r\n"},
		{"mixed cr majority", "a\rb\rc\n", "", "\r"},
		{"single lf", "a\n", "", "\n"},
		{"single crlf", "a\r\n", "", "\r\n"},
		{"single cr", "a\r", "", "\r"},
		{"empty input uses default", "", "\r\n", "\r\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPreferredLineEnding(tt.input, tt.defaultEOL); got != tt.want {
				t.Errorf("GetPreferredLineEnding(%q, %q) = %q, want %q",
					tt.input, tt.defaultEOL, got, tt.want)
			}
		})
	}
}
