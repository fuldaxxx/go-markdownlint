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
	"sort"
	"strings"
	"unicode"
)

// parseInline tokenizes inline content text starting at (line, col) and appends
// child tokens to parent. text must be a single logical line's content (no
// newlines); multi-line containers call this once per line.
func (p *parser) parseInline(parent *Token, text string, line, col int) {
	runes := []rune(text)

	toks := tokenizeInline(runes, line, col, p.definedLabels)
	for _, t := range toks {
		addChild(parent, t)
	}
}

// tokenizeInline produces inline tokens for runes at the given 1-based line/col.
// defined is the set of normalized link reference definition labels; it is used
// to decide whether bracket forms are real links/images or undefined references.
func tokenizeInline(runes []rune, line, startCol int, defined map[string]bool) []*Token {
	var out []*Token

	dataStart := -1
	flushData := func(end int) {
		if dataStart >= 0 && end > dataStart {
			seg := string(runes[dataStart:end])
			out = append(
				out,
				&Token{
					Type:        TypeData,
					StartLine:   line,
					StartColumn: startCol + dataStart,
					EndLine:     line,
					EndColumn:   startCol + end - 1,
					Text:        seg,
				},
			)
		}

		dataStart = -1
	}

	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '\\' && i+1 < len(runes) && isASCIIPunct(runes[i+1]):
			flushData(i)
			esc := &Token{
				Type:        TypeCharacterEscape,
				StartLine:   line,
				StartColumn: startCol + i,
				EndLine:     line,
				EndColumn:   startCol + i + 1,
				Text:        string(runes[i : i+2]),
			}
			out = append(out, esc)
			i += 2
		case r == '`':
			if tok, ni := scanCodeSpan(runes, i, line, startCol); tok != nil {
				flushData(i)

				out = append(out, tok)
				i = ni

				continue
			}

			if dataStart < 0 {
				dataStart = i
			}

			i++
		case r == '<':
			if tok, ni := scanAutolinkOrHTML(runes, i, line, startCol); tok != nil {
				flushData(i)

				out = append(out, tok)
				i = ni

				continue
			}

			if dataStart < 0 {
				dataStart = i
			}

			i++
		case r == '!' && i+1 < len(runes) && runes[i+1] == '[':
			if tok, ni := scanLinkOrImage(runes, i, line, startCol, true, defined); tok != nil {
				flushData(i)

				out = append(out, tok)
				i = ni

				continue
			}

			if dataStart < 0 {
				dataStart = i
			}

			i++
		case r == '[':
			if tok, ni := scanLinkOrImage(runes, i, line, startCol, false, defined); tok != nil {
				flushData(i)

				out = append(out, tok)
				i = ni

				continue
			}

			if dataStart < 0 {
				dataStart = i
			}

			i++
		case r == '*' || r == '_':
			// Emit each maximal run of '*' or '_' as its own data token so bare
			// markers stay individually available (MD037) and so emphasis
			// resolution can match runs precisely.
			flushData(i)

			j := i
			for j < len(runes) && runes[j] == r {
				j++
			}

			out = append(
				out,
				&Token{
					Type:        TypeData,
					StartLine:   line,
					StartColumn: startCol + i,
					EndLine:     line,
					EndColumn:   startCol + j - 1,
					Text:        string(runes[i:j]),
				},
			)
			i = j
		default:
			if dataStart < 0 {
				dataStart = i
			}

			i++
		}
	}

	flushData(len(runes))
	out = resolveEmphasis(out, runes, startCol)
	out = extractLiteralAutolinks(out)

	return out
}

// isEscaped reports whether the rune at index i is preceded by an odd number of
// backslashes (i.e. it is backslash-escaped).
func isEscaped(runes []rune, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
		n++
	}

	return n%2 == 1
}

func isASCIIPunct(r rune) bool {
	return strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r)
}

