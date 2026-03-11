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

// Verify the public API of the assert package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package assert_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
)

// Verifies that [assert.Equalf] reports failures only when values differ, and that it forwards the caller-provided
// format string and arguments to Fatalf.
func Test_Equalf(t *testing.T) {
	t.Parallel()

	const msgFmt = "Not equal - got %t, want %t."

	for tcName, tc := range map[string]struct {
		gotInput  bool
		wantInput bool
		wantMsg   string
		wantHelp  int
	}{
		"When the compared values are equal, no failure is reported.": {
			gotInput:  true,
			wantInput: true,
			wantMsg:   "",
			wantHelp:  1,
		},
		"When the compared values are not equal, a failure is reported.": {
			gotInput:  true,
			wantInput: false,
			wantMsg:   "Not equal - got true, want false.",
			wantHelp:  1,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			tbSpy := new(TbSpy)

			// Act.
			assert.Equalf(tbSpy, tc.gotInput, tc.wantInput, msgFmt, tc.gotInput, tc.wantInput)

			// Assert.
			if tbSpy.failureMsg != tc.wantMsg {
				t.Fatalf("Failure message = %q, want %q", tbSpy.failureMsg, tc.wantMsg)
			}

			if tbSpy.helperCalls != tc.wantHelp {
				t.Fatalf("Helper calls = %d, want %d", tbSpy.helperCalls, tc.wantHelp)
			}
		})
	}
}
