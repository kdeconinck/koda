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
	"github.com/kdeconinck/koda/internal/ast"
	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/parser"
	"github.com/kdeconinck/koda/internal/scanner"
	"github.com/kdeconinck/koda/internal/validator"
)

// Returns a validated language profile for the provided Koda source text.
// Scanner, parser, and validator setup are kept inside tests so the compiler is still exercised through its public API.
func validatedSource(t *testing.T, src string) ast.LanguageProfile {
	t.Helper()

	tokens, scanErr := scanner.Scan(src)

	assert.Nilf(t, scanErr, "\n\n"+
		"UT Name:  Compiler setup for compiler test.\n"+
		"\033[32mExpected (Scanner error): <nil>\033[0m\n"+
		"\033[31mActual (Scanner error):   %v\033[0m\n\n",
		scanErr)

	profile, parseErr := parser.Parse(tokens)

	assert.Nilf(t, parseErr, "\n\n"+
		"UT Name:  Compiler setup for compiler test.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		parseErr)

	validateErr := validator.Validate(profile)

	assert.Nilf(t, validateErr, "\n\n"+
		"UT Name:  Compiler setup for compiler test.\n"+
		"\033[32mExpected (Validator error): <nil>\033[0m\n"+
		"\033[31mActual (Validator error):   %v\033[0m\n\n",
		validateErr)

	return profile
}

// Verifies that [compiler.Compile] accepts a valid language profile and resolves declarations into the compiled runtime
// model.
func Test_Compile_AcceptsValidLanguageProfile(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "digits" VALUES ['0'..'9']
	DEFINE CHARSET "alpha" VALUES ['a'..'z', 'A'..'Z', '_']

	SECTION TOKENS {
		DEFINE "brace_open" LITERAL "{"
		DEFINE "brace_close" LITERAL "}"
		DEFINE "colon" LITERAL ":"
		DEFINE "comma" LITERAL ","
		DEFINE "number" SEQUENCE "digits"
		DEFINE "key" ENCLOSED_BY '"' '"'
	}

	SECTION RULES {
		RULE "no_trailing_comma" {
			MATCH "comma"
			CANNOT_BE_FOLLOWED_BY "brace_close"
			ERROR "Trailing commas break standard JSON."
		}

		RULE "pair_format" {
			MATCH "key"
			MUST_BE_FOLLOWED_BY "colon"
			ERROR "JSON keys must be followed by a colon."
		}
	}
}
`

	// Arrange.
	profile := validatedSource(t, input)

	// Act.
	got := compiler.Compile(profile)

	// Assert.
	assert.Equalf(t, got.Name, "json", "\n\n"+
		"UT Name:         When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Name): %q\033[0m\n"+
		"\033[31mActual (Name):   %q\033[0m\n\n",
		"json", got.Name)

	assert.Equalf(t, len(got.CharSets), 2, "\n\n"+
		"UT Name:                  When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Charset count): %d\033[0m\n"+
		"\033[31mActual (Charset count):   %d\033[0m\n\n",
		2, len(got.CharSets))

	assert.Equalf(t, len(got.TokenDefinitions), 6, "\n\n"+
		"UT Name:                When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Token count): %d\033[0m\n"+
		"\033[31mActual (Token count):   %d\033[0m\n\n",
		6, len(got.TokenDefinitions))

	assert.Equalf(t, len(got.Rules), 2, "\n\n"+
		"UT Name:               When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Rule count): %d\033[0m\n"+
		"\033[31mActual (Rule count):   %d\033[0m\n\n",
		2, len(got.Rules))

	_, hasZero := got.CharSets[0].Characters['0']

	assert.Equalf(t, hasZero, true, "\n\n"+
		"UT Name:                  When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected ('0' in digits): %t\033[0m\n"+
		"\033[31mActual ('0' in digits):   %t\033[0m\n\n",
		true, hasZero)

	_, hasNine := got.CharSets[0].Characters['9']

	assert.Equalf(t, hasNine, true, "\n\n"+
		"UT Name:                  When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected ('9' in digits): %t\033[0m\n"+
		"\033[31mActual ('9' in digits):   %t\033[0m\n\n",
		true, hasNine)

	assert.Equalf(t, got.TokenDefinitions[4].Kind, compiler.TokenDefinitionKindSequence, "\n\n"+
		"UT Name:                  When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Sequence kind): %s\033[0m\n"+
		"\033[31mActual (Sequence kind):   %s\033[0m\n\n",
		compiler.TokenDefinitionKindSequence.String(), got.TokenDefinitions[4].Kind.String())

	assert.Equalf(t, got.TokenDefinitions[4].CharSetID, compiler.CharSetID(0), "\n\n"+
		"UT Name:                        When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Sequence charset ID): %d\033[0m\n"+
		"\033[31mActual (Sequence charset ID):   %d\033[0m\n\n",
		compiler.CharSetID(0), got.TokenDefinitions[4].CharSetID)

	assert.Equalf(t, got.Rules[1].Constraint.Kind, compiler.ConstraintKindMustBeFollowedBy, "\n\n"+
		"UT Name:                    When the input is valid, compilation succeeds.\n"+
		"\033[32mExpected (Constraint kind): %s\033[0m\n"+
		"\033[31mActual (Constraint kind):   %s\033[0m\n\n",
		compiler.ConstraintKindMustBeFollowedBy.String(), got.Rules[1].Constraint.Kind.String())
}
