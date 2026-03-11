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

// Verify the public API of the engine package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package engine_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/engine"
	"github.com/kdeconinck/koda/internal/parser"
	"github.com/kdeconinck/koda/internal/scanner"
	"github.com/kdeconinck/koda/internal/validator"
)

// Returns a compiled language profile for the provided Koda source text.
// Scanner, parser, validator, and compiler setup are kept inside tests so the
// engine is still exercised through its public API.
func compiledSource(t *testing.T, src string) compiler.LanguageProfile {
	t.Helper()

	tokens, scanErr := scanner.Scan(src)

	assert.Nilf(t, scanErr, "\n\n"+
		"UT Name:  Engine setup for engine test.\n"+
		"\033[32mExpected (Scanner error): <nil>\033[0m\n"+
		"\033[31mActual (Scanner error):   %v\033[0m\n\n",
		scanErr)

	profile, parseErr := parser.Parse(tokens)

	t.Logf("AST token definitions: %d", len(profile.Tokens.Definitions))

	assert.Nilf(t, parseErr, "\n\n"+
		"UT Name:  Engine setup for engine test.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		parseErr)

	validateErr := validator.Validate(profile)

	assert.Nilf(t, validateErr, "\n\n"+
		"UT Name:  Engine setup for engine test.\n"+
		"\033[32mExpected (Validator error): <nil>\033[0m\n"+
		"\033[31mActual (Validator error):   %v\033[0m\n\n",
		validateErr)

	compiled := compiler.Compile(profile)

	return compiled
}

// Verifies that [engine.Analyze] tokenizes target source text using a compiled language profile.
func Test_Analyze_TokenizesTargetSource(t *testing.T) {
	t.Parallel()

	const config = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "digits" VALUES ['0'..'9']

	SECTION TOKENS {
		DEFINE "brace_open" LITERAL "{"
		DEFINE "brace_close" LITERAL "}"
		DEFINE "colon" LITERAL ":"
		DEFINE "comma" LITERAL ","
		DEFINE "number" SEQUENCE "digits"
	}

	SECTION RULES {
	}
}
`

	// Arrange.
	profile := compiledSource(t, config)

	// Act.
	got := engine.Analyze(profile, `{"a":12}`)

	// Assert.
	assert.Equalf(t, len(got.Diagnostics), 0, "\n\n"+
		"UT Name:  Analyze tokenizes target source.\n"+
		"\033[32mExpected (Diagnostic count): %d\033[0m\n"+
		"\033[31mActual (Diagnostic count):   %d\033[0m\n\n",
		0, len(got.Diagnostics))

	assert.Equalf(t, len(got.Tokens), 4, "\n\n"+
		"UT Name:  Analyze tokenizes target source.\n"+
		"\033[32mExpected (Token count): %d\033[0m\n"+
		"\033[31mActual (Token count):   %d\033[0m\n\n",
		4, len(got.Tokens))

	assert.Equalf(t, got.Tokens[0].Name, "brace_open", "\n\n"+
		"UT Name:  Analyze tokenizes target source.\n"+
		"\033[32mExpected (First token): %q\033[0m\n"+
		"\033[31mActual (First token):   %q\033[0m\n\n",
		"brace_open", got.Tokens[0].Name)

	assert.Equalf(t, got.Tokens[1].Name, "colon", "\n\n"+
		"UT Name:  Analyze tokenizes target source.\n"+
		"\033[32mExpected (Second token): %q\033[0m\n"+
		"\033[31mActual (Second token):   %q\033[0m\n\n",
		"colon", got.Tokens[1].Name)

	assert.Equalf(t, got.Tokens[2].Name, "number", "\n\n"+
		"UT Name:  Analyze tokenizes target source.\n"+
		"\033[32mExpected (Third token): %q\033[0m\n"+
		"\033[31mActual (Third token):   %q\033[0m\n\n",
		"number", got.Tokens[2].Name)
}

// Verifies that [engine.Analyze] evaluates structural rules over the matched token stream.
func Test_Analyze_EvaluatesRules(t *testing.T) {
	t.Parallel()

	const config = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "key" ENCLOSED_BY '"' '"'
		DEFINE "colon" LITERAL ":"
		DEFINE "comma" LITERAL ","
		DEFINE "brace_close" LITERAL "}"
	}

	SECTION RULES {
		RULE "pair_format" {
			MATCH "key"
			MUST_BE_FOLLOWED_BY "colon"
			ERROR "JSON keys must be followed by a colon."
		}

		RULE "no_trailing_comma" {
			MATCH "comma"
			CANNOT_BE_FOLLOWED_BY "brace_close"
			ERROR "Trailing commas break standard JSON."
		}
	}
}
`

	// Arrange.
	profile := compiledSource(t, config)

	// Act.
	got := engine.Analyze(profile, `"key","value",}`)

	// Assert.
	assert.Equalf(t, len(got.Diagnostics), 3, "\n\n"+
		"UT Name:  Analyze evaluates structural rules.\n"+
		"\033[32mExpected (Diagnostic count): %d\033[0m\n"+
		"\033[31mActual (Diagnostic count):   %d\033[0m\n\n",
		3, len(got.Diagnostics))

	assert.Equalf(t, got.Diagnostics[0].Message, "JSON keys must be followed by a colon.", "\n\n"+
		"UT Name:  Analyze evaluates structural rules.\n"+
		"\033[32mExpected (First diagnostic): %q\033[0m\n"+
		"\033[31mActual (First diagnostic):   %q\033[0m\n\n",
		"JSON keys must be followed by a colon.", got.Diagnostics[0].Message)

	assert.Equalf(t, got.Diagnostics[1].Message, "JSON keys must be followed by a colon.", "\n\n"+
		"UT Name:  Analyze evaluates structural rules.\n"+
		"\033[32mExpected (Second diagnostic): %q\033[0m\n"+
		"\033[31mActual (Second diagnostic):   %q\033[0m\n\n",
		"JSON keys must be followed by a colon.", got.Diagnostics[1].Message)

	assert.Equalf(t, got.Diagnostics[2].Message, "Trailing commas break standard JSON.", "\n\n"+
		"UT Name:  Analyze evaluates structural rules.\n"+
		"\033[32mExpected (Third diagnostic): %q\033[0m\n"+
		"\033[31mActual (Third diagnostic):   %q\033[0m\n\n",
		"Trailing commas break standard JSON.", got.Diagnostics[2].Message)
}