// scanCodeSpan scans a backtick code span. Returns nil if no closing run.
func scanCodeSpan(runes []rune, start, line, startCol int) (*Token, int) {
	n := 0
	for start+n < len(runes) && runes[start+n] == '`' {
		n++
	}
	// find closing run of exactly n backticks
	i := start + n
	for i < len(runes) {
		if runes[i] == '`' {
			m := 0
			for i+m < len(runes) && runes[i+m] == '`' {
				m++
			}

			if m == n {
				end := i + m
				ct := &Token{
					Type:        TypeCodeText,
					StartLine:   line,
					StartColumn: startCol + start,
					EndLine:     line,
					EndColumn:   startCol + end - 1,
					Text:        string(runes[start:end]),
				}
				addChild(
					ct,
					&Token{
						Type:        TypeCodeTextSequence,
						StartLine:   line,
						StartColumn: startCol + start,
						EndLine:     line,
						EndColumn:   startCol + start + n - 1,
						Text:        string(runes[start : start+n]),
					},
				)

				dataStart, dataEnd := start+n, i
				if dataEnd > dataStart {
					addChild(
						ct,
						&Token{
							Type:        TypeCodeTextData,
							StartLine:   line,
							StartColumn: startCol + dataStart,
							EndLine:     line,
							EndColumn:   startCol + dataEnd - 1,
							Text:        string(runes[dataStart:dataEnd]),
						},
					)
				}

				addChild(
					ct,
					&Token{
						Type:        TypeCodeTextSequence,
						StartLine:   line,
						StartColumn: startCol + i,
						EndLine:     line,
						EndColumn:   startCol + end - 1,
						Text:        string(runes[i:end]),
					},
				)

				return ct, end
			}

			i += m

			continue
		}

		i++
	}

	return nil, start
}

