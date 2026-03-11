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

// Verify the public API of the loc package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package loc_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/loc"
)

// Verifies that [loc.Span.IsValid] accepts only spans with valid positions whose end does not come before the start.
func Test_Span_IsValid(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input loc.Span
		want  bool
	}{
		"When start and end are valid on the same line and end is after start, the span is valid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   1,
					Column: 1,
				},
				End: loc.Position{
					Line:   1,
					Column: 5,
				},
			},
			want: true,
		},
		"When start and end are equal, the span is valid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   2,
					Column: 3,
				},
				End: loc.Position{
					Line:   2,
					Column: 3,
				},
			},
			want: true,
		},
		"When end is on a later line, the span is valid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   2,
					Column: 10,
				},
				End: loc.Position{
					Line:   3,
					Column: 1,
				},
			},
			want: true,
		},
		"When start is invalid, the span is invalid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   0,
					Column: 1,
				},
				End: loc.Position{
					Line:   1,
					Column: 1,
				},
			},
			want: false,
		},
		"When end is invalid, the span is invalid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   1,
					Column: 1,
				},
				End: loc.Position{
					Line:   1,
					Column: 0,
				},
			},
			want: false,
		},
		"When end is before start on the same line, the span is invalid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   4,
					Column: 8,
				},
				End: loc.Position{
					Line:   4,
					Column: 7,
				},
			},
			want: false,
		},
		"When end is on an earlier line, the span is invalid.": {
			input: loc.Span{
				Start: loc.Position{
					Line:   5,
					Column: 1,
				},
				End: loc.Position{
					Line:   4,
					Column: 99,
				},
			},
			want: false,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := tc.input.IsValid()

			// Assert.
			assert.Equalf(t, got, tc.want, "\n\n"+
				"UT Name:  %s\n"+
				"\033[32mExpected: %t\033[0m\n"+
				"\033[31mActual:   %t\033[0m\n\n",
				tcName, tc.want, got)
		})
	}
}
