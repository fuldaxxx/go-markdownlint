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

// Package fix implements applyFix/applyFixes.
package fix

import (
	"sort"
	"strings"

	"github.com/fuldaxxx/go-markdownlint/internal/helpers"
	"github.com/fuldaxxx/go-markdownlint/internal/types"
)

func normalize(fi types.FixInfo, lineNumber int) types.FixInfo {
	out := types.FixInfo{
		LineNumber:  fi.LineNumber,
		EditColumn:  fi.EditColumn,
		DeleteCount: fi.DeleteCount,
		InsertText:  fi.InsertText,
	}
	if out.LineNumber == 0 {
		out.LineNumber = lineNumber
	}

	if out.EditColumn == 0 {
		out.EditColumn = 1
	}

	return out
}

// ApplyFix applies a single fix to a line. ok is false when the line is
// deleted (DeleteCount == -1). Columns index by rune to match the parser's
// character-based columns.
func ApplyFix(line string, fi types.FixInfo, lineEnding string) (string, bool) {
	n := normalize(fi, 0)

	editIndex := n.EditColumn - 1
	if n.DeleteCount == -1 {
		return "", false
	}

	r := []rune(line)
	if editIndex > len(r) {
		editIndex = len(r)
	}

	endIndex := editIndex + n.DeleteCount
	if endIndex > len(r) {
		endIndex = len(r)
	}

	insert := strings.ReplaceAll(n.InsertText, "\n", lineEnding)

	return string(r[:editIndex]) + insert + string(r[endIndex:]), true
}

// ApplyFixes applies as many of the given fixes as possible to input.
func ApplyFixes(input string, errors []types.Error) string {
	lineEnding := helpers.GetPreferredLineEnding(input, "\n")
	lines := helpers.SplitLines(input)

	var fixInfos []types.FixInfo

	for _, e := range errors {
		if e.FixInfo != nil {
			fixInfos = append(fixInfos, normalize(*e.FixInfo, e.LineNumber))
		}
	}

	// Sort: bottom-to-top, line-deletes last, right-to-left, longest-insert-first.
	sort.SliceStable(fixInfos, func(i, j int) bool {
		a, b := fixInfos[i], fixInfos[j]
		if a.LineNumber != b.LineNumber {
			return a.LineNumber > b.LineNumber
		}

		aDel := a.DeleteCount == -1

		bDel := b.DeleteCount == -1
		if aDel != bDel {
			// a deleting => a sorts after b (returns false); b deleting => a before.
			return bDel
		}

		if a.EditColumn != b.EditColumn {
			return a.EditColumn > b.EditColumn
		}

		return len([]rune(a.InsertText)) > len([]rune(b.InsertText))
	})

	// Remove duplicate entries.
	deduped := fixInfos[:0:0]

	var last *types.FixInfo

	for i := range fixInfos {
		fi := fixInfos[i]
		if last == nil || fi.LineNumber != last.LineNumber || fi.EditColumn != last.EditColumn ||
			fi.DeleteCount != last.DeleteCount || fi.InsertText != last.InsertText {
			deduped = append(deduped, fi)
			last = &deduped[len(deduped)-1]
		}
	}

	fixInfos = deduped

	// Collapse insert/no-delete followed by no-insert/delete at same position.
	prev := types.FixInfo{LineNumber: -1}

	for i := range fixInfos {
		fi := &fixInfos[i]
		if fi.LineNumber == prev.LineNumber && fi.EditColumn == prev.EditColumn &&
			fi.InsertText == "" && fi.DeleteCount > 0 &&
			prev.InsertText != "" && prev.DeleteCount == 0 {
			fi.InsertText = prev.InsertText
			// Zero out the previous entry.
			for j := i - 1; j >= 0; j-- {
				if fixInfos[j].LineNumber == prev.LineNumber &&
					fixInfos[j].EditColumn == prev.EditColumn &&
					fixInfos[j].InsertText == prev.InsertText &&
					fixInfos[j].DeleteCount == 0 {
					fixInfos[j].LineNumber = 0
					break
				}
			}
		}

		prev = *fi
	}

	collapsed := fixInfos[:0:0]
	for _, fi := range fixInfos {
		if fi.LineNumber != 0 {
			collapsed = append(collapsed, fi)
		}
	}

	fixInfos = collapsed

	// Apply non-overlapping fixes.
	deleted := make([]bool, len(lines))
	lastLineIndex := -1
	lastEditIndex := -1

	for _, fi := range fixInfos {
		lineIndex := fi.LineNumber - 1
		editIndex := fi.EditColumn - 1

		if lineIndex < 0 || lineIndex >= len(lines) {
			continue
		}

		overlapOK := lineIndex != lastLineIndex || fi.DeleteCount == -1
		if !overlapOK {
			adj := 1
			if fi.DeleteCount > 0 {
				adj = 0
			}

			overlapOK = (editIndex + fi.DeleteCount) <= (lastEditIndex - adj)
		}

		if overlapOK {
			fixed, ok := ApplyFix(lines[lineIndex], fi, lineEnding)
			if ok {
				lines[lineIndex] = fixed
			} else {
				deleted[lineIndex] = true
			}
		}

		lastLineIndex = lineIndex
		lastEditIndex = editIndex
	}

	var out []string

	for i, l := range lines {
		if !deleted[i] {
			out = append(out, l)
		}
	}

	return strings.Join(out, lineEnding)
}
