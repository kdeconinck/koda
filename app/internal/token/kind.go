// =====================================================================================================================
// = LICENSE:       Copyright (c) 2026 Kevin De Coninck
// =
// =                Permission is hereby granted, free of charge, to any person
// =                obtaining a copy of this software and associated documentation
// =                files (the "Software"), to deal in the Software without
// =                restriction, including without limitation the rights to use,
// =                copy, modify, merge, publish, distribute, sublicense, and/or sell
// =                copies of the Software, and to permit persons to whom the
// =                Software is furnished to do so, subject to the following
// =                conditions:
// =
// =                The above copyright notice and this permission notice shall be
// =                included in all copies or substantial portions of the Software.
// =
// =                THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// =                EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
// =                OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// =                NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// =                HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// =                WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// =                FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// =                OTHER DEALINGS IN THE SOFTWARE.
// =====================================================================================================================

// Package token defines the lexical token model used by the Koda scanner and parser.
//
// Tokens represent the smallest meaningful pieces of a `.core` configuration file.
// The scanner produces tokens, and the parser consumes them to build structured language profiles.
package token

// Kind identifies the broad category of a scanned token.
type Kind int

const (
	// KindUnknown represents an unspecified token kind.
	KindUnknown Kind = iota

	// KindKeyword represents a reserved word in the Koda configuration language, such as LANG, DEFINE, or SECTION.
	KindKeyword

	// KindLiteral represents a double-quoted string literal, such as "json" or "brace_open".
	KindLiteral

	// KindCharacter represents a single-quoted character literal, such as 'a', '_', or '"'.
	KindCharacter

	// KindSymbol represents a structural symbol or operator, such as {, }, [, ], ,, or ...
	KindSymbol

	// KindIdentifier represents an unquoted non-keyword name, such as LITERAL, SEQUENCE, or ENCLOSED_BY.
	KindIdentifier
)

// String returns the human-readable name of k.
func (k Kind) String() string {
	switch k {
	case KindKeyword:
		return "Keyword"

	case KindLiteral:
		return "Literal"

	case KindCharacter:
		return "Character"

	case KindSymbol:
		return "Symbol"

	case KindIdentifier:
		return "Identifier"

	default:
		return "Unknown"
	}
}
