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
	"testing"

	"github.com/ldmonster/go-markdownlint/internal/types"
)

// capture returns an OnError callback and a pointer to the slice it appends to.
func capture() (types.OnError, *[]types.ErrorInfo) {
	var got []types.ErrorInfo
	onError := func(ei types.ErrorInfo) { got = append(got, ei) }
	return onError, &got
}

func TestAddError(t *testing.T) {
	onError, got := capture()
	rng := &[2]int{3, 4}
	fi := &types.FixInfo{LineNumber: 5, EditColumn: 2, DeleteCount: 1, InsertText: "x"}
	AddError(onError, 7, "detail", "context", rng, fi)

	if len(*got) != 1 {
		t.Fatalf("expected 1 error, got %d", len(*got))
	}
	e := (*got)[0]
	if e.LineNumber != 7 || e.Detail != "detail" || e.Context != "context" {
		t.Errorf("unexpected error fields: %+v", e)
	}
	if e.Range != rng {
		t.Errorf("range not propagated")
	}
	if e.FixInfo != fi {
		t.Errorf("fixInfo not propagated")
	}
	if e.Information != nil {
		t.Errorf("expected nil Information, got %v", e.Information)
	}
}

func TestAddErrorWithInfo(t *testing.T) {
	onError, got := capture()
	info, _ := url.Parse("https://example.com/rules/MD001")
	AddErrorWithInfo(onError, 2, "d", "c", info, nil, nil)

	if len(*got) != 1 {
		t.Fatalf("expected 1 error, got %d", len(*got))
	}
	e := (*got)[0]
	if e.LineNumber != 2 || e.Detail != "d" || e.Context != "c" {
		t.Errorf("unexpected fields: %+v", e)
	}
	if e.Information == nil || e.Information.String() != "https://example.com/rules/MD001" {
		t.Errorf("information not propagated: %v", e.Information)
	}
}

func TestAddErrorDetailIf(t *testing.T) {
	tests := []struct {
		name           string
		expected       interface{}
		actual         interface{}
		detail         string
		wantFired      bool
		wantDetailHead string
	}{
		{"equal ints suppress", 3, 3, "", false, ""},
		{"different ints fire", 3, 4, "", true, "Expected: 3; Actual: 4"},
		{"with extra detail", 1, 2, "more", true, "Expected: 1; Actual: 2; more"},
		{"equal strings suppress", "a", "a", "", false, ""},
		{"different strings fire", "a", "b", "", true, "Expected: a; Actual: b"},
		{"float equals int representation suppress", 3.0, 3, "", false, ""},
		{"bool difference fires", true, false, "", true, "Expected: true; Actual: false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onError, got := capture()
			AddErrorDetailIf(onError, 1, tt.expected, tt.actual, tt.detail, "ctx", nil, nil)
			if tt.wantFired {
				if len(*got) != 1 {
					t.Fatalf("expected error to fire, got %d", len(*got))
				}
				if (*got)[0].Detail != tt.wantDetailHead {
					t.Errorf("detail = %q, want %q", (*got)[0].Detail, tt.wantDetailHead)
				}
				if (*got)[0].Context != "ctx" {
					t.Errorf("context = %q, want ctx", (*got)[0].Context)
				}
			} else if len(*got) != 0 {
				t.Fatalf("expected no error, got %d", len(*got))
			}
		})
	}
}

func TestAddErrorContext(t *testing.T) {
	onError, got := capture()
	// Embedded newlines are normalized via NewLineRe to "\n".
	AddErrorContext(onError, 1, "abc\r\ndef", false, false, nil, nil)
	if len(*got) != 1 {
		t.Fatalf("expected 1 error, got %d", len(*got))
	}
	if (*got)[0].Context != "abc\ndef" {
		t.Errorf("context = %q, want %q", (*got)[0].Context, "abc\ndef")
	}
	if (*got)[0].Detail != "" {
		t.Errorf("detail should be empty, got %q", (*got)[0].Detail)
	}
}

func TestAddErrorContextEllipsified(t *testing.T) {
	onError, got := capture()
	long := "0123456789012345678901234567890123456789" // 40 chars
	AddErrorContext(onError, 1, long, true, false, nil, nil)
	if len(*got) != 1 {
		t.Fatalf("expected 1 error")
	}
	// start-only ellipsis: first 30 chars + "..."
	want := long[:30] + "..."
	if (*got)[0].Context != want {
		t.Errorf("context = %q, want %q", (*got)[0].Context, want)
	}
}