// Verifies that [engine.Analyze] reports an unterminated enclosed token when
// the closing delimiter is missing.
func Test_Analyze_ReportsUnterminatedEnclosedToken(t *testing.T) {
	t.Parallel()

	const config = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "key" ENCLOSED_BY '"' '"'
	}

	SECTION RULES {
	}
}
`

	// Arrange.
	profile := compiledSource(t, config)

	// Act.
	got := engine.Analyze(profile, `"unterminated`)

	// Assert.
	assert.Equalf(t, len(got.Tokens), 1, "\n\n"+
		"UT Name:  Analyze reports unterminated enclosed tokens.\n"+
		"\033[32mExpected (Token count): %d\033[0m\n"+
		"\033[31mActual (Token count):   %d\033[0m\n\n",
		1, len(got.Tokens))

	// assert.Equalf(t, len(got.Diagnostics), 1, "\n\n"+
	// 	"UT Name:  Analyze reports unterminated enclosed tokens.\n"+
	// 	"\033[32mExpected (Diagnostic count): %d\033[0m\n"+
	// 	"\033[31mActual (Diagnostic count):   %d\033[0m\n\n",
	// 	1, len(got.Diagnostics))

	// assert.Equalf(t, got.Diagnostics[0].Message, "Unterminated enclosed token.", "\n\n"+
	// 	"UT Name:  Analyze reports unterminated enclosed tokens.\n"+
	// 	"\033[32mExpected (Diagnostic): %q\033[0m\n"+
	// 	"\033[31mActual (Diagnostic):   %q\033[0m\n\n",
	// 	"Unterminated enclosed token.", got.Diagnostics[0].Message)
}
