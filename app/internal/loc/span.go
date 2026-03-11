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

// Span identifies a half-open range in a source file.
//
// Start points to where the item begins, and End points to the position immediately after the item ends.
// This matches the common start-inclusive, end-exclusive convention used by scanners and parsers.
type Span struct {
	// Start is the first position included in the span.
	Start Position

	// End is the first position after the span.
	End Position
}

// IsValid reports whether s contains a valid source span.
//
// A span is considered valid when both positions are valid and the end does not come before the start.
func (s Span) IsValid() bool {
	if !s.Start.IsValid() || !s.End.IsValid() {
		return false
	}

	if s.End.Line < s.Start.Line {
		return false
	}

	if s.End.Line == s.Start.Line && s.End.Column < s.Start.Column {
		return false
	}

	return true
}
