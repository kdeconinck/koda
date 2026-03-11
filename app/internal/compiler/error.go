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

// Error represents a compilation failure encountered while compiling a
// validated Koda language profile.
//
// Compiler errors describe situations where the parsed and validated AST cannot
// be transformed into the compiled runtime model.
type Error struct {
	// Message contains the human-readable compiler error description.
	Message string

	// Span identifies the source range where compilation failed.
	Span loc.Span
}

// Error returns the human-readable compiler error message.
func (err Error) Error() string {
	return err.Message
}
