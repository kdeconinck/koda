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

import "fmt"

// TbSpy is a strict test double for [assert.TB].
//
// Unlike embedding [testing.TB], this type implements only the methods required by the public API.
// That keeps the spy small and makes the test intent explicit.
// The spy records the last failure message passed to Fatalf. It does not actually terminate the test, which allows the
// test to inspect results.
type TbSpy struct {
	// The last failure message passed to Fatalf.
	failureMsg string

	// The number of times the Helper method was called.
	helperCalls int
}

// Fatalf records the formatted failure message instead of stopping the test.
func (tb *TbSpy) Fatalf(format string, args ...any) {
	tb.failureMsg = fmt.Sprintf(format, args...)
}

// Helper records that the helper marker was invoked.
func (tb *TbSpy) Helper() {
	tb.helperCalls += 1
}
