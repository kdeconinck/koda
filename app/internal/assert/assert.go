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

// Package assert provides small helper functions for writing tests.
//
// The helpers in this package are intentionally minimal and are designed to work with Go's standard [testing] package
// without introducing any external dependencies.
//
// Note that these helpers accept a small interface (TB) rather than [testing.TB].
// The [testing.TB] interface includes an unexported method, which prevents custom implementations outside the standard
// library. By accepting a minimal interface, we can still work with [testing.T] and [testing.B], while also enabling
// strict test doubles.
package assert

import _ "testing"

// TB is the minimal subset of [testing.T] and [testing.B] needed by this package.
//
// Both [testing.T] and [testing.B] satisfy this interface.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}