var (
	autolinkURIRe   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]{1,31}:[^<>\x00-\x20]*$`)
	autolinkEmailRe = regexp.MustCompile(
		`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`,
	)
	htmlInlineRe = regexp.MustCompile(
		`(?s)^<(/[a-zA-Z][a-zA-Z0-9-]*\s*>|[a-zA-Z][a-zA-Z0-9-]*(\s+[a-zA-Z_:][a-zA-Z0-9_.:-]*(\s*=\s*("[^"]*"|'[^']*'|[^\s"'=<>` + "`" + `]+))?)*\s*/?>|!--.*?-->)`,
	)
)

func scanAutolinkOrHTML(runes []rune, start, line, startCol int) (*Token, int) {
	// Find the closing '>' for a potential autolink.
	end := -1

	for i := start + 1; i < len(runes); i++ {
		if runes[i] == '>' {
			end = i
			break
		}

		if runes[i] == ' ' {
			break
		}
	}

	if end > start {
		inner := string(runes[start+1 : end])
		if autolinkURIRe.MatchString(inner) {
			a := &Token{
				Type:        TypeAutolink,
				StartLine:   line,
				StartColumn: startCol + start,
				EndLine:     line,
				EndColumn:   startCol + end,
				Text:        string(runes[start : end+1]),
			}
			addChild(
				a,
				&Token{
					Type:        TypeAutolinkProtocol,
					StartLine:   line,
					StartColumn: startCol + start + 1,
					EndLine:     line,
					EndColumn:   startCol + end - 1,
					Text:        inner,
				},
			)

			return a, end + 1
		}

		if autolinkEmailRe.MatchString(inner) {
			a := &Token{
				Type:        TypeAutolink,
				StartLine:   line,
				StartColumn: startCol + start,
				EndLine:     line,
				EndColumn:   startCol + end,
				Text:        string(runes[start : end+1]),
			}
			addChild(
				a,
				&Token{
					Type:        TypeAutolinkEmail,
					StartLine:   line,
					StartColumn: startCol + start + 1,
					EndLine:     line,
					EndColumn:   startCol + end - 1,
					Text:        inner,
				},
			)

			return a, end + 1
		}
	}
	// HTML inline.
	if m := htmlInlineRe.FindString(string(runes[start:])); m != "" {
		mr := []rune(m)
		h := &Token{
			Type:        TypeHTMLText,
			StartLine:   line,
			StartColumn: startCol + start,
			EndLine:     line,
			EndColumn:   startCol + start + len(mr) - 1,
			Text:        m,
		}
		addChild(
			h,
			&Token{
				Type:        TypeHTMLTextData,
				StartLine:   line,
				StartColumn: startCol + start,
				EndLine:     line,
				EndColumn:   startCol + start + len(mr) - 1,
				Text:        m,
			},
		)

		return h, start + len(mr)
	}

	return nil, start
}

// scanLinkOrImage scans [text](dest) / [text][ref] / [text][] / [text].
//
// Inline links (with a `(dest)` resource) are always real link/image tokens.
// Reference and shortcut forms are real link/image tokens only when their
// effective label matches a collected definition (via defined). Otherwise the
// brackets are not a link: when the label text is non-empty and contains no
// ']', a synthetic undefinedReference* token is emitted (mirroring
// markdownlint's micromark shim); otherwise nil is returned so the caller
// leaves the '['/']' as plain data.
func scanLinkOrImage(
	runes []rune,
	start, line, startCol int,
	image bool,
	defined map[string]bool,
) (*Token, int) {
	labelOpen := start
	if image {
		labelOpen = start + 1 // '[' after '!'
	}
	// find matching ']'
	depth := 0
	close := -1

	for i := labelOpen; i < len(runes); i++ {
		if isEscaped(runes, i) {
			continue
		}

		switch runes[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				close = i
			}
		}

		if close >= 0 {
			break
		}
	}

	if close < 0 {
		return nil, start
	}

	labelText := string(runes[labelOpen+1 : close])

	tokType := TypeLink
	if image {
		tokType = TypeImage
	}

	end := close + 1

	buildLink := func() *Token {
		link := &Token{Type: tokType, StartLine: line, StartColumn: startCol + start}
		label := &Token{
			Type:        TypeLabel,
			StartLine:   line,
			StartColumn: startCol + start,
			EndLine:     line,
			EndColumn:   startCol + close,
			Text:        string(runes[start : close+1]),
		}
		addChild(link, label)

		lt := &Token{
			Type:        TypeLabelText,
			StartLine:   line,
			StartColumn: startCol + labelOpen + 1,
			EndLine:     line,
			EndColumn:   startCol + close - 1,
			Text:        labelText,
		}
		addChild(label, lt)

		for _, c := range tokenizeInline([]rune(labelText), line, startCol+labelOpen+1, defined) {
			addChild(lt, c)
		}

		return link
	}

	// inline resource (dest): always a real link/image.
	if end < len(runes) && runes[end] == '(' {
		rclose := -1

		for i := end + 1; i < len(runes); i++ {
			if runes[i] == ')' {
				rclose = i
				break
			}
		}

		if rclose >= 0 {
			link := buildLink()
			res := &Token{
				Type:        TypeResource,
				StartLine:   line,
				StartColumn: startCol + end,
				EndLine:     line,
				EndColumn:   startCol + rclose,
				Text:        string(runes[end : rclose+1]),
			}
			addChild(link, res)

			destAll := strings.TrimSpace(string(runes[end+1 : rclose]))

			dest := destAll
			if sp := strings.IndexAny(destAll, " \t"); sp >= 0 {
				dest = destAll[:sp]
			}

			if dest != "" {
				doff := end + 1 + indexRune(runes[end+1:rclose], dest)
				rd := &Token{
					Type:        TypeResourceDestination,
					StartLine:   line,
					StartColumn: startCol + doff,
					EndLine:     line,
					EndColumn:   startCol + doff + runeLen(dest) - 1,
					Text:        dest,
				}
				addChild(res, rd)
				// micromark wraps the destination string in a literal (<...>) or
				// raw container; rules navigate through this intermediate level.
				if strings.HasPrefix(dest, "<") && strings.HasSuffix(dest, ">") &&
					runeLen(dest) >= 2 {
					inner := dest[1 : len(dest)-1]
					lit := &Token{
						Type:        TypeResourceDestinationLiteral,
						StartLine:   line,
						StartColumn: startCol + doff,
						EndLine:     line,
						EndColumn:   startCol + doff + runeLen(dest) - 1,
						Text:        dest,
					}
					addChild(rd, lit)

					if inner != "" {
						addChild(
							lit,
							&Token{
								Type:        TypeResourceDestinationString,
								StartLine:   line,
								StartColumn: startCol + doff + 1,
								EndLine:     line,
								EndColumn:   startCol + doff + runeLen(dest) - 2,
								Text:        inner,
							},
						)
					}
				} else {
					raw := &Token{
						Type:        TypeResourceDestinationRaw,
						StartLine:   line,
						StartColumn: startCol + doff,
						EndLine:     line,
						EndColumn:   startCol + doff + runeLen(dest) - 1,
						Text:        dest,
					}
					addChild(rd, raw)
					addChild(
						raw,
						&Token{
							Type:        TypeResourceDestinationString,
							StartLine:   line,
							StartColumn: startCol + doff,
							EndLine:     line,
							EndColumn:   startCol + doff + runeLen(dest) - 1,
							Text:        dest,
						},
					)
				}
			}

			link.EndLine = line
			link.EndColumn = startCol + rclose

			return link, rclose + 1
		}
	}

	// reference [ref] (full [text][ref] or collapsed [text][]).
	if end < len(runes) && runes[end] == '[' {
		rclose := -1

		for i := end + 1; i < len(runes); i++ {
			if runes[i] == ']' && !isEscaped(runes, i) {
				rclose = i
				break
			}
		}

		if rclose >= 0 {
			refText := string(runes[end+1 : rclose])
			// Effective label: the ref if present, otherwise the link text.
			effective := refText
			if strings.TrimSpace(effective) == "" {
				effective = labelText
			}

			if defined[normalizeLabel(effective)] {
				link := buildLink()
				ref := &Token{
					Type:        TypeReference,
					StartLine:   line,
					StartColumn: startCol + end,
					EndLine:     line,
					EndColumn:   startCol + rclose,
					Text:        string(runes[end : rclose+1]),
				}
				addChild(link, ref)

				if refText != "" {
					rs := &Token{
						Type:        TypeReferenceString,
						StartLine:   line,
						StartColumn: startCol + end + 1,
						EndLine:     line,
						EndColumn:   startCol + rclose - 1,
						Text:        refText,
					}
					addChild(ref, rs)
					// Give referenceString inline children so helpers that read
					// its concatenated child text (refTokenText) see the label.
					for _, c := range tokenizeInline([]rune(refText), line, startCol+end+1, defined) {
						addChild(rs, c)
					}
				}

				link.EndLine = line
				link.EndColumn = startCol + rclose

				return link, rclose + 1
			}
			// Not defined: emit an undefined reference (full or collapsed).
			urType := TypeUndefinedReferenceFull
			if strings.TrimSpace(refText) == "" {
				urType = TypeUndefinedReferenceCollapsed
			}

			if tok := buildUndefinedReference(
				urType,
				runes,
				start,
				labelOpen,
				close,
				rclose,
				line,
				startCol,
			); tok != nil {
				return tok, rclose + 1
			}

			return nil, start
		}
	}

	// shortcut [text].
	if defined[normalizeLabel(labelText)] {
		link := buildLink()
		link.EndLine = line
		link.EndColumn = startCol + close

		return link, end
	}

	if tok := buildUndefinedReference(
		TypeUndefinedReferenceShortcut,
		runes,
		start,
		labelOpen,
		close,
		close,
		line,
		startCol,
	); tok != nil {
		return tok, end
	}

	return nil, start
}

