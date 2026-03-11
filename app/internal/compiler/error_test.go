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

// Verify the public API of the compiler package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package compiler_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/loc"
)

// Verifies that [compiler.Error.Error] returns the stored error message.
func Test_Error_Error(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input compiler.Error
		want  string
	}{
		"When the error contains a message, Error returns that message.": {
			input: compiler.Error{
				Message: "Unknown token reference `colon`.",
				Span: loc.Span{
					Start: loc.Position{
						Line:   1,
						Column: 1,
					},
					End: loc.Position{
						Line:   1,
						Column: 6,
					},
				},
			},
			want: "Unknown token reference `colon`.",
		},
		"When the error contains no message, Error returns an empty string.": {
			input: compiler.Error{},
			want:  "",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			got := tc.input.Error()

			assert.Equalf(t, got, tc.want, "\n\n"+
				"UT Name:  %s\n"+
				"\033[32mExpected: %q\033[0m\n"+
				"\033[31mActual:   %q\033[0m\n\n",
				tcName, tc.want, got)
		})
	}
}
