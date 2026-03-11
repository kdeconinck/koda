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

import "github.com/kdeconinck/koda/internal/loc"

// Token represents a single scanned token from a `.core` source file.
//
// A token keeps its broad category separate from the exact source text that produced it.
// This makes scanner output easier to inspect and allows the parser to reason about both token type and original
// lexeme.
type Token struct {
	// Kind is the broad category of the token.
	Kind Kind

	// Lexeme is the exact source text associated with the token, without any additional interpretation.
	Lexeme string

	// Span is the precise source range where the token was found.
	Span loc.Span
}