// buildUndefinedReference constructs a synthetic undefinedReference* token
// spanning runes[start:spanClose+1], wrapping an undefinedReference child whose
// children are the label's inline tokens. The label content is the text between
// the label's brackets (runes[labelOpen+1:close]); the wrapper's concatenated
// child text (trimmed) must equal that label. Returns nil if the label is empty
// or contains a ']', in which case the caller leaves the brackets as plain
// data.
func buildUndefinedReference(
	urType TokenType,
	runes []rune,
	start, labelOpen, close, spanClose, line, startCol int,
) *Token {
	labelText := string(runes[labelOpen+1 : close])
	if strings.TrimSpace(labelText) == "" || strings.Contains(labelText, "]") {
		return nil
	}

	ur := &Token{
		Type:        urType,
		StartLine:   line,
		StartColumn: startCol + start,
		EndLine:     line,
		EndColumn:   startCol + spanClose,
		Text:        string(runes[start : spanClose+1]),
	}
	inner := &Token{
		Type:        TypeUndefinedReference,
		StartLine:   line,
		StartColumn: startCol + start,
		EndLine:     line,
		EndColumn:   startCol + spanClose,
		Text:        string(runes[start : spanClose+1]),
	}
	addChild(ur, inner)
	// The undefinedReference child's concatenated child text must equal the
	// label (consumers join children .Text directly), so emit a single data
	// token replicating the raw data covering the label.
	addChild(
		inner,
		&Token{
			Type:        TypeData,
			StartLine:   line,
			StartColumn: startCol + labelOpen + 1,
			EndLine:     line,
			EndColumn:   startCol + close - 1,
			Text:        labelText,
		},
	)

	return ur
}

