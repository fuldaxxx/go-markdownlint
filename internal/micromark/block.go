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
	"strings"
)

// lline is a logical line: an index into the source lines plus the content
// remaining after stripping the enclosing container prefixes, and the 1-based
// column in the original line where that content begins.
type lline struct {
	idx  int
	text string
	col  int
}

// parseBlocks parses source lines [start,end) under parent. Container parsers
// build their own lline slices.
func (p *parser) parseBlocks(parent *Token, start, end int) {
	lls := make([]lline, 0, end-start)
	for i := start; i < end; i++ {
		lls = append(lls, lline{idx: i, text: p.lines[i].text, col: 1})
	}

	p.parseLines(parent, lls)
}

func (p *parser) parseLines(parent *Token, lls []lline) {
	i := 0
	for i < len(lls) {
		ll := lls[i]
		switch {
		case isBlank(ll.text):
			i++
		case blockquoteRe.MatchString(ll.text):
			j := i
			for j < len(lls) && (blockquoteRe.MatchString(lls[j].text) || (!isBlank(lls[j].text) && j > i && isLazyContinuation(lls[j].text))) {
				j++
			}

			p.parseBlockQuote(parent, lls[i:j])
			i = j
		case isFenceOpen(ll.text):
			j := p.parseFencedCode(parent, lls, i)
			i = j
		case atxRe.MatchString(strings.TrimLeft(ll.text, " ")) && leadingSpaces(ll.text) < 4 && isATX(ll.text):
			p.parseATXHeading(parent, ll)

			i++
		case thematicRe.MatchString(ll.text):
			p.parseThematicBreak(parent, ll)

			i++
		case isListItem(ll.text) && leadingSpaces(ll.text) < 4:
			j := p.parseList(parent, lls, i)
			i = j
		case htmlBlockRe.MatchString(ll.text):
			j := p.parseHTMLBlock(parent, lls, i)
			i = j
		case leadingSpaces(ll.text) >= 4:
			j := p.parseIndentedCode(parent, lls, i)
			i = j
		case defRe.MatchString(ll.text):
			p.parseDefinition(parent, ll)

			i++
		default:
			j := p.parseParagraph(parent, lls, i)
			i = j
		}
	}
}

func isLazyContinuation(text string) bool {
	// A non-blank line that isn't itself a new block start can lazily continue a
	// blockquote/paragraph. Kept conservative.
	return !isBlank(text) && !isFenceOpen(text) && !thematicRe.MatchString(text) &&
		!isListItem(text) && !htmlBlockRe.MatchString(text)
}

// canLazyContinue reports whether text is plain paragraph content that can be
// part of a lazy continuation inside a list item. It is stricter than
// isLazyContinuation: an ATX heading or a blockquote also interrupts a
// paragraph, so neither may be lazily continued.
func canLazyContinue(text string) bool {
	trimmed := strings.TrimLeft(text, " ")

	return isLazyContinuation(text) &&
		!blockquoteRe.MatchString(text) &&
		!(atxRe.MatchString(trimmed) && leadingSpaces(text) < 4)
}

func isATX(text string) bool {
	return atxRe.MatchString(text)
}

func isFenceOpen(text string) bool {
	return fenceOpenRe.MatchString(text) && leadingSpaces(text) < 4
}

func isListItem(text string) bool {
	return bulletRe.MatchString(text) || orderedRe.MatchString(text) ||
		bulletEmptyRe.MatchString(text) || orderedEmptyRe.MatchString(text)
}

// --- ATX heading ---

