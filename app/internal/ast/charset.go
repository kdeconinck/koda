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

// Package ast defines the parsed syntax model of the Koda configuration language.
//
// The types in this package represent the structure produced by the parser.
// They preserve names, declaration order, and source spans so later stages such as validation and compilation can work
// with the original configuration accurately.
package ast

import "github.com/kdeconinck/koda/internal/loc"

// CharSetItemKind identifies the kind of value stored in a charset item.
type CharSetItemKind int

const (
	// CharSetItemKindUnknown represents an unspecified charset item kind.
	CharSetItemKindUnknown CharSetItemKind = iota

	// CharSetItemKindSingle represents a single character item such as '_'.
	CharSetItemKindSingle

	// CharSetItemKindRange represents a range item such as 'a'..'z'.
	CharSetItemKindRange
)

// String returns the human-readable name of k.
func (k CharSetItemKind) String() string {
	switch k {
	case CharSetItemKindSingle:
		return "Single"

	case CharSetItemKindRange:
		return "Range"

	default:
		return "Unknown"
	}
}

// CharSetItem represents a single value or range inside a charset declaration.
//
// Exactly one of the following forms is expected:
//
//   - A single character value.
//   - A character range with both start and end populated.
type CharSetItem struct {
	// Kind identifies whether the item is a single value or a range.
	Kind CharSetItemKind

	// Value stores the character for a single-value item.
	Value rune

	// Start stores the first character of a range item.
	Start rune

	// End stores the last character of a range item.
	End rune

	// Span identifies the exact source range of the charset item.
	Span loc.Span
}

// CharSet represents a named charset declaration in a Koda language profile.
//
// Example:
//
//	DEFINE CHARSET "digits" VALUES ['0'..'9']
type CharSet struct {
	// Name is the declared charset name.
	Name string

	// Items contains the declared values and ranges in source order.
	Items []CharSetItem

	// Span identifies the exact source range of the full charset declaration.
	Span loc.Span
}
