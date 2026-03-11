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

// Verify the public API of the error package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package validator_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/ast"
	"github.com/kdeconinck/koda/internal/parser"
	"github.com/kdeconinck/koda/internal/scanner"
	"github.com/kdeconinck/koda/internal/validator"
)

// Returns a parsed language profile for the provided Koda source text.
// Scanner and parser setup are kept inside tests so the validator is still exercised through its public API.
func parseSource(t *testing.T, src string) ast.LanguageProfile {
	t.Helper()

	tokens, scanErr := scanner.Scan(src)

	assert.Nilf(t, scanErr, "\n\n"+
		"UT Name:                  Parser setup for validator test.\n"+
		"\033[32mExpected (Scanner error): <nil>\033[0m\n"+
		"\033[31mActual (Scanner error):   %v\033[0m\n\n",
		scanErr)

	profile, parseErr := parser.Parse(tokens)

	assert.Nilf(t, parseErr, "\n\n"+
		"UT Name:                 Parser setup for validator test.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		parseErr)

	return profile
}

// Verifies that [validator.Validate] accepts a semantically valid language profile.
func Test_Validate_AcceptsValidLanguageProfile(t *testing.T) {
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
	profile := parseSource(t, input)

	// Act.
	err := validator.Validate(profile)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                    When validating a semantically valid language profile, validation passes.\n"+
		"\033[32mExpected (Validator error): <nil>\033[0m\n"+
		"\033[31mActual (Validator error):   %v\033[0m\n\n",
		err)
}

// Verifies that [validator.Validate] rejects invalid semantic forms in parsed language profiles.
func Test_Validate_RejectsInvalidLanguageProfileForms(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input     string
		wantError string
	}{
		"When a charset name is declared twice, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "digits" VALUES ['0'..'9']
	DEFINE CHARSET "digits" VALUES ['1'..'8']

	SECTION TOKENS {
	}

	SECTION RULES {
	}
}
`,
			wantError: "Duplicate charset name `digits`.",
		},
		"When a token definition name is declared twice, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "number" LITERAL "1"
		DEFINE "number" LITERAL "2"
	}

	SECTION RULES {
	}
}
`,
			wantError: "Duplicate token definition name `number`.",
		},
		"When a rule name is declared twice, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "comma" LITERAL ","
		DEFINE "brace_close" LITERAL "}"
	}

	SECTION RULES {
		RULE "no_trailing_comma" {
			MATCH "comma"
			CANNOT_BE_FOLLOWED_BY "brace_close"
			ERROR "Trailing commas break standard JSON."
		}

		RULE "no_trailing_comma" {
			MATCH "comma"
			CANNOT_BE_FOLLOWED_BY "brace_close"
			ERROR "Trailing commas still break standard JSON."
		}
	}
}
`,
			wantError: "Duplicate rule name `no_trailing_comma`.",
		},
		"When a charset range is descending, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "letters" VALUES ['z'..'a']

	SECTION TOKENS {
	}

	SECTION RULES {
	}
}
`,
			wantError: "Invalid charset range `z`..`a` in charset `letters`.",
		},
		"When a sequence token definition refers to an unknown charset, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "number" SEQUENCE "digits"
	}

	SECTION RULES {
	}
}
`,
			wantError: "Unknown charset reference `digits` in token definition `number`.",
		},
		"When a rule MATCH clause refers to an unknown token, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "colon" LITERAL ":"
	}

	SECTION RULES {
		RULE "pair_format" {
			MATCH "key"
			MUST_BE_FOLLOWED_BY "colon"
			ERROR "JSON keys must be followed by a colon."
		}
	}
}
`,
			wantError: "Unknown token reference `key` in MATCH clause of rule `pair_format`.",
		},
		"When a rule constraint refers to an unknown token, validation fails.": {
			input: `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "key" ENCLOSED_BY '"' '"'
	}

	SECTION RULES {
		RULE "pair_format" {
			MATCH "key"
			MUST_BE_FOLLOWED_BY "colon"
			ERROR "JSON keys must be followed by a colon."
		}
	}
}
`,
			wantError: "Unknown token reference `colon` in constraint of rule `pair_format`.",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			profile := parseSource(t, tc.input)

			// Act.
			err := validator.Validate(profile)

			// Assert.
			assert.NotNilf(t, err, "\n\n"+
				"UT Name:                    %s\n"+
				"\033[32mExpected (Validator error): NOT <nil>\033[0m\n"+
				"\033[31mActual (Validator error):   <nil>\033[0m\n\n",
				tcName)

			assert.Equalf(t, err.Error(), tc.wantError, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): %q\033[0m\n"+
				"\033[31mActual (Error):   %q\033[0m\n\n",
				tcName, tc.wantError, err.Error())
		})
	}
}