func (p *parser) parseATXHeading(parent *Token, ll lline) {
	m := atxRe.FindStringSubmatch(ll.text)
	indent := m[1]
	hashes := m[2]
	rest := m[3]
	lineNo := ll.idx + 1
	startCol := ll.col + runeLen(indent)
	// Leading indentation is a sibling linePrefix preceding the heading (matches
	// micromark; MD023 relies on this ordering).
	if len(indent) > 0 {
		addChild(
			parent,
			&Token{
				Type:        TypeLinePrefix,
				StartLine:   lineNo,
				StartColumn: ll.col,
				EndLine:     lineNo,
				EndColumn:   ll.col + runeLen(indent) - 1,
				Text:        indent,
			},
		)
	}

	contentEndCol := ll.col + runeLen(strings.TrimRight(ll.text, " \t")) - 1
	if contentEndCol < startCol {
		contentEndCol = startCol
	}

	h := &Token{
		Type:        TypeAtxHeading,
		StartLine:   lineNo,
		StartColumn: startCol,
		EndLine:     lineNo,
		EndColumn:   contentEndCol,
		Text:        strings.TrimRight(ll.text[len(indent):], " \t"),
	}
	addChild(parent, h)
	addChild(
		h,
		&Token{
			Type:        TypeAtxHeadingSequence,
			StartLine:   lineNo,
			StartColumn: startCol,
			EndLine:     lineNo,
			EndColumn:   startCol + len(hashes) - 1,
			Text:        hashes,
		},
	)

	restCol := startCol + runeLen(hashes)
	leadingLen := len(rest) - len(strings.TrimLeft(rest, " \t"))
	leading := rest[:leadingLen]
	body := rest[leadingLen:]
	// Closing-sequence detection: trailing run of # preceded by whitespace.
	closing := ""
	closingWS := ""
	textPart := strings.TrimRight(body, " \t")

	e := len(textPart)
	for e > 0 && textPart[e-1] == '#' {
		e--
	}

	if e < len(textPart) { // trailing run of '#'
		run := textPart[e:]
		beforeRun := textPart[:e]

		trimmedBefore := strings.TrimRight(beforeRun, " \t")
		if len(trimmedBefore) < len(beforeRun) {
			closing = run
			closingWS = beforeRun[len(trimmedBefore):]
			textPart = trimmedBefore
		}
	}

	if leadingLen > 0 {
		addChild(
			h,
			&Token{
				Type:        TypeWhitespace,
				StartLine:   lineNo,
				StartColumn: restCol,
				EndLine:     lineNo,
				EndColumn:   restCol + runeLen(leading) - 1,
				Text:        leading,
			},
		)
	}

	textCol := restCol + runeLen(leading)
	if textPart != "" {
		htxt := &Token{
			Type:        TypeAtxHeadingText,
			StartLine:   lineNo,
			StartColumn: textCol,
			EndLine:     lineNo,
			EndColumn:   textCol + runeLen(textPart) - 1,
			Text:        textPart,
		}
		addChild(h, htxt)
		p.parseInline(htxt, textPart, lineNo, textCol)
	}

	if closing != "" {
		twsCol := textCol + runeLen(textPart)
		if closingWS != "" {
			addChild(
				h,
				&Token{
					Type:        TypeWhitespace,
					StartLine:   lineNo,
					StartColumn: twsCol,
					EndLine:     lineNo,
					EndColumn:   twsCol + runeLen(closingWS) - 1,
					Text:        closingWS,
				},
			)
		}

		ccol := twsCol + runeLen(closingWS)
		addChild(
			h,
			&Token{
				Type:        TypeAtxHeadingSequence,
				StartLine:   lineNo,
				StartColumn: ccol,
				EndLine:     lineNo,
				EndColumn:   ccol + len(closing) - 1,
				Text:        closing,
			},
		)
	}
}

// --- Thematic break ---

func (p *parser) parseThematicBreak(parent *Token, ll lline) {
	lineNo := ll.idx + 1
	addChild(
		parent,
		&Token{
			Type:        TypeThematicBreak,
			StartLine:   lineNo,
			StartColumn: ll.col,
			EndLine:     lineNo,
			EndColumn:   ll.col + runeLen(ll.text) - 1,
			Text:        ll.text,
		},
	)
}

// --- Fenced code ---