func indexRune(runes []rune, sub string) int {
	s := string(runes)

	b := strings.Index(s, sub)
	if b < 0 {
		return 0
	}

	return runeLen(s[:b])
}

// --- Emphasis / strong (CommonMark-style delimiter run resolution) ---

// delimRun describes a maximal run of '*' or '_' delimiter characters found
// among the inline tokens. It records the index of the corresponding data
// token in the token slice along with simplified flanking information.
type delimRun struct {
	idx      int  // index into the token slice
	char     rune // '*' or '_'
	count    int  // number of remaining (unconsumed) delimiter characters
	canOpen  bool
	canClose bool
}

// resolveEmphasis matches delimiter-run data tokens into emphasis/strong tokens
// using a simplified CommonMark delimiter stack. Delimiter runs that are not
// matched remain as bare data tokens so that rules like MD037 can inspect them.
func resolveEmphasis(out []*Token, runes []rune, startCol int) []*Token {
	// Identify delimiter runs and compute flanking based on the original runes.
	var runsList []delimRun

	for idx, t := range out {
		if t.Type != TypeData || t.Text == "" {
			continue
		}

		c := []rune(t.Text)[0]
		if c != '*' && c != '_' {
			continue
		}
		// Ensure the data token is entirely a single delimiter character run.
		allSame := true

		for _, r := range t.Text {
			if r != c {
				allSame = false
				break
			}
		}

		if !allSame {
			continue
		}
		// Determine the runes immediately before and after this run in the
		// source line for flanking computation.
		startRune := t.StartColumn - startCol
		endRune := startRune + runeLen(t.Text) // one past last delimiter rune

		before, after := ' ', ' '
		if startRune-1 >= 0 && startRune-1 < len(runes) {
			before = runes[startRune-1]
		}

		if endRune >= 0 && endRune < len(runes) {
			after = runes[endRune]
		}

		beforeWS := isUnicodeWhitespace(before)
		afterWS := isUnicodeWhitespace(after)
		beforePunct := isPunct(before)
		afterPunct := isPunct(after)

		// Left-flanking: not followed by whitespace, and either not followed by
		// punctuation or preceded by whitespace/punctuation.
		leftFlanking := !afterWS && (!afterPunct || beforeWS || beforePunct)
		// Right-flanking: not preceded by whitespace, and either not preceded by
		// punctuation or followed by whitespace/punctuation.
		rightFlanking := !beforeWS && (!beforePunct || afterWS || afterPunct)

		canOpen := leftFlanking

		canClose := rightFlanking
		if c == '_' {
			// For '_', intraword emphasis is restricted.
			canOpen = leftFlanking && (!rightFlanking || beforePunct)
			canClose = rightFlanking && (!leftFlanking || afterPunct)
		}

		runsList = append(runsList, delimRun{
			idx:      idx,
			char:     c,
			count:    runeLen(t.Text),
			canOpen:  canOpen,
			canClose: canClose,
		})
	}

	if len(runsList) == 0 {
		return out
	}

	// Track how many delimiter characters have been consumed from the right end
	// of each run (when acting as an opener) and from the left end (when acting
	// as a closer). Openers consume from their right, closers from their left,
	// so nested spans line up correctly (e.g. ***x*** -> strong(emphasis(x))).
	type match struct {
		openIdx, closeIdx int // index into out
		openCol, closeCol int // 1-based start column of consumed sequence
		count             int // delimiters consumed (1 emphasis, 2 strong)
		strong            bool
	}

	var matches []match

	// openConsumed[i]/closeConsumed[i] count delimiters already taken from run i.
	openConsumed := make([]int, len(runsList))
	closeConsumed := make([]int, len(runsList))

	for ci := 0; ci < len(runsList); ci++ {
		closer := &runsList[ci]
		for closer.canClose && closer.count > 0 {
			matched := false

			for oi := ci - 1; oi >= 0; oi-- {
				opener := &runsList[oi]
				if !opener.canOpen || opener.count == 0 || opener.char != closer.char {
					continue
				}
				// CommonMark "rule of 3".
				if (opener.canClose || closer.canOpen) &&
					(opener.count+closer.count)%3 == 0 &&
					(opener.count%3 != 0 || closer.count%3 != 0) {
					continue
				}

				use := 1
				if opener.count >= 2 && closer.count >= 2 {
					use = 2
				}

				openTok := out[opener.idx]
				closeTok := out[closer.idx]
				// Opener sequence occupies the rightmost `use` chars not yet
				// consumed; closer sequence occupies the leftmost `use` chars.
				openSeqCol := openTok.EndColumn - openConsumed[oi] - use + 1
				closeSeqCol := closeTok.StartColumn + closeConsumed[ci]
				matches = append(matches, match{
					openIdx:  opener.idx,
					closeIdx: closer.idx,
					openCol:  openSeqCol,
					closeCol: closeSeqCol,
					count:    use,
					strong:   use == 2,
				})
				opener.count -= use
				closer.count -= use
				openConsumed[oi] += use
				closeConsumed[ci] += use
				matched = true

				break
			}

			if !matched {
				break
			}
		}
	}

	if len(matches) == 0 {
		return out
	}

	// Build a working list of "items": one per non-delimiter token, and one per
	// individual delimiter character (so partial consumption of a run is clean).
	type item struct {
		tok   *Token
		delim bool // true if this is a single delimiter character item
		used  bool // consumed by an emphasis/strong span
	}

	var items []item

	for _, t := range out {
		if isDelimRunToken(t) {
			rs := []rune(t.Text)
			for k, r := range rs {
				col := t.StartColumn + k
				items = append(items, item{
					tok: &Token{
						Type:        TypeData,
						StartLine:   t.StartLine,
						StartColumn: col,
						EndLine:     t.EndLine,
						EndColumn:   col,
						Text:        string(r),
					},
					delim: true,
				})
			}
		} else {
			items = append(items, item{tok: t})
		}
	}

	// Build emphasis innermost-first: largest openCol first, then smallest
	// closeCol. After building, the emphasis token replaces its span's items.
	sort.SliceStable(matches, func(a, b int) bool {
		if matches[a].openCol != matches[b].openCol {
			return matches[a].openCol > matches[b].openCol
		}

		return matches[a].closeCol < matches[b].closeCol
	})

	for _, m := range matches {
		openLo, openHi := m.openCol, m.openCol+m.count-1
		closeLo, closeHi := m.closeCol, m.closeCol+m.count-1

		var typ, seqTyp, textTyp TokenType
		if m.strong {
			typ, seqTyp, textTyp = TypeStrong, TypeStrongSequence, TypeStrongText
		} else {
			typ, seqTyp, textTyp = TypeEmphasis, TypeEmphasisSequence, TypeEmphasisText
		}

		line0 := out[m.openIdx].StartLine

		em := &Token{
			Type:        typ,
			StartLine:   line0,
			StartColumn: openLo,
			EndLine:     line0,
			EndColumn:   closeHi,
		}
		openSeq := &Token{
			Type:        seqTyp,
			StartLine:   line0,
			StartColumn: openLo,
			EndLine:     line0,
			EndColumn:   openHi,
			Text:        strings.Repeat(string(out[m.openIdx].Text[0]), m.count),
		}
		closeSeq := &Token{
			Type:        seqTyp,
			StartLine:   line0,
			StartColumn: closeLo,
			EndLine:     line0,
			EndColumn:   closeHi,
			Text:        strings.Repeat(string(out[m.closeIdx].Text[0]), m.count),
		}
		et := &Token{
			Type:        textTyp,
			StartLine:   line0,
			StartColumn: openHi + 1,
			EndLine:     line0,
			EndColumn:   closeLo - 1,
		}

		addChild(em, openSeq)
		addChild(em, et)

		// Gather text children: items strictly between openHi and closeLo that
		// are not yet used; mark opener/closer delimiter items as used and
		// collapse the span into a single emphasis item placed at the opener.
		emPos := -1

		for k := range items {
			it := &items[k]
			if it.used {
				continue
			}

			c := it.tok.StartColumn
			switch {
			case c >= openLo && c <= openHi && it.delim:
				it.used = true

				if emPos == -1 {
					emPos = k
				}
			case c >= closeLo && c <= closeHi && it.delim:
				it.used = true
			case c > openHi && c < closeLo:
				addChild(et, it.tok)
				it.used = true
			}
		}

		addChild(em, closeSeq)

		if emPos >= 0 {
			items[emPos] = item{tok: em, used: false}
		}
	}

	// Reassemble: emit remaining items, merging adjacent leftover single-char
	// delimiter items that are contiguous and share the same character.
	var result []*Token

	for k := 0; k < len(items); k++ {
		it := items[k]
		if it.used {
			continue
		}

		if it.delim {
			// Merge run of contiguous same-char delimiter items.
			ch := it.tok.Text
			startCol2 := it.tok.StartColumn
			endCol2 := it.tok.EndColumn

			text := it.tok.Text
			for k+1 < len(items) && !items[k+1].used && items[k+1].delim &&
				items[k+1].tok.Text == ch && items[k+1].tok.StartColumn == endCol2+1 {
				k++
				endCol2 = items[k].tok.EndColumn
				text += items[k].tok.Text
			}

			result = append(
				result,
				&Token{
					Type:        TypeData,
					StartLine:   it.tok.StartLine,
					StartColumn: startCol2,
					EndLine:     it.tok.EndLine,
					EndColumn:   endCol2,
					Text:        text,
				},
			)

			continue
		}

		result = append(result, it.tok)
	}

	return result
}

