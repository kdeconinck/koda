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

// Package engine executes compiled Koda language profiles against target source files.
//
// The engine tokenizes target source text using compiled token definitions and then evaluates compiled structural rules
// to produce diagnostics.
package engine

import (
	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/loc"
)

// Token represents a token matched in a target source file.
//
// Runtime tokens are different from `.core` scanner tokens.
// These values are produced by applying compiled Koda token definitions to real source input.
type Token struct {
	// DefinitionID identifies the compiled token definition that produced this token.
	DefinitionID compiler.TokenDefinitionID

	// Name is the original token-definition name.
	Name string

	// Lexeme is the exact source text matched for this token.
	Lexeme string

	// Span identifies the exact source range of the matched token.
	Span loc.Span
}