func (p *parser) parseFencedCode(parent *Token, lls []lline, i int) int {
	open := lls[i]
	m := fenceOpenRe.FindStringSubmatch(open.text)
	indent := m[1]
	fence := m[2]
	info := strings.TrimSpace(m[3])
	fenceChar := fence[0]
	lineNo := open.idx + 1
	code := &Token{Type: TypeCodeFenced, StartLine: lineNo, StartColumn: open.col + runeLen(indent)}
	addChild(parent, code)
	// Opening fence.
	openFence := &Token{
		Type:        TypeCodeFencedFence,
		StartLine:   lineNo,
		StartColumn: open.col + runeLen(indent),
		EndLine:     lineNo,
		EndColumn:   open.col + runeLen(open.text) - 1,
		Text:        open.text[len(indent):],
	}
	addChild(code, openFence)

	seqCol := open.col + runeLen(indent)
	addChild(
		openFence,
		&Token{
			Type:        TypeCodeFencedFenceSeq,
			StartLine:   lineNo,
			StartColumn: seqCol,
			EndLine:     lineNo,
			EndColumn:   seqCol + len(fence) - 1,
			Text:        fence,
		},
	)

	if info != "" {
		infoCol := seqCol + len(fence) + leadingSpaces(m[3])
		addChild(
			openFence,
			&Token{
				Type:        TypeCodeFencedFenceInfo,
				StartLine:   lineNo,
				StartColumn: infoCol,
				EndLine:     lineNo,
				EndColumn:   infoCol + runeLen(info) - 1,
				Text:        info,
			},
		)
	}

	j := i + 1
	lastLine := lineNo

	for j < len(lls) {
		ll := lls[j]
		ln := ll.idx + 1
		t := strings.TrimRight(ll.text, " \t")

		closing := len(t) >= len(fence) && allChar(strings.TrimLeft(t, " "), fenceChar) &&
			leadingSpaces(t) < 4
		if closing {
			addChild(
				code,
				&Token{
					Type:        TypeCodeFencedFence,
					StartLine:   ln,
					StartColumn: ll.col + leadingSpaces(ll.text),
					EndLine:     ln,
					EndColumn:   ll.col + runeLen(ll.text) - 1,
					Text:        strings.TrimLeft(ll.text, " "),
				},
			)
			lastLine = ln
			j++

			break
		}

		content := ll.text
		if runeLen(indent) > 0 && strings.HasPrefix(content, indent) {
			content = content[len(indent):]
		}

		if content != "" {
			addChild(
				code,
				&Token{
					Type:        TypeCodeFlowValue,
					StartLine:   ln,
					StartColumn: ll.col + runeLen(indent),
					EndLine:     ln,
					EndColumn:   ll.col + runeLen(ll.text) - 1,
					Text:        content,
				},
			)
		}

		lastLine = ln
		j++
	}

	code.EndLine = lastLine

	code.EndColumn = open.col + runeLen(p.lines[lastLine-1].text) - 1
	if code.EndColumn < code.StartColumn {
		code.EndColumn = code.StartColumn
	}

	return j
}

func allChar(s string, c byte) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] != c {
			return false
		}
	}

	return true
}

// --- Indented code ---

func (p *parser) parseIndentedCode(parent *Token, lls []lline, i int) int {
	start := lls[i]
	lineNo := start.idx + 1
	code := &Token{Type: TypeCodeIndented, StartLine: lineNo, StartColumn: start.col}
	addChild(parent, code)

	j := i
	lastLine := lineNo

	for j < len(lls) {
		ll := lls[j]
		if isBlank(ll.text) {
			// blank lines are allowed within indented code if more follows
			k := j + 1
			for k < len(lls) && isBlank(lls[k].text) {
				k++
			}

			if k < len(lls) && leadingSpaces(lls[k].text) >= 4 {
				j = k
				continue
			}

			break
		}

		if leadingSpaces(ll.text) < 4 {
			break
		}

		ln := ll.idx + 1
		content := ll.text[4:]
		addChild(
			code,
			&Token{
				Type:        TypeCodeFlowValue,
				StartLine:   ln,
				StartColumn: ll.col + 4,
				EndLine:     ln,
				EndColumn:   ll.col + runeLen(ll.text) - 1,
				Text:        content,
			},
		)
		lastLine = ln
		j++
	}

	code.EndLine = lastLine
	code.EndColumn = start.col + runeLen(p.lines[lastLine-1].text) - 1

	return j
}

