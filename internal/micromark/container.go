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

package micromark

import (
	"regexp"
	"strings"
)

var bqLineRe = regexp.MustCompile(`^( {0,3})(>)( ?)(.*)$`)

func (p *parser) parseBlockQuote(parent *Token, lls []lline) {
	first := lls[0]
	last := lls[len(lls)-1]
	bq := &Token{
		Type:      TypeBlockQuote,
		StartLine: first.idx + 1, StartColumn: first.col + leadingSpaces(first.text),
		EndLine: last.idx + 1, EndColumn: last.col + runeLen(last.text) - 1,
	}
	addChild(parent, bq)

	inner := make([]lline, 0, len(lls))
	for _, ll := range lls {
		m := bqLineRe.FindStringSubmatch(ll.text)
		if m == nil {
			// Lazy continuation line.
			inner = append(inner, ll)
			continue
		}

		ln := ll.idx + 1
		indent := m[1]
		space := m[3]
		markerCol := ll.col + runeLen(indent)
		prefixText := m[2] + space
		prefix := &Token{
			Type:        TypeBlockQuotePrefix,
			StartLine:   ln,
			StartColumn: markerCol,
			EndLine:     ln,
			EndColumn:   markerCol + runeLen(prefixText) - 1,
			Text:        prefixText,
		}
		addChild(bq, prefix)
		addChild(
			prefix,
			&Token{
				Type:        TypeBlockQuoteMarker,
				StartLine:   ln,
				StartColumn: markerCol,
				EndLine:     ln,
				EndColumn:   markerCol,
				Text:        ">",
			},
		)

		if space != "" {
			addChild(
				prefix,
				&Token{
					Type:        TypeBlockQuotePrefixWhitespace,
					StartLine:   ln,
					StartColumn: markerCol + 1,
					EndLine:     ln,
					EndColumn:   markerCol + 1,
					Text:        space,
				},
			)
		}

		contentCol := markerCol + 1 + runeLen(space)
		// Extra spaces after the single-space prefix become a linePrefix sibling
		// (matches micromark; MD027 detects this) — but only when the content is
		// inline (paragraph) text, not a nested block whose own indentation rules
		// apply (list/code/heading/fence/hr/nested quote).
		content := m[4]
		extra := content[:len(content)-len(strings.TrimLeft(content, " \t"))]
		rest := strings.TrimLeft(content, " \t")

		startsSubBlock := rest == "" || isFenceOpen(content) || isListItem(content) ||
			thematicRe.MatchString(content) || (atxRe.MatchString(content) && isATX(content)) ||
			blockquoteRe.MatchString(content) || leadingSpaces(content) >= 4
		if extra != "" && !startsSubBlock {
			addChild(
				bq,
				&Token{
					Type:        TypeLinePrefix,
					StartLine:   ln,
					StartColumn: contentCol,
					EndLine:     ln,
					EndColumn:   contentCol + runeLen(extra) - 1,
					Text:        extra,
				},
			)
		}

		inner = append(inner, lline{idx: ll.idx, text: content, col: contentCol})
	}

	p.parseLines(bq, inner)
}

func (p *parser) parseList(parent *Token, lls []lline, i int) int {
	ordered := orderedRe.MatchString(lls[i].text) || orderedEmptyRe.MatchString(lls[i].text)

	listType := TypeListUnordered
	if ordered {
		listType = TypeListOrdered
	}

	list := &Token{
		Type:        listType,
		StartLine:   lls[i].idx + 1,
		StartColumn: lls[i].col + leadingSpaces(lls[i].text),
	}
	addChild(parent, list)

	j := i

	lastLine := lls[i].idx + 1
	for j < len(lls) {
		ll := lls[j]
		if isBlank(ll.text) {
			// A blank may separate items; include if next is a same-type item or
			// indented content. Otherwise the list ends.
			k := j + 1
			for k < len(lls) && isBlank(lls[k].text) {
				k++
			}

			if k < len(lls) &&
				(sameListType(lls[k].text, ordered) && leadingSpaces(lls[k].text) < 4) {
				j = k
				continue
			}

			break
		}

		if !sameListType(ll.text, ordered) || leadingSpaces(ll.text) >= 4 {
			break
		}

		j2 := p.parseListItem(list, lls, j, ordered)
		lastLine = lls[j2-1].idx + 1
		j = j2
	}

	list.EndLine = lastLine

	list.EndColumn = lls[i].col + runeLen(p.lines[lastLine-1].text) - 1
	if list.EndColumn < list.StartColumn {
		list.EndColumn = list.StartColumn
	}

	return j
}

func sameListType(text string, ordered bool) bool {
	if ordered {
		return orderedRe.MatchString(text) || orderedEmptyRe.MatchString(text)
	}

	return bulletRe.MatchString(text) || bulletEmptyRe.MatchString(text)
}

