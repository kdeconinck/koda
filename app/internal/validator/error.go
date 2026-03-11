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

// Package validator validates parsed Koda configuration models.
//
// The validator performs semantic analysis on the parsed AST.
// It checks that names are unique, references can be resolved, and declarations are structurally meaningful after
// parsing has completed.
package validator

import "github.com/kdeconinck/koda/internal/loc"

// Error represents a semantic validation failure encountered in a parsed Koda language profile.
//
// Validator errors describe situations where the parsed configuration is syntactically valid but semantically invalid.
type Error struct {
	// Message contains the human-readable validation error description.
	Message string

	// Span identifies the source range where the validator detected the error.
	Span loc.Span
}

// Error returns the human-readable validator error message.
func (err Error) Error() string {
	return err.Message
}
