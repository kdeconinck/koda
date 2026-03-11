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

// Verify the public API of the ast package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package ast_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/ast"
)

// Verifies that [ast.TokenDefinitionKind.String] returns stable human-readable names for all exported token-definition
// kinds.
func Test_TokenDefinitionKind_String(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input ast.TokenDefinitionKind
		want  string
	}{
		"When the kind is unknown, String returns Unknown.": {
			input: ast.TokenDefinitionKindUnknown,
			want:  "Unknown",
		},
		"When the kind is literal, String returns Literal.": {
			input: ast.TokenDefinitionKindLiteral,
			want:  "Literal",
		},
		"When the kind is sequence, String returns Sequence.": {
			input: ast.TokenDefinitionKindSequence,
			want:  "Sequence",
		},
		"When the kind is enclosed-by, String returns EnclosedBy.": {
			input: ast.TokenDefinitionKindEnclosedBy,
			want:  "EnclosedBy",
		},
		"When the kind is an unsupported value, String returns Unknown.": {
			input: ast.TokenDefinitionKind(999),
			want:  "Unknown",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := tc.input.String()

			// Assert.
			assert.Equalf(t, got, tc.want, "\n\n"+
				"UT Name:  %s\n"+
				"\033[32mExpected: %s\033[0m\n"+
				"\033[31mActual:   %s\033[0m\n\n",
				tcName, tc.want, got)
		})
	}
}
