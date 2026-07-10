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
	"regexp"
	"strings"

	"github.com/fuldaxxx/go-markdownlint/internal/mdhelpers"
	mm "github.com/fuldaxxx/go-markdownlint/internal/micromark"
)

// DefInfo holds a definition's line index and destination.
type DefInfo struct {
	LineIndex   int
	Destination string
}

// DupDef holds a duplicate definition's normalized label and line index.
type DupDef struct {
	Label     string
	LineIndex int
}

// ReferenceLinkImageData holds information about reference links and images.
type ReferenceLinkImageData struct {
	References            map[string][][3]int // normalized label -> list of [lineIndex, colIndex, length]
	Shortcuts             map[string][][3]int
	Definitions           map[string]DefInfo
	DuplicateDefinitions  []DupDef
	DefinitionLineIndices []int
}

var refWhitespaceRe = regexp.MustCompile(`\s+`)

func normalizeReference(s string) string {
	return refWhitespaceRe.ReplaceAllString(strings.TrimSpace(strings.ToLower(s)), " ")
}

func refTokenText(t *mm.Token) string {
	if t == nil {
		return ""
	}

	var sb strings.Builder

	for _, c := range t.Children {
		if c.Type != mm.TypeBlockQuotePrefix {
			sb.WriteString(c.Text)
		}
	}

	return sb.String()
}

// GetReferenceLinkImageData scans tokens (a flat token list) and returns
// reference/shortcut/definition data. Ported from helpers.cjs
// getReferenceLinkImageData.
func GetReferenceLinkImageData(flat []*mm.Token) *ReferenceLinkImageData {
	data := &ReferenceLinkImageData{
		References:  map[string][][3]int{},
		Shortcuts:   map[string][][3]int{},
		Definitions: map[string]DefInfo{},
	}
	addReference := func(token *mm.Token, label string, isShortcut bool) {
		datum := [3]int{token.StartLine - 1, token.StartColumn - 1, len(token.Text)}
		ref := normalizeReference(label)

		dict := data.References
		if isShortcut {
			dict = data.Shortcuts
		}

		dict[ref] = append(dict[ref], datum)
	}

	filtered := mdhelpers.FilterFlat(flat, []mm.TokenType{
		mm.TypeDefinition,
		mm.TypeGfmFootnoteDefinition,
		mm.TypeDefinitionLabelString,
		mm.TypeGfmFootnoteDefinitionLabelString,
		mm.TypeGfmFootnoteCall,
		mm.TypeImage,
		mm.TypeLink,
		mm.TypeUndefinedReferenceCollapsed,
		mm.TypeUndefinedReferenceFull,
		mm.TypeUndefinedReferenceShortcut,
	}, false)

	for _, token := range filtered {
		switch token.Type {
		case mm.TypeDefinition, mm.TypeGfmFootnoteDefinition:
			for i := token.StartLine; i <= token.EndLine; i++ {
				data.DefinitionLineIndices = append(data.DefinitionLineIndices, i-1)
			}
		case mm.TypeDefinitionLabelString, mm.TypeGfmFootnoteDefinitionLabelString:
			labelPrefix := ""
			if token.Type == mm.TypeGfmFootnoteDefinitionLabelString {
				labelPrefix = "^"
			}

			ref := normalizeReference(labelPrefix + token.Text)
			if _, ok := data.Definitions[ref]; ok {
				data.DuplicateDefinitions = append(
					data.DuplicateDefinitions,
					DupDef{ref, token.StartLine - 1},
				)
			} else {
				parent := mdhelpers.GetParentOfType(token, []mm.TokenType{mm.TypeDefinition})
				dest := ""

				if parent != nil {
					ds := mdhelpers.GetDescendantsByType([]*mm.Token{parent}, [][]mm.TokenType{
						{mm.TypeDefinitionDestination},
						{mm.TypeDefinitionDestinationRaw, mm.TypeDefinitionDestinationString},
					})
					if len(ds) > 0 {
						dest = ds[0].Text
					}
				}

				data.Definitions[ref] = DefInfo{LineIndex: token.StartLine - 1, Destination: dest}
			}
		case mm.TypeGfmFootnoteCall, mm.TypeImage, mm.TypeLink:
			isShortcut := len(token.Children) == 1
			isFullOrCollapsed := len(token.Children) == 2 && !anyChildType(token, mm.TypeResource)
			labelText := firstDescendant(token, mm.TypeLabel, mm.TypeLabelText)
			referenceString := firstDescendant(token, mm.TypeReference, mm.TypeReferenceString)
			label := refTokenText(labelText)

			if !isShortcut && !isFullOrCollapsed {
				var marker, str *mm.Token

				for _, c := range token.Children {
					switch c.Type {
					case mm.TypeGfmFootnoteCallMarker:
						marker = c
					case mm.TypeGfmFootnoteCallString:
						str = c
					}
				}

				if marker != nil && str != nil {
					label = marker.Text + str.Text
					isShortcut = true
				}
			}

			if isShortcut || isFullOrCollapsed {
				refLabel := refTokenText(referenceString)
				if refLabel == "" {
					refLabel = label
				}

				addReference(token, refLabel, isShortcut)
			}
		case mm.TypeUndefinedReferenceCollapsed,
			mm.TypeUndefinedReferenceFull,
			mm.TypeUndefinedReferenceShortcut:
			ud := firstDescendant(token, mm.TypeUndefinedReference)
			label := ""

			if ud != nil {
				var sb strings.Builder
				for _, c := range ud.Children {
					sb.WriteString(c.Text)
				}

				label = sb.String()
			}

			addReference(token, label, token.Type == mm.TypeUndefinedReferenceShortcut)
		}
	}

	return data
}

func anyChildType(token *mm.Token, t mm.TokenType) bool {
	for _, c := range token.Children {
		if c.Type == t {
			return true
		}
	}

	return false
}

// firstDescendant follows a single-type path and returns the first descendant.
func firstDescendant(token *mm.Token, path ...mm.TokenType) *mm.Token {
	steps := make([][]mm.TokenType, len(path))
	for i, p := range path {
		steps[i] = []mm.TokenType{p}
	}

	ds := mdhelpers.GetDescendantsByType([]*mm.Token{token}, steps)
	if len(ds) == 0 {
		return nil
	}

	return ds[0]
}
