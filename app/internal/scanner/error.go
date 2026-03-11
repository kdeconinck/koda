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

// Package scanner converts raw `.core` configuration source text into lexical tokens.
//
// The scanner performs lexical analysis only.
// It does not interpret the meaning of tokens or validate their structural relationships.
package scanner

import "github.com/kdeconinck/koda/internal/loc"

// Error represents a lexical failure encountered while scanning a `.core` file.
//
// Scanner errors describe situations where the source text cannot be converted into a valid token stream, such as an
// unterminated literal, an invalid range operator, or an unexpected character.
type Error struct {
	// Message is the human-readable error description.
	Message string

	// Span is the precise source range where the error occurred.
	Span loc.Span
}

// Error returns the human-readable scanner error message.
func (err Error) Error() string {
	return err.Message
}
