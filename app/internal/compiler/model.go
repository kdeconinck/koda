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

// Package compiler defines the compiled runtime model of the Koda configuration language.
//
// The types in this package represent a validated language profile in a runtime-friendly form.
// References are resolved, character sets are normalized, and token definitions and rules are prepared for efficient
// execution by the engine.
package compiler

import "github.com/kdeconinck/koda/internal/loc"

// CharSetID identifies a compiled charset within a compiled language profile.
type CharSetID int

// TokenDefinitionID identifies a compiled token definition within a compiled language profile.
type TokenDefinitionID int

// TokenDefinitionKind identifies the kind of compiled token definition.
type TokenDefinitionKind int

const (
	// TokenDefinitionKindUnknown represents an unspecified token-definition kind.
	TokenDefinitionKindUnknown TokenDefinitionKind = iota

	// TokenDefinitionKindLiteral represents a fixed literal token definition.
	TokenDefinitionKindLiteral

	// TokenDefinitionKindSequence represents a token definition backed by a
	// compiled charset.
	TokenDefinitionKindSequence

	// TokenDefinitionKindEnclosedBy represents a token definition enclosed by
	// start and end delimiters.
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

// ConstraintKind identifies the kind of compiled structural constraint.
type ConstraintKind int

const (
	// ConstraintKindUnknown represents an unspecified constraint kind.
	ConstraintKindUnknown ConstraintKind = iota

	// ConstraintKindMustBeFollowedBy requires the matched token to be followed by the referenced token.
	ConstraintKindMustBeFollowedBy

	// ConstraintKindCannotBeFollowedBy forbids the matched token from being followed by the referenced token.
	ConstraintKindCannotBeFollowedBy
)

// String returns the human-readable name of k.
func (k ConstraintKind) String() string {
	switch k {
	case ConstraintKindMustBeFollowedBy:
		return "MustBeFollowedBy"

	case ConstraintKindCannotBeFollowedBy:
		return "CannotBeFollowedBy"

	default:
		return "Unknown"
	}
}

// CharSet represents a compiled charset ready for runtime matching.
//
// Character ranges from the parsed AST are normalized into an explicit set of allowed characters so membership checks
// are simple and direct at runtime.
type CharSet struct {
	// ID is the stable compiled identifier of the charset.
	ID CharSetID

	// Name is the original declared charset name.
	Name string

	// Characters contains the full normalized set of allowed characters.
	Characters map[rune]struct{}

	// Span identifies the source range of the original charset declaration.
	Span loc.Span
}

// TokenDefinition represents a compiled token definition ready for runtime matching.
//
// All external references are already resolved. Sequence definitions reference a compiled charset ID rather than a
// charset name.
type TokenDefinition struct {
	// ID is the stable compiled identifier of the token definition.
	ID TokenDefinitionID

	// Kind identifies the token-definition form.
	Kind TokenDefinitionKind

	// Name is the original declared token name.
	Name string

	// Value stores the fixed literal value for literal token definitions.
	Value string

	// CharSetID stores the resolved charset reference for sequence token definitions.
	CharSetID CharSetID

	// Start stores the opening delimiter for enclosed token definitions.
	Start string

	// End stores the closing delimiter for enclosed token definitions.
	End string

	// Span identifies the source range of the original token definition.
	Span loc.Span
}

// Constraint represents a compiled rule constraint ready for runtime evaluation.
//
// Token references are resolved to stable token-definition IDs.
type Constraint struct {
	// Kind identifies the structural constraint form.
	Kind ConstraintKind

	// TokenID stores the resolved token reference used by the constraint.
	TokenID TokenDefinitionID

	// Span identifies the source range of the original constraint declaration.
	Span loc.Span
}

// Rule represents a compiled structural rule ready for runtime evaluation.
//
// Token references are resolved to stable token-definition IDs so the engine can evaluate rules without repeated string
// lookups.
type Rule struct {
	// Name is the original declared rule name.
	Name string

	// MatchTokenID identifies the compiled token definition that triggers the rule.
	MatchTokenID TokenDefinitionID

	// Constraint is the compiled structural constraint of the rule.
	Constraint Constraint

	// ErrorMessage is the diagnostic message emitted when the rule fails.
	ErrorMessage string

	// Span identifies the source range of the original rule declaration.
	Span loc.Span
}

// LanguageProfile represents a compiled Koda language profile ready for use by the runtime engine.
//
// Names are preserved for diagnostics, while cross-references are resolved to stable IDs for efficient execution.
type LanguageProfile struct {
	// Name is the declared language name.
	Name string

	// Extensions contains the declared file extensions in source order.
	Extensions []string

	// CharSets contains the compiled charsets in declaration order.
	CharSets []CharSet

	// TokenDefinitions contains the compiled token definitions in declaration order.
	TokenDefinitions []TokenDefinition

	// Rules contains the compiled rules in declaration order.
	Rules []Rule

	// Span identifies the source range of the original language profile.
	Span loc.Span
}
