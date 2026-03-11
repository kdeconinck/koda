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

// TokenDefinitionKind identifies the kind of token definition.
type TokenDefinitionKind int

const (
	// TokenDefinitionKindUnknown represents an unspecified token definition kind.
	TokenDefinitionKindUnknown TokenDefinitionKind = iota

	// TokenDefinitionKindLiteral represents a fixed literal value.
	TokenDefinitionKindLiteral

	// TokenDefinitionKindSequence represents a sequence based on a named charset.
	TokenDefinitionKindSequence

	// TokenDefinitionKindEnclosedBy represents text enclosed by start and end delimiters.
	TokenDefinitionKindEnclosedBy
)

// String returns the human-readable name of k.
func (k TokenDefinitionKind) String() string {
	switch k {
	case TokenDefinitionKindLiteral:
		return "Literal"

	case TokenDefinitionKindSequence:
		return "Sequence"

	case TokenDefinitionKindEnclosedBy:
		return "EnclosedBy"

	default:
		return "Unknown"
	}
}

// TokenDefinition represents a token definition inside `SECTION TOKENS`.
//
// Exactly one logical definition kind is expected for each value, determined by Kind.
type TokenDefinition struct {
	// Kind identifies the token-definition form.
	Kind TokenDefinitionKind

	// Name is the declared token name.
	Name string

	// Value stores the literal text for a literal definition.
	Value string

	// CharSetName stores the referenced charset name for a sequence definition.
	CharSetName string

	// Start stores the opening delimiter for an enclosed definition.
	Start string

	// End stores the closing delimiter for an enclosed definition.
	End string

	// Span identifies the exact source range of the full token definition.
	Span loc.Span
}