// isDelimRunToken reports whether t is a data token consisting entirely of a
// single repeated '*' or '_' character.
func isDelimRunToken(t *Token) bool {
	if t.Type != TypeData || t.Text == "" {
		return false
	}

	c := t.Text[0]
	if c != '*' && c != '_' {
		return false
	}

	for i := 0; i < len(t.Text); i++ {
		if t.Text[i] != c {
			return false
		}
	}

	return true
}

func isUnicodeWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v' ||
		unicode.IsSpace(r)
}

func isPunct(r rune) bool {
	return isASCIIPunct(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// --- GFM literal autolinks ---

var (
	wwwRe   = regexp.MustCompile(`(?i)\bwww\.[a-z0-9.-]+\.[a-z]{2,}[^\s<]*`)
	httpRe  = regexp.MustCompile(`(?i)\bhttps?://[a-z0-9.-]+\.[a-z]{2,}[^\s<]*`)
	emailRe = regexp.MustCompile(`(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
)

func extractLiteralAutolinks(out []*Token) []*Token {
	var result []*Token

	for _, t := range out {
		if t.Type != TypeData {
			result = append(result, t)
			continue
		}

		result = append(result, splitLiteralAutolinks(t)...)
	}

	return result
}

func splitLiteralAutolinks(t *Token) []*Token {
	text := t.Text
	// find earliest match among the three patterns
	type m struct{ lo, hi int }

	best := m{-1, -1}

	for _, re := range []*regexp.Regexp{httpRe, wwwRe, emailRe} {
		if loc := re.FindStringIndex(text); loc != nil {
			if best.lo == -1 || loc[0] < best.lo {
				best = m{loc[0], loc[1]}
			}
		}
	}

	if best.lo == -1 {
		return []*Token{t}
	}

	var result []*Token

	if best.lo > 0 {
		rs := runeLen(text[:best.lo])
		result = append(
			result,
			&Token{
				Type:        TypeData,
				StartLine:   t.StartLine,
				StartColumn: t.StartColumn,
				EndLine:     t.EndLine,
				EndColumn:   t.StartColumn + rs - 1,
				Text:        text[:best.lo],
			},
		)
	}

	matchText := text[best.lo:best.hi]
	rs := runeLen(text[:best.lo])
	re := runeLen(text[:best.hi])
	la := &Token{
		Type:        TypeLiteralAutolink,
		StartLine:   t.StartLine,
		StartColumn: t.StartColumn + rs,
		EndLine:     t.EndLine,
		EndColumn:   t.StartColumn + re - 1,
		Text:        matchText,
	}
	result = append(result, la)

	if best.hi < len(text) {
		trail := &Token{
			Type:        TypeData,
			StartLine:   t.StartLine,
			StartColumn: t.StartColumn + re,
			EndLine:     t.EndLine,
			EndColumn:   t.EndColumn,
			Text:        text[best.hi:],
		}
		result = append(result, splitLiteralAutolinks(trail)...)
	}

	return result
}