func (p *parser) parseListItem(list *Token, lls []lline, i int, ordered bool) int {
	ll := lls[i]
	ln := ll.idx + 1

	var indent, marker, value, sep, rest string

	if ordered {
		if m := orderedRe.FindStringSubmatch(ll.text); m != nil {
			indent, value, marker, sep, rest = m[1], m[2], m[3], m[4], m[5]
		} else {
			m = orderedEmptyRe.FindStringSubmatch(ll.text)
			indent, value, marker, sep, rest = m[1], m[2], m[3], m[4], ""
		}
	} else {
		if m := bulletRe.FindStringSubmatch(ll.text); m != nil {
			indent, marker, sep, rest = m[1], m[2], m[3], m[4]
		} else {
			m = bulletEmptyRe.FindStringSubmatch(ll.text)
			indent, marker, sep, rest = m[1], m[2], m[3], ""
		}
	}

	markerCol := ll.col + runeLen(indent)
	prefixText := indent + value + marker + sep
	prefix := &Token{
		Type:        TypeListItemPrefix,
		StartLine:   ln,
		StartColumn: markerCol,
		EndLine:     ln,
		EndColumn:   ll.col + runeLen(prefixText) - 1,
		Text:        prefixText,
	}
	addChild(list, prefix)

	if len(indent) > 0 {
		addChild(
			prefix,
			&Token{
				Type:        TypeLinePrefix,
				StartLine:   ln,
				StartColumn: ll.col,
				EndLine:     ln,
				EndColumn:   ll.col + runeLen(indent) - 1,
				Text:        indent,
			},
		)
	}

	if ordered {
		addChild(
			prefix,
			&Token{
				Type:        TypeListItemValue,
				StartLine:   ln,
				StartColumn: markerCol,
				EndLine:     ln,
				EndColumn:   markerCol + runeLen(value) - 1,
				Text:        value,
			},
		)
		mcol := markerCol + runeLen(value)
		addChild(
			prefix,
			&Token{
				Type:        TypeListItemMarker,
				StartLine:   ln,
				StartColumn: mcol,
				EndLine:     ln,
				EndColumn:   mcol,
				Text:        marker,
			},
		)
	} else {
		addChild(
			prefix,
			&Token{
				Type:        TypeListItemMarker,
				StartLine:   ln,
				StartColumn: markerCol,
				EndLine:     ln,
				EndColumn:   markerCol,
				Text:        marker,
			},
		)
	}

	if sep != "" {
		scol := ll.col + runeLen(indent+value+marker)
		addChild(
			prefix,
			&Token{
				Type:        TypeListItemPrefixWhitespace,
				StartLine:   ln,
				StartColumn: scol,
				EndLine:     ln,
				EndColumn:   scol + runeLen(sep) - 1,
				Text:        sep,
			},
		)
	}

	contentIndent := runeLen(indent + value + marker + sep)
	contentCol := ll.col + contentIndent

	// Gather item lines: the first line's rest plus continuation lines indented
	// at least to contentIndent (or blank).
	item := []lline{{idx: ll.idx, text: rest, col: contentCol}}

	j := i + 1
	for j < len(lls) {
		nl := lls[j]
		if isBlank(nl.text) {
			item = append(item, lline{idx: nl.idx, text: "", col: contentCol})
			j++

			continue
		}

		if leadingSpaces(nl.text) >= contentIndent {
			item = append(
				item,
				lline{
					idx:  nl.idx,
					text: nl.text[minInt(contentIndent, len(nl.text)):],
					col:  nl.col + contentIndent,
				},
			)
			j++

			continue
		}

		break
	}
	// Trim trailing blank lines from the item.
	for len(item) > 1 && isBlank(item[len(item)-1].text) {
		item = item[:len(item)-1]
		j--
	}

	p.parseLines(
		prefix.Parent,
		item,
	) // attach content to the list (micromark nests under listItem; we use list)

	return j
}

// --- GFM table ---

func (p *parser) parseTable(parent *Token, textLines []lline) {
	first := textLines[0]
	last := textLines[len(textLines)-1]
	table := &Token{
		Type:      TypeTable,
		StartLine: first.idx + 1, StartColumn: first.col + leadingSpaces(first.text),
		EndLine: last.idx + 1, EndColumn: last.col + runeLen(last.text) - 1,
	}
	addChild(parent, table)

	for k, ll := range textLines {
		ln := ll.idx + 1

		rowType := TypeTableRow

		switch k {
		case 0:
			rowType = TypeTableHeader
		case 1:
			rowType = TypeTableDelimiterRow
		}

		row := &Token{
			Type:        rowType,
			StartLine:   ln,
			StartColumn: ll.col + leadingSpaces(ll.text),
			EndLine:     ln,
			EndColumn:   ll.col + runeLen(ll.text) - 1,
			Text:        strings.TrimSpace(ll.text),
		}
		addChild(table, row)
		p.parseTableCells(row, ll, k == 1)
	}
}

func (p *parser) parseTableCells(row *Token, ll lline, delimiter bool) {
	ln := ll.idx + 1
	text := ll.text
	col := ll.col
	idx := 0
	runes := []rune(text)
	cellStart := 0
	emitCell := func(start, end int) {
		seg := string(runes[start:end])

		cellType := TypeTableData
		if delimiter {
			cellType = TypeTableDelimiter
		}

		c := &Token{
			Type:        cellType,
			StartLine:   ln,
			StartColumn: col + start,
			EndLine:     ln,
			EndColumn:   col + end - 1,
			Text:        seg,
		}
		addChild(row, c)
	}

	for idx < len(runes) {
		if runes[idx] == '|' && (idx == 0 || runes[idx-1] != '\\') {
			if idx > cellStart {
				emitCell(cellStart, idx)
			}

			addChild(
				row,
				&Token{
					Type:        TypeTableCellDivider,
					StartLine:   ln,
					StartColumn: col + idx,
					EndLine:     ln,
					EndColumn:   col + idx,
					Text:        "|",
				},
			)
			cellStart = idx + 1
		}

		idx++
	}

	if cellStart < len(runes) {
		emitCell(cellStart, len(runes))
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}
