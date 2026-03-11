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

// Package parser transforms lexical tokens into parsed Koda configuration models.
//
// The parser performs syntactic analysis only.
// It validates that tokens appear in a structurally correct order and shape, but it does not perform semantic
// validation such as checking for duplicate names or unresolved references.
package parser

import "github.com/kdeconinck/koda/internal/loc"

// Error represents a syntactic failure encountered while parsing a `.core` token stream.
//
// Parser errors describe situations where the token sequence does not match the grammar of the Koda configuration
// language.
type Error struct {
	// Message is the human-readable error description.
	Message string

	// Span is the precise source range where the error occurred.
	Span loc.Span
}

// Error returns the human-readable parser error message.
func (err Error) Error() string {
	return err.Message
}
