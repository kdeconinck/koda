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

// Package source defines lightweight types for tracking positions within a `.core` source file.
//
// These types are shared across scanning, parsing, validation, and diagnostics.
// They allow the engine to report precise locations for tokens, declarations, and errors.
package loc

// Position identifies a specific location in a source file.
//
// Line and Column are both 1-based so they are easier for humans to read in diagnostics and test output.
type Position struct {
	// Line is the 1-based line number in the source file.
	Line int

	// Column is the 1-based column number in the source file.
	Column int
}

// IsValid reports whether p contains a valid human-readable source position.
//
// A position is considered valid when both the line and column are greater than zero.
func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}
