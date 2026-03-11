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

// TokenSection represents the `SECTION TOKENS` block of a language profile.
type TokenSection struct {
	// Definitions contains the token definitions declared in this section.
	Definitions []TokenDefinition

	// Span identifies the exact source range of the full token section.
	Span loc.Span
}

// RuleSection represents the `SECTION RULES` block of a language profile.
type RuleSection struct {
	// Rules contains the rules declared in this section.
	Rules []Rule

	// Span identifies the exact source range of the full rule section.
	Span loc.Span
}

// LanguageProfile represents a complete parsed Koda language profile.
//
// Example:
//
//	LANG "json" EXTENSIONS ["json"] {
//	    ...
//	}
type LanguageProfile struct {
	// Name is the declared language name.
	Name string

	// Extensions contains the declared file extensions in source order.
	Extensions []string

	// CharSets contains the charset declarations in source order.
	CharSets []CharSet

	// Tokens contains the parsed token section.
	Tokens TokenSection

	// Rules contains the parsed rule section.
	Rules RuleSection

	// Span identifies the exact source range of the full language profile.
	Span loc.Span
}
