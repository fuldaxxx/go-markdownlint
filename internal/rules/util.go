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

import "github.com/fuldaxxx/go-markdownlint/internal/types"

// rng builds a 1-based [column, length] range pointer.
func rng(column, length int) *[2]int { return &[2]int{column, length} }

// fixInsert builds a FixInfo that inserts text at editColumn.
func fixInsert(editColumn int, insertText string) *types.FixInfo {
	return &types.FixInfo{EditColumn: editColumn, InsertText: insertText}
}

// fixDelete builds a FixInfo that deletes deleteCount characters at editColumn.
func fixDelete(editColumn, deleteCount int) *types.FixInfo {
	return &types.FixInfo{EditColumn: editColumn, DeleteCount: deleteCount}
}

// fixReplace builds a FixInfo that deletes then inserts at editColumn.
func fixReplace(editColumn, deleteCount int, insertText string) *types.FixInfo {
	return &types.FixInfo{EditColumn: editColumn, DeleteCount: deleteCount, InsertText: insertText}
}

// runeLen returns the rune count of s (columns are rune-based).
func runeLen(s string) int { return len([]rune(s)) }