// --- HTML block ---

func (p *parser) parseHTMLBlock(parent *Token, lls []lline, i int) int {
	start := lls[i]
	lineNo := start.idx + 1
	isComment := strings.HasPrefix(strings.TrimLeft(start.text, " "), "<!--")
	j := i
	lastLine := lineNo

	var sb strings.Builder

	for j < len(lls) {
		ll := lls[j]
		if j > i && isBlank(ll.text) {
			break
		}

		if j > i {
			sb.WriteString("\n")
		}

		sb.WriteString(ll.text)
		lastLine = ll.idx + 1
		end := isComment && strings.Contains(ll.text, "-->")
		j++

		if end {
			break
		}
	}

	html := &Token{
		Type:        TypeHTMLFlow,
		StartLine:   lineNo,
		StartColumn: start.col + leadingSpaces(start.text),
		EndLine:     lastLine,
		EndColumn:   start.col + runeLen(p.lines[lastLine-1].text) - 1,
		Text:        strings.TrimLeft(sb.String(), " "),
	}
	addChild(parent, html)
	addChild(
		html,
		&Token{
			Type:        TypeHTMLFlowData,
			StartLine:   lineNo,
			StartColumn: html.StartColumn,
			EndLine:     lastLine,
			EndColumn:   html.EndColumn,
			Text:        html.Text,
		},
	)

	return j
}

// --- Definition ---

func (p *parser) parseDefinition(parent *Token, ll lline) {
	m := defRe.FindStringSubmatch(ll.text)
	label := m[1]
	dest := m[2]
	lineNo := ll.idx + 1
	def := &Token{
		Type:        TypeDefinition,
		StartLine:   lineNo,
		StartColumn: ll.col,
		EndLine:     lineNo,
		EndColumn:   ll.col + runeLen(ll.text) - 1,
		Text:        ll.text,
	}
	addChild(parent, def)

	labelStart := strings.Index(ll.text, "[") + 1
	lcol := ll.col + runeLen(ll.text[:labelStart])
	ls := &Token{
		Type:        TypeDefinitionLabelString,
		StartLine:   lineNo,
		StartColumn: lcol,
		EndLine:     lineNo,
		EndColumn:   lcol + runeLen(label) - 1,
		Text:        label,
	}
	addChild(def, ls)

	destStart := strings.Index(ll.text, dest)
	dcol := ll.col + runeLen(ll.text[:destStart])
	dd := &Token{
		Type:        TypeDefinitionDestination,
		StartLine:   lineNo,
		StartColumn: dcol,
		EndLine:     lineNo,
		EndColumn:   dcol + runeLen(dest) - 1,
		Text:        dest,
	}
	addChild(def, dd)
	addChild(
		dd,
		&Token{
			Type:        TypeDefinitionDestinationString,
			StartLine:   lineNo,
			StartColumn: dcol,
			EndLine:     lineNo,
			EndColumn:   dcol + runeLen(dest) - 1,
			Text:        dest,
		},
	)
}

// --- Paragraph (with possible setext heading and tables) ---

