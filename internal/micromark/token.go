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

// Package micromark provides a Markdown tokenizer that produces a token tree
// compatible with the micromark token model used by markdownlint. Tokens carry
// 1-based line/column spans, the source text for the span, and parent/child
// links. A flattened depth-first token list is attached to the parsed Document
// for fast type filtering.
package micromark

// TokenType is the string name of a token (e.g. "atxHeading"). Keeping the
// string value identical lets rules and parser tests compare directly.
type TokenType = string

// Token type constants. These mirror the micromark token vocabulary referenced
// by the markdownlint rules and helpers.
const (
	// Document root (synthetic).
	TypeData TokenType = "data"

	// Headings.
	TypeAtxHeading         TokenType = "atxHeading"
	TypeAtxHeadingSequence TokenType = "atxHeadingSequence"
	TypeAtxHeadingText     TokenType = "atxHeadingText"
	TypeSetextHeading      TokenType = "setextHeading"
	TypeSetextHeadingText  TokenType = "setextHeadingText"
	TypeSetextHeadingLine  TokenType = "setextHeadingLine"

	// Code.
	TypeCodeFenced          TokenType = "codeFenced"
	TypeCodeFencedFence     TokenType = "codeFencedFence"
	TypeCodeFencedFenceInfo TokenType = "codeFencedFenceInfo"
	TypeCodeFencedFenceMeta TokenType = "codeFencedFenceMeta"
	TypeCodeFencedFenceSeq  TokenType = "codeFencedFenceSequence"
	TypeCodeFlowValue       TokenType = "codeFlowValue"
	TypeCodeIndented        TokenType = "codeIndented"
	TypeCodeText            TokenType = "codeText"
	TypeCodeTextData        TokenType = "codeTextData"
	TypeCodeTextPadding     TokenType = "codeTextPadding"
	TypeCodeTextSequence    TokenType = "codeTextSequence"

	// Lists.
	TypeListOrdered              TokenType = "listOrdered"
	TypeListUnordered            TokenType = "listUnordered"
	TypeListItemPrefix           TokenType = "listItemPrefix"
	TypeListItemMarker           TokenType = "listItemMarker"
	TypeListItemValue            TokenType = "listItemValue"
	TypeListItemIndent           TokenType = "listItemIndent"
	TypeListItemPrefixWhitespace TokenType = "listItemPrefixWhitespace"

	// Blockquote.
	TypeBlockQuote                 TokenType = "blockQuote"
	TypeBlockQuoteMarker           TokenType = "blockQuoteMarker"
	TypeBlockQuotePrefix           TokenType = "blockQuotePrefix"
	TypeBlockQuotePrefixWhitespace TokenType = "blockQuotePrefixWhitespace"

	// Inline + structural.
	TypeEmphasis         TokenType = "emphasis"
	TypeEmphasisSequence TokenType = "emphasisSequence"
	TypeEmphasisText     TokenType = "emphasisText"
	TypeStrong           TokenType = "strong"
	TypeStrongSequence   TokenType = "strongSequence"
	TypeStrongText       TokenType = "strongText"
	TypeParagraph        TokenType = "paragraph"
	TypeContent          TokenType = "content"
	TypeThematicBreak    TokenType = "thematicBreak"
	TypeLineEnding       TokenType = "lineEnding"
	TypeLineEndingBlank  TokenType = "lineEndingBlank"
	TypeLinePrefix       TokenType = "linePrefix"
	TypeWhitespace       TokenType = "whitespace"

	// Links/images.
	TypeLink                        TokenType = "link"
	TypeImage                       TokenType = "image"
	TypeLabel                       TokenType = "label"
	TypeLabelText                   TokenType = "labelText"
	TypeLabelImage                  TokenType = "labelImage"
	TypeLabelLink                   TokenType = "labelLink"
	TypeReference                   TokenType = "reference"
	TypeReferenceString             TokenType = "referenceString"
	TypeResource                    TokenType = "resource"
	TypeResourceDestination         TokenType = "resourceDestination"
	TypeResourceDestinationString   TokenType = "resourceDestinationString"
	TypeResourceDestinationRaw      TokenType = "resourceDestinationRaw"
	TypeResourceDestinationLiteral  TokenType = "resourceDestinationLiteral"
	TypeResourceTitle               TokenType = "resourceTitle"
	TypeResourceTitleString         TokenType = "resourceTitleString"
	TypeDefinition                  TokenType = "definition"
	TypeDefinitionLabelString       TokenType = "definitionLabelString"
	TypeDefinitionDestination       TokenType = "definitionDestination"
	TypeDefinitionDestinationRaw    TokenType = "definitionDestinationRaw"
	TypeDefinitionDestinationString TokenType = "definitionDestinationString"

	// Autolinks.
	TypeAutolink             TokenType = "autolink"
	TypeAutolinkEmail        TokenType = "autolinkEmail"
	TypeAutolinkProtocol     TokenType = "autolinkProtocol"
	TypeLiteralAutolink      TokenType = "literalAutolink"
	TypeLiteralAutolinkEmail TokenType = "literalAutolinkEmail"
	TypeLiteralAutolinkHTTP  TokenType = "literalAutolinkHttp"
	TypeLiteralAutolinkWww   TokenType = "literalAutolinkWww"

	// HTML.
	TypeHTMLFlow     TokenType = "htmlFlow"
	TypeHTMLFlowData TokenType = "htmlFlowData"
	TypeHTMLText     TokenType = "htmlText"
	TypeHTMLTextData TokenType = "htmlTextData"

	// Tables (GFM).
	TypeTable             TokenType = "table"
	TypeTableHead         TokenType = "tableHead"
	TypeTableHeader       TokenType = "tableHeader"
	TypeTableRow          TokenType = "tableRow"
	TypeTableData         TokenType = "tableData"
	TypeTableContent      TokenType = "tableContent"
	TypeTableDelimiter    TokenType = "tableDelimiter"
	TypeTableDelimiterRow TokenType = "tableDelimiterRow"
	TypeTableCellDivider  TokenType = "tableCellDivider"

	// Footnotes (GFM).
	TypeGfmFootnoteDefinition            TokenType = "gfmFootnoteDefinition"
	TypeGfmFootnoteDefinitionLabelString TokenType = "gfmFootnoteDefinitionLabelString"
	TypeGfmFootnoteDefinitionIndent      TokenType = "gfmFootnoteDefinitionIndent"
	TypeGfmFootnoteCall                  TokenType = "gfmFootnoteCall"
	TypeGfmFootnoteCallMarker            TokenType = "gfmFootnoteCallMarker"
	TypeGfmFootnoteCallString            TokenType = "gfmFootnoteCallString"

	// Math.
	TypeMathFlow     TokenType = "mathFlow"
	TypeMathText     TokenType = "mathText"
	TypeMathTextData TokenType = "mathTextData"

	// Synthetic (markdownlint-only): undefined references.
	TypeUndefinedReference          TokenType = "undefinedReference"
	TypeUndefinedReferenceShortcut  TokenType = "undefinedReferenceShortcut"
	TypeUndefinedReferenceFull      TokenType = "undefinedReferenceFull"
	TypeUndefinedReferenceCollapsed TokenType = "undefinedReferenceCollapsed"

	// Character escapes/references.
	TypeCharacterEscape      TokenType = "characterEscape"
	TypeCharacterEscapeValue TokenType = "characterEscapeValue"
	TypeCharacterReference   TokenType = "characterReference"
	TypeHardBreakEscape      TokenType = "hardBreakEscape"
)

// Token is a node in the micromark-compatible token tree. Line and column
// values are 1-based. Text is the source substring covered by the token.
type Token struct {
	Type        TokenType
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
	Text        string
	Children    []*Token
	Parent      *Token

	// inHTMLFlow marks tokens produced by the HTML-flow reparse. Helpers use it
	// to include or exclude Markdown embedded inside HTML blocks.
	inHTMLFlow bool
}

// InHTMLFlow reports whether the token was produced by the HTML-flow reparse.
func (t *Token) InHTMLFlow() bool { return t.inHTMLFlow }

// Document is the result of parsing: the top-level tokens plus a flattened
// depth-first list of every token for fast filtering.
type Document struct {
	Children []*Token
	Flat     []*Token
}
