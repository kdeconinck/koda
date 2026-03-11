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

// ConstraintKind identifies the kind of structural rule constraint.
type ConstraintKind int

const (
	// ConstraintKindUnknown represents an unspecified constraint kind.
	ConstraintKindUnknown ConstraintKind = iota

	// ConstraintKindMustBeFollowedBy requires the matched token to be followed
	// by the referenced token.
	ConstraintKindMustBeFollowedBy

	// ConstraintKindCannotBeFollowedBy forbids the matched token from being
	// followed by the referenced token.
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

// Constraint represents the structural condition declared inside a rule block.
type Constraint struct {
	// Kind identifies the constraint form.
	Kind ConstraintKind

	// TokenName is the referenced token name used by the constraint.
	TokenName string

	// Span identifies the exact source range of the constraint declaration.
	Span loc.Span
}

// Rule represents a named rule inside `SECTION RULES`.
//
// Example:
//
//	RULE "pair_format" {
//	    MATCH "key"
//	    MUST_BE_FOLLOWED_BY "colon"
//	    ERROR "JSON keys must be followed by a colon."
//	}
type Rule struct {
	// Name is the declared rule name.
	Name string

	// MatchTokenName is the token name targeted by the rule.
	MatchTokenName string

	// Constraint is the structural rule condition.
	Constraint Constraint

	// ErrorMessage is the diagnostic message declared by the rule.
	ErrorMessage string

	// Span identifies the exact source range of the full rule block.
	Span loc.Span
}