func (p *parser) parseParagraph(parent *Token, lls []lline, i int) int {
	j := i

	var textLines []lline

	for j < len(lls) {
		ll := lls[j]
		if isBlank(ll.text) {
			break
		}

		if j > i {
			// interrupting constructs
			if isFenceOpen(ll.text) ||
				(isListItem(ll.text) && leadingSpaces(ll.text) < 4 && interruptsList(ll.text)) ||
				thematicRe.MatchString(ll.text) ||
				atxRe.MatchString(ll.text) && isATX(ll.text) ||
				blockquoteRe.MatchString(ll.text) ||
				htmlBlockRe.MatchString(ll.text) {
				break
			}
			// setext underline
			if setextRe.MatchString(ll.text) {
				p.emitSetext(parent, textLines, ll)
				return j + 1
			}
		}

		textLines = append(textLines, ll)
		j++
	}

	if len(textLines) == 0 {
		return i + 1
	}
	// GFM table detection: header line + delimiter row.
	if len(textLines) >= 2 && strings.Contains(textLines[0].text, "|") &&
		tableDelimRe.MatchString(textLines[1].text) {
		p.parseTable(parent, textLines)
		return j
	}

	p.emitParagraph(parent, textLines)

	return j
}

func interruptsList(text string) bool {
	// Only list items that start with content can interrupt a paragraph (and
	// ordered lists must start at 1). Kept simple: bullets and "1." interrupt.
	if m := orderedRe.FindStringSubmatch(text); m != nil {
		return m[2] == "1" && m[4] != ""
	}

	if m := bulletRe.FindStringSubmatch(text); m != nil {
		return m[4] != ""
	}

	return false
}

func (p *parser) emitParagraph(parent *Token, textLines []lline) {
	first := textLines[0]
	last := textLines[len(textLines)-1]
	para := &Token{
		Type:      TypeParagraph,
		StartLine: first.idx + 1, StartColumn: first.col + leadingSpaces(first.text),
		EndLine: last.idx + 1, EndColumn: last.col + runeLen(last.text) - 1,
	}
	addChild(parent, para)

	for k, ll := range textLines {
		ln := ll.idx + 1
		content := strings.TrimLeft(ll.text, " ")
		col := ll.col + (runeLen(ll.text) - runeLen(content))
		p.parseInline(para, content, ln, col)

		if k < len(textLines)-1 {
			addChild(
				para,
				&Token{
					Type:        TypeLineEnding,
					StartLine:   ln,
					StartColumn: ll.col + runeLen(ll.text),
					EndLine:     ln,
					EndColumn:   ll.col + runeLen(ll.text),
					Text:        "\n",
				},
			)
		}
	}
}

func (p *parser) emitSetext(parent *Token, textLines []lline, underline lline) {
	if len(textLines) == 0 {
		// Treat as thematic-like; emit paragraph fallback.
		return
	}

	first := textLines[0]
	last := textLines[len(textLines)-1]
	uln := underline.idx + 1
	h := &Token{
		Type:      TypeSetextHeading,
		StartLine: first.idx + 1, StartColumn: first.col + leadingSpaces(first.text),
		EndLine: uln, EndColumn: underline.col + runeLen(underline.text) - 1,
	}
	addChild(parent, h)

	htxt := &Token{
		Type:      TypeSetextHeadingText,
		StartLine: first.idx + 1, StartColumn: first.col + leadingSpaces(first.text),
		EndLine: last.idx + 1, EndColumn: last.col + runeLen(last.text) - 1,
	}
	addChild(h, htxt)

	for k, ll := range textLines {
		ln := ll.idx + 1
		content := strings.TrimLeft(ll.text, " ")
		col := ll.col + (runeLen(ll.text) - runeLen(content))
		p.parseInline(htxt, content, ln, col)

		if k < len(textLines)-1 {
			addChild(
				htxt,
				&Token{
					Type:        TypeLineEnding,
					StartLine:   ln,
					StartColumn: ll.col + runeLen(ll.text),
					EndLine:     ln,
					EndColumn:   ll.col + runeLen(ll.text),
					Text:        "\n",
				},
			)
		}
	}

	ul := strings.TrimSpace(underline.text)
	addChild(
		h,
		&Token{
			Type:        TypeSetextHeadingLine,
			StartLine:   uln,
			StartColumn: underline.col + leadingSpaces(underline.text),
			EndLine:     uln,
			EndColumn:   underline.col + runeLen(underline.text) - 1,
			Text:        ul,
		},
	)
}
