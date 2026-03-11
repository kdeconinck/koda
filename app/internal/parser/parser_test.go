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

// Verify the public API of the parser package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package parser_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/ast"
	"github.com/kdeconinck/koda/internal/loc"
	"github.com/kdeconinck/koda/internal/parser"
	"github.com/kdeconinck/koda/internal/scanner"
	"github.com/kdeconinck/koda/internal/token"
)

// Returns a source span with the provided start and end positions.
// This helper keeps test setup small and readable.
func span(startLine, startColumn, endLine, endColumn int) loc.Span {
	return loc.Span{
		Start: loc.Position{
			Line:   startLine,
			Column: startColumn,
		},
		End: loc.Position{
			Line:   endLine,
			Column: endColumn,
		},
	}
}

// Returns scanned tokens for the provided source text.
// Scanner setup is kept inside tests so the parser is still exercised through its public API.
func scanSource(t *testing.T, src string) []token.Token {
	t.Helper()

	got, err := scanner.Scan(src)

	assert.Nilf(t, err, "\n\n"+
		"UT Name:                  %s\n"+
		"\033[32mExpected (Scanner error): <nil>\033[0m\n"+
		"\033[31mActual (Scanner error):   %v\033[0m\n\n",
		"Scanner setup for parser test.", err)

	return got
}

// Verifies that [scanner.Scan] recognizes string literals and preserves their raw escape content in the produced
// lexeme.
func Test_Scan_RecognizesStringLiterals(t *testing.T) {
	t.Parallel()

	// Helper function:
	indent := func(s string, indentSize int) string {
		prefix := strings.Repeat(" ", indentSize)

		return strings.ReplaceAll(s, "\n", "\n"+prefix)
	}

	for snippet, wantErr := range map[string]string{
		"":                                        "Expected keyword `LANG`, but reached the end of input.",
		"RULES { }":                               "Expected keyword `LANG`.",
		"LANG":                                    "Expected literal, but reached the end of input.",
		"LANG \"java\"":                           "Expected keyword `EXTENSIONS`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS":                "Expected symbol `[`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS {":              "Expected symbol `[`.",
		"LANG \"java\" EXTENSIONS [":              "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS []":             "Expected symbol `{`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [}":             "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\",":    "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\", ]":  "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\"":     "Expected symbol `]`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" }":   "Expected symbol `]`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ]":   "Expected symbol `{`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {": "Expected keyword `SECTION`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET":                                                                                                                                            "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET RULES":                                                                                                                                      "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\"":                                                                                                                                "Expected keyword `VALUES`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" TOKENS":                                                                                                                         "Expected keyword `VALUES`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES":                                                                                                                         "Expected symbol `[`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES {":                                                                                                                       "Expected symbol `[`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [":                                                                                                                       "Expected character literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ \"1\"":                                                                                                                 "Expected character literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '\\z'":                                                                                                                 "Unsupported character escape \"\\\\z\"",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1'":                                                                                                                   "Expected symbol `]`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '0', ":                                                                                                                 "Expected character literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '0', \"1\"":                                                                                                            "Expected character literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '0', '1'":                                                                                                              "Expected symbol `]`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '0', '1' }":                                                                                                            "Expected symbol `]`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' }":                                                                                                                 "Expected symbol `]`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' .. ":                                                                                                               "Expected character literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' .. \"9\"":                                                                                                          "Expected character literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' .. '\\z'":                                                                                                          "Unsupported character escape \"\\\\z\"",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' .. '9'":                                                                                                            "Expected symbol `]`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  DEFINE CHARSET \"numbers\" VALUES [ '1' .. '9' }":                                                                                                          "Expected symbol `]`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\nTOKENS":                                                                                                                                                      "Expected keyword `SECTION`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION":                                                                                                                                                   "Expected keyword `TOKENS`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION RULES":                                                                                                                                             "Expected keyword `TOKENS`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS":                                                                                                                                            "Expected symbol `{`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS [":                                                                                                                                          "Expected symbol `{`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {":                                                                                                                                          "Expected keyword `DEFINE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    RULES":                                                                                                                               "Expected keyword `DEFINE`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE":                                                                                                                              "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE RULES":                                                                                                                        "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\"":                                                                                                                      "Expected one of `LITERAL`, `SEQUENCE` or `ENCLOSED_BY`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" RULES":                                                                                                                "Expected one of `LITERAL`, `SEQUENCE` or `ENCLOSED_BY`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" UNDEFINED":                                                                                                            "Expected one of `LITERAL`, `SEQUENCE` or `ENCLOSED_BY`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" LITERAL":                                                                                                              "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" LITERAL RULES":                                                                                                        "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" LITERAL \"key\"":                                                                                                      "Expected keyword `DEFINE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" SEQUENCE":                                                                                                             "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" SEQUENCE RULES":                                                                                                       "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" SEQUENCE \"numbers\"":                                                                                                 "Expected keyword `DEFINE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY":                                                                                                          "Expected character literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY RULES":                                                                                                    "Expected character literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY '@'":                                                                                                      "Expected character literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY '\\z' '@'":                                                                                                "Unsupported character escape \"\\\\z\"",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY '@' '@'":                                                                                                  "Expected keyword `DEFINE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS {\n    DEFINE \"key\" ENCLOSED_BY '@' '\\z'":                                                                                                "Unsupported character escape \"\\\\z\"",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }":                                                                                                                                        "Expected keyword `SECTION`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  RULES":                                                                                                                               "Expected keyword `SECTION`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION":                                                                                                                             "Expected keyword `RULES`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION TOKENS":                                                                                                                      "Expected keyword `RULES`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES":                                                                                                                       "Expected symbol `{`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES [":                                                                                                                     "Expected symbol `{`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {":                                                                                                                     "Expected keyword `RULE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    DEFINE":                                                                                                         "Expected keyword `RULE`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE":                                                                                                           "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE DEFINE":                                                                                                    "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\"":                                                                                                  "Expected symbol `{`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" [":                                                                                                "Expected symbol `{`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {":                                                                                                "Expected keyword `MATCH`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      DEFINE":                                                                                  "Expected keyword `MATCH`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH":                                                                                   "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH DEFINE":                                                                            "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"":                                                                           "Expected one of `MUST_BE_FOLLOWED_BY` or `CANNOT_BE_FOLLOWED_BY`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      DEFINE":                                                             "Expected one of `MUST_BE_FOLLOWED_BY` or `CANNOT_BE_FOLLOWED_BY`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY":                                                "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY DEFINE":                                         "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      CANNOT_BE_FOLLOWED_BY":                                              "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      CANNOT_BE_FOLLOWED_BY DEFINE":                                       "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"":                                      "Expected keyword `ERROR`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR":                         "Expected literal, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR DEFINE":                  "Expected literal.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR \"message\"":             "Expected symbol `}`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR \"message\"\n    ]":      "Expected symbol `}`.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR \"message\"\n    }":      "Expected keyword `RULE`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES {\n    RULE \"rule\" {\n      MATCH \"key\"\n      MUST_BE_FOLLOWED_BY \"colon\"\n      ERROR \"message\"\n    }\n  }": "Expected symbol `}`, but reached the end of input.",
		"LANG \"java\" EXTENSIONS [ \"java\" ] {\n  SECTION TOKENS { }\n  SECTION RULES { }\n} RULE":                                                                                                           "Unexpected tokens after the end of the `LANG` block.",
	} {

		// Arrange.
		tokens := scanSource(t, snippet)

		// Act.
		_, err := parser.Parse(tokens)

		// Assert.
		assert.NotNilf(t, err, "\n\n"+
			"UT Name:                When the input is NOT correct, parsing fails.\n"+
			"Input:                  %s\n"+
			"\033[32mExpected (Parser error): NOT <nil>\033[0m\n"+
			"\033[31mActual (Parser error):   %v\033[0m\n\n",
			indent(snippet, 26), err)

		assert.Equalf(t, err.Error(), wantErr, "\n\n"+
			"UT Name:                  When the input is NOT correct, parsing fails.\n"+
			"Input:                    %s\n"+
			"\033[32mExpected (Parser error): %q\033[0m\n"+
			"\033[31mActual (Parser error):   %q\033[0m\n\n",
			indent(snippet, 26), wantErr, err.Error())
	}
}

// Verifies that [parser.Parse] accepts a single file extension in the `EXTENSIONS` list.
func Test_Parse_SingleExtension(t *testing.T) {
	t.Parallel()

	const input = `
LANG "java" EXTENSIONS ["java"] {
	SECTION TOKENS { }
	SECTION RULES { }
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When the input contains a single extension, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Extensions), 1, "\n\n"+
		"UT Name:                 When the input contains a single extension, parsing succeeds.\n"+
		"\033[32mExpected (# Extensions): %d\033[0m\n"+
		"\033[31mActual (# Extensions):   %d\033[0m\n\n",
		1, len(got.Extensions))

	assert.Equalf(t, got.Extensions[0], "java", "\n\n"+
		"UT Name:                When the input contains a single extension, parsing succeeds.\n"+
		"\033[32mExpected (Extension 0): %q\033[0m\n"+
		"\033[31mActual (Extension 0):   %q\033[0m\n\n",
		"java", got.Extensions[0])
}

// Verifies that [parser.Parse] accepts multiple file extensions in the `EXTENSIONS` list.
func Test_Parse_MultipleExtensions(t *testing.T) {
	t.Parallel()

	const input = `
LANG "java" EXTENSIONS ["java", "jav"] {
	SECTION TOKENS { }
	SECTION RULES { }
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When the input contains multiple extensions, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Extensions), 2, "\n\n"+
		"UT Name:                 When the input contains multiple extensions, parsing succeeds.\n"+
		"\033[32mExpected (# Extensions): %d\033[0m\n"+
		"\033[31mActual (# Extensions):   %d\033[0m\n\n",
		2, len(got.Extensions))

	assert.Equalf(t, got.Extensions[0], "java", "\n\n"+
		"UT Name:                When the input contains multiple extensions, parsing succeeds.\n"+
		"\033[32mExpected (Extension 0): %q\033[0m\n"+
		"\033[31mActual (Extension 0):   %q\033[0m\n\n",
		"java", got.Extensions[0])

	assert.Equalf(t, got.Extensions[1], "jav", "\n\n"+
		"UT Name:                When the input contains multiple extensions, parsing succeeds.\n"+
		"\033[32mExpected (Extension 1): %q\033[0m\n"+
		"\033[31mActual (Extension 1):   %q\033[0m\n\n",
		"jav", got.Extensions[1])
}

// Verifies that [parser.Parse] accepts the minimal valid Koda profile shape.
func Test_Parse_MinimalLanguageProfile(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS [ "json" ] {
	SECTION TOKENS { }
	SECTION RULES { }
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, got.Name, "json", "\n\n"+
		"UT Name:         When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (Name): %q\033[0m\n"+
		"\033[31mActual (Name):   %q\033[0m\n\n",
		"json", got.Name)

	assert.Equalf(t, len(got.Extensions), 1, "\n\n"+
		"UT Name:                 When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (# Extensions): %d\033[0m\n"+
		"\033[31mActual (# Extensions):   %d\033[0m\n\n",
		1, len(got.Extensions))

	assert.Equalf(t, len(got.CharSets), 0, "\n\n"+
		"UT Name:               When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (# Charsets): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		0, len(got.CharSets))

	assert.Equalf(t, len(got.Tokens.Definitions), 0, "\n\n"+
		"UT Name:                        When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (# Token definitions): %d\033[0m\n"+
		"\033[31mActual (# Token definitions):   %d\033[0m\n\n",
		0, len(got.Tokens.Definitions))

	assert.Equalf(t, len(got.Rules.Rules), 0, "\n\n"+
		"UT Name:            When the input is valid (minimal), parsing succeeds.\n"+
		"\033[32mExpected (# Rules): %d\033[0m\n"+
		"\033[31mActual (# Rules):   %d\033[0m\n\n",
		0, len(got.Rules.Rules))
}

// Verifies that [parser.Parse] accepts charset declarations with valid escape sequences.
func Test_Parse_EscapeSequencesInCharSets(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "escaped" VALUES [ '\\', '\'', '\"', '\n', '\r', '\t' ]

	SECTION TOKENS {}
	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.CharSets), 1, "\n\n"+
		"UT Name:               When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (# Charsets): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		1, len(got.CharSets))

	assert.Equalf(t, len(got.CharSets[0].Items), 6, "\n\n"+
		"UT Name:                        When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - # Items): %d\033[0m\n"+
		"\033[31mActual (Charset 0 - # Items):   %d\033[0m\n\n",
		6, len(got.CharSets[0].Items))

	assert.Equalf(t, got.CharSets[0].Name, "escaped", "\n\n"+
		"UT Name:                     When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 0 - Name):   %q\033[0m\n\n",
		"escaped", got.CharSets[0].Name)

	assert.Equalf(t, got.CharSets[0].Items[0].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[0].Value, rune('\\'), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Value):   %q\033[0m\n\n",
		rune('\\'), got.CharSets[0].Items[0].Value)

	assert.Equalf(t, got.CharSets[0].Items[1].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[1].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[1].Value, rune('\''), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Value):   %q\033[0m\n\n",
		rune('\''), got.CharSets[0].Items[1].Value)

	assert.Equalf(t, got.CharSets[0].Items[2].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 2 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 2 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[2].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[2].Value, rune('"'), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 2 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 2 - Value):   %q\033[0m\n\n",
		rune('"'), got.CharSets[0].Items[2].Value)

	assert.Equalf(t, got.CharSets[0].Items[3].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 3 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 3 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[3].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[3].Value, rune('\n'), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 3 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 3 - Value):   %q\033[0m\n\n",
		rune('\n'), got.CharSets[0].Items[3].Value)

	assert.Equalf(t, got.CharSets[0].Items[4].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 4 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 4 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[4].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[4].Value, rune('\r'), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 4 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 4 - Value):   %q\033[0m\n\n",
		rune('\r'), got.CharSets[0].Items[4].Value)

	assert.Equalf(t, got.CharSets[0].Items[5].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 5 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 5 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[5].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[5].Value, rune('\t'), "\n\n"+
		"UT Name:                              When parsing a Charset with a valid escape sequence, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 5 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 5 - Value):   %q\033[0m\n\n",
		rune('\t'), got.CharSets[0].Items[5].Value)
}

// Verifies that [parser.Parse] accepts charset declarations with single-value items.
func Test_Parse_SingleValueCharSets(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "binary" VALUES [ '0', '1' ]

	SECTION TOKENS {}
	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.CharSets), 1, "\n\n"+
		"UT Name:               When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (# Charsets): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		1, len(got.CharSets))

	assert.Equalf(t, len(got.CharSets[0].Items), 2, "\n\n"+
		"UT Name:                        When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - # Items): %d\033[0m\n"+
		"\033[31mActual (Charset 0 - # Items):   %d\033[0m\n\n",
		2, len(got.CharSets[0].Items))

	assert.Equalf(t, got.CharSets[0].Name, "binary", "\n\n"+
		"UT Name:                     When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 0 - Name):   %q\033[0m\n\n",
		"binary", got.CharSets[0].Name)

	assert.Equalf(t, got.CharSets[0].Items[0].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[0].Value, rune('0'), "\n\n"+
		"UT Name:                              When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Value):   %q\033[0m\n\n",
		rune('0'), got.CharSets[0].Items[0].Value)

	assert.Equalf(t, got.CharSets[0].Items[1].Value, rune('1'), "\n\n"+
		"UT Name:                            When parsing a Charset with a single-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Value):   %q\033[0m\n\n",
		rune('1'), got.CharSets[0].Items[1].Value)
}

// Verifies that [parser.Parse] accepts charset declarations with range-value items.
func Test_Parse_RangeValueCharSets(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "digits" VALUES [ '0' .. '9' ]

	SECTION TOKENS {}
	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.CharSets), 1, "\n\n"+
		"UT Name:               When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (# Charset): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		1, len(got.CharSets))

	assert.Equalf(t, len(got.CharSets[0].Items), 1, "\n\n"+
		"UT Name:                        When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - # Items): %d\033[0m\n"+
		"\033[31mActual (Charset 0 - # Items):   %d\033[0m\n\n",
		1, len(got.CharSets[0].Items))

	assert.Equalf(t, got.CharSets[0].Name, "digits", "\n\n"+
		"UT Name:                     When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 0 - Name):   %q\033[0m\n\n",
		"digits", got.CharSets[0].Name)

	assert.Equalf(t, got.CharSets[0].Items[0].Kind, ast.CharSetItemKindRange, "\n\n"+
		"UT Name:                             When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindRange.String(),
		got.CharSets[0].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[0].Start, rune('0'), "\n\n"+
		"UT Name:                              When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Start): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Start):   %q\033[0m\n\n",
		rune('0'), got.CharSets[0].Items[0].Start)

	assert.Equalf(t, got.CharSets[0].Items[0].End, rune('9'), "\n\n"+
		"UT Name:                              When parsing a Charset with a range-value item, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - End): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - End):   %q\033[0m\n\n",
		rune('9'), got.CharSets[0].Items[0].End)
}

// Verifies that [parser.Parse] accepts charset declarations with both single-value and range-value items.
func Test_Parse_SingleAndRangeValueMixedCharSets(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "alpha" VALUES [ 'a'..'z', 'A'..'Z', '_' ]

	SECTION TOKENS {}
	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.CharSets), 1, "\n\n"+
		"UT Name:               When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (# Charsets): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		1, len(got.CharSets))

	assert.Equalf(t, len(got.CharSets[0].Items), 3, "\n\n"+
		"UT Name:                        When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - # Items): %d\033[0m\n"+
		"\033[31mActual (Charset 0 - # Items):   %d\033[0m\n\n",
		3, len(got.CharSets[0].Items))

	assert.Equalf(t, got.CharSets[0].Name, "alpha", "\n\n"+
		"UT Name:                     When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 0 - Name):   %q\033[0m\n\n",
		"alpha", got.CharSets[0].Name)

	assert.Equalf(t, got.CharSets[0].Items[0].Kind, ast.CharSetItemKindRange, "\n\n"+
		"UT Name:                             When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindRange.String(),
		got.CharSets[0].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[0].Start, rune('a'), "\n\n"+
		"UT Name:                              When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Start): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Start):   %q\033[0m\n\n",
		rune('a'), got.CharSets[0].Items[0].Start)

	assert.Equalf(t, got.CharSets[0].Items[0].End, rune('z'), "\n\n"+
		"UT Name:                              When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - End): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - End):   %q\033[0m\n\n",
		rune('z'), got.CharSets[0].Items[0].End)

	assert.Equalf(t, got.CharSets[0].Items[1].Kind, ast.CharSetItemKindRange, "\n\n"+
		"UT Name:                             When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindRange.String(),
		got.CharSets[0].Items[1].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[1].Start, rune('A'), "\n\n"+
		"UT Name:                              When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Start): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Start):   %q\033[0m\n\n",
		rune('A'), got.CharSets[0].Items[1].Start)

	assert.Equalf(t, got.CharSets[0].Items[1].End, rune('Z'), "\n\n"+
		"UT Name:                              When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - End): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - End):   %q\033[0m\n\n",
		rune('Z'), got.CharSets[0].Items[1].End)

	assert.Equalf(t, got.CharSets[0].Items[2].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 2 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 2 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[2].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[2].Value, rune('_'), "\n\n"+
		"UT Name:                              When parsing a Charset with both single-value and range-value items, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 2 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 2 - Value):   %q\033[0m\n\n",
		rune('_'), got.CharSets[0].Items[1].Start)
}

// Verifies that [parser.Parse] accepts multiple charset declarations.
func Test_Parse_MultipleCharSets(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	DEFINE CHARSET "binary" VALUES [ '0', '1' ]
	DEFINE CHARSET "digits" VALUES [ '0' .. '9' ]

	SECTION TOKENS {}
	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.CharSets), 2, "\n\n"+
		"UT Name:               When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (# Charsets): %d\033[0m\n"+
		"\033[31mActual (# Charsets):   %d\033[0m\n\n",
		2, len(got.CharSets))

	assert.Equalf(t, len(got.CharSets[0].Items), 2, "\n\n"+
		"UT Name:                        When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - # Items): %d\033[0m\n"+
		"\033[31mActual (Charset 0 - # Items):   %d\033[0m\n\n",
		2, len(got.CharSets[0].Items))

	assert.Equalf(t, got.CharSets[0].Name, "binary", "\n\n"+
		"UT Name:                     When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 0 - Name):   %q\033[0m\n\n",
		"binary", got.CharSets[0].Name)

	assert.Equalf(t, got.CharSets[0].Items[0].Kind, ast.CharSetItemKindSingle, "\n\n"+
		"UT Name:                             When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindSingle.String(),
		got.CharSets[0].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[0].Items[0].Value, rune('0'), "\n\n"+
		"UT Name:                              When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Value):   %q\033[0m\n\n",
		rune('0'), got.CharSets[0].Items[0].Value)

	assert.Equalf(t, got.CharSets[0].Items[1].Value, rune('1'), "\n\n"+
		"UT Name:                              When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 1 - Value): %q\033[0m\n"+
		"\033[31mActual (Charset 0, Item 1 - Value):   %q\033[0m\n\n",
		rune('1'), got.CharSets[0].Items[1].Value)

	assert.Equalf(t, got.CharSets[1].Name, "digits", "\n\n"+
		"UT Name:                     When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 1 - Name): %q\033[0m\n"+
		"\033[31mActual (Charset 1 - Name):   %q\033[0m\n\n",
		"digits", got.CharSets[1].Name)

	assert.Equalf(t, got.CharSets[1].Items[0].Kind, ast.CharSetItemKindRange, "\n\n"+
		"UT Name:                             When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 0, Item 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Charset 0, Item 0 - Kind):   %s\033[0m\n\n",
		ast.CharSetItemKindRange.String(),
		got.CharSets[1].Items[0].Kind.String())

	assert.Equalf(t, got.CharSets[1].Items[0].Start, rune('0'), "\n\n"+
		"UT Name:                              When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 1, Item 0 - Start): %d\033[0m\n"+
		"\033[31mActual (Charset 1, Item 0 - Start):   %d\033[0m\n\n",
		rune('0'), got.CharSets[1].Items[0].Start)

	assert.Equalf(t, got.CharSets[1].Items[0].End, rune('9'), "\n\n"+
		"UT Name:                              When parsing multiple Charsets, parsing succeeds.\n"+
		"\033[32mExpected (Charset 1, Item 0 - End): %q\033[0m\n"+
		"\033[31mActual (Charset 1, Item 0 - End):   %q\033[0m\n\n",
		rune('9'), got.CharSets[1].Items[0].End)
}

// Verifies that [parser.Parse] accepts a literal token definition.
func Test_Parse_LiteralTokenDefinition(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "brace_open" LITERAL "{"
	}

	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a literal token definition, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Tokens.Definitions), 1, "\n\n"+
		"UT Name:                        When parsing a literal token definition, parsing succeeds.\n"+
		"\033[32mExpected (# Token definitions): %d\033[0m\n"+
		"\033[31mActual (# Token definitions):   %d\033[0m\n\n",
		1, len(got.Tokens.Definitions))

	assert.Equalf(t, got.Tokens.Definitions[0].Kind, ast.TokenDefinitionKindLiteral, "\n\n"+
		"UT Name:                              When parsing a literal token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindLiteral.String(), got.Tokens.Definitions[0].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[0].Value, "{", "\n\n"+
		"UT Name:                               When parsing a literal token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Value): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Value):   %s\033[0m\n\n",
		"{", got.Tokens.Definitions[0].Value)
}

// Verifies that [parser.Parse] accepts a sequence token definition.
func Test_Parse_SequenceTokenDefinition(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "brace_open" SEQUENCE "digits"
	}

	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a sequence token definition, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Tokens.Definitions), 1, "\n\n"+
		"UT Name:                        When parsing a sequence token definition, parsing succeeds.\n"+
		"\033[32mExpected (# Token definitions): %d\033[0m\n"+
		"\033[31mActual (# Token definitions):   %d\033[0m\n\n",
		1, len(got.Tokens.Definitions))

	assert.Equalf(t, got.Tokens.Definitions[0].Kind, ast.TokenDefinitionKindSequence, "\n\n"+
		"UT Name:                              When parsing a sequence token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindSequence.String(), got.Tokens.Definitions[0].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[0].CharSetName, "digits", "\n\n"+
		"UT Name:                               When parsing a sequence token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Charset name): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Charset name):   %s\033[0m\n\n",
		"{", got.Tokens.Definitions[0].CharSetName)
}

// Verifies that [parser.Parse] accepts an "Enclosed By" token definition.
func Test_Parse_EnclosedByTokenDefinition(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "brace_open" ENCLOSED_BY '"' '|'
	}

	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing an \"Enclosed By\" token definition, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Tokens.Definitions), 1, "\n\n"+
		"UT Name:                        When parsing an \"Enclosed By\" token definition, parsing succeeds.\n"+
		"\033[32mExpected (# Token definitions): %d\033[0m\n"+
		"\033[31mActual (# Token definitions):   %d\033[0m\n\n",
		1, len(got.Tokens.Definitions))

	assert.Equalf(t, got.Tokens.Definitions[0].Kind, ast.TokenDefinitionKindEnclosedBy, "\n\n"+
		"UT Name:                              When parsing an \"Enclosed By\" token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindEnclosedBy.String(), got.Tokens.Definitions[0].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[0].Start, "\"", "\n\n"+
		"UT Name:                               When parsing an \"Enclosed By\" token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Start): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Start):   %s\033[0m\n\n",
		"\"", got.Tokens.Definitions[0].Start)

	assert.Equalf(t, got.Tokens.Definitions[0].End, "|", "\n\n"+
		"UT Name:                               When parsing an \"Enclosed By\" token definition, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - End): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - End):   %s\033[0m\n\n",
		"|", got.Tokens.Definitions[0].End)
}

// Verifies that [parser.Parse] accepts multiple token definitions.
func Test_Parse_MultipleTokenDefinitions(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {
		DEFINE "brace_open" LITERAL "{"
		DEFINE "number" SEQUENCE "digits"
		DEFINE "key" ENCLOSED_BY '"' '|'
	}

	SECTION RULES {}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Tokens.Definitions), 3, "\n\n"+
		"UT Name:                        When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (# Token definitions): %d\033[0m\n"+
		"\033[31mActual (# Token definitions):   %d\033[0m\n\n",
		3, len(got.Tokens.Definitions))

	assert.Equalf(t, got.Tokens.Definitions[0].Kind, ast.TokenDefinitionKindLiteral, "\n\n"+
		"UT Name:                              When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindLiteral.String(), got.Tokens.Definitions[0].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[0].Value, "{", "\n\n"+
		"UT Name:                               When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 0 - Value): %s\033[0m\n"+
		"\033[31mActual (Token definition 0 - Value):   %s\033[0m\n\n",
		"{", got.Tokens.Definitions[0].Value)

	assert.Equalf(t, got.Tokens.Definitions[1].Kind, ast.TokenDefinitionKindSequence, "\n\n"+
		"UT Name:                              When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 1 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 1 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindSequence.String(), got.Tokens.Definitions[1].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[1].CharSetName, "digits", "\n\n"+
		"UT Name:                                      When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 1 - Charset name): %s\033[0m\n"+
		"\033[31mActual (Token definition 1 - Charset name):   %s\033[0m\n\n",
		"{", got.Tokens.Definitions[1].CharSetName)

	assert.Equalf(t, got.Tokens.Definitions[2].Kind, ast.TokenDefinitionKindEnclosedBy, "\n\n"+
		"UT Name:                              When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 2 - Kind): %s\033[0m\n"+
		"\033[31mActual (Token definition 2 - Kind):   %s\033[0m\n\n",
		ast.TokenDefinitionKindEnclosedBy.String(), got.Tokens.Definitions[2].Kind.String())

	assert.Equalf(t, got.Tokens.Definitions[2].Start, "\"", "\n\n"+
		"UT Name:                               When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 2 - Start): %s\033[0m\n"+
		"\033[31mActual (Token definition 2 - Start):   %s\033[0m\n\n",
		"\"", got.Tokens.Definitions[2].Start)

	assert.Equalf(t, got.Tokens.Definitions[2].End, "|", "\n\n"+
		"UT Name:                               When parsing multiple token definitions, parsing succeeds.\n"+
		"\033[32mExpected (Token definition 2 - End): %s\033[0m\n"+
		"\033[31mActual (Token definition 2 - End):   %s\033[0m\n\n",
		"|", got.Tokens.Definitions[2].End)
}

// Verifies that [parser.Parse] accepts a "Cannot be followed by" rule definition.
func Test_Parse_CannotBeFollowedByRuleDefinition(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {}

	SECTION RULES {
		RULE "no_trailing_comma" {
			MATCH "comma"
			CANNOT_BE_FOLLOWED_BY "brace_close"
			ERROR "Trailing commas break standard JSON."
		}
	}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a \"Cannot be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Rules.Rules), 1, "\n\n"+
		"UT Name:            When parsing a \"Cannot be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (# Rules): %d\033[0m\n"+
		"\033[31mActual (# Rules):   %d\033[0m\n\n",
		1, len(got.Rules.Rules))

	assert.Equalf(t, got.Rules.Rules[0].Name, "no_trailing_comma", "\n\n"+
		"UT Name:                  When parsing a \"Cannot be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Name):   %q\033[0m\n\n",
		"no_trailing_comma", got.Rules.Rules[0].Name)

	assert.Equalf(t, got.Rules.Rules[0].Constraint.Kind, ast.ConstraintKindCannotBeFollowedBy, "\n\n"+
		"UT Name:                             When parsing a \"Cannot be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Constraint kind): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Constraint kind):   %q\033[0m\n\n",
		ast.ConstraintKindCannotBeFollowedBy.String(), got.Rules.Rules[0].Constraint.Kind.String())

	assert.Equalf(t, got.Rules.Rules[0].ErrorMessage, "Trailing commas break standard JSON.", "\n\n"+
		"UT Name:                           When parsing a \"Cannot be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Error message): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Error message):   %q\033[0m\n\n",
		"Trailing commas break standard JSON.", got.Rules.Rules[0].ErrorMessage)
}

// Verifies that [parser.Parse] accepts a "Must be followed by" rule definition.
func Test_Parse_MustBeFollowedByRuleDefinition(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {}

	SECTION RULES {
		RULE "pair_format" {
			MATCH "key"
			MUST_BE_FOLLOWED_BY "colon"
			ERROR "JSON keys must be followed by a colon."
		}
	}
}
`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a \"Must be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Rules.Rules), 1, "\n\n"+
		"UT Name:            When parsing a \"Must be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (# Rules): %d\033[0m\n"+
		"\033[31mActual (# Rules):   %d\033[0m\n\n",
		1, len(got.Rules.Rules))

	assert.Equalf(t, got.Rules.Rules[0].Name, "pair_format", "\n\n"+
		"UT Name:                  When parsing a \"Must be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Name):   %q\033[0m\n\n",
		"pair_format", got.Rules.Rules[0].Name)

	assert.Equalf(t, got.Rules.Rules[0].Constraint.Kind, ast.ConstraintKindMustBeFollowedBy, "\n\n"+
		"UT Name:                             When parsing a \"Must be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Constraint kind): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Constraint kind):   %q\033[0m\n\n",
		ast.ConstraintKindMustBeFollowedBy.String(), got.Rules.Rules[0].Constraint.Kind.String())

	assert.Equalf(t, got.Rules.Rules[0].ErrorMessage, "JSON keys must be followed by a colon.", "\n\n"+
		"UT Name:                           When parsing a \"Must be followed by\" rule definition, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Error message): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Error message):   %q\033[0m\n\n",
		"JSON keys must be followed by a colon.", got.Rules.Rules[0].ErrorMessage)
}

// Verifies that [parser.Parse] accepts multiple rule definitions.
func Test_Parse_MultipleRuleDefinitions(t *testing.T) {
	t.Parallel()

	const input = `
LANG "json" EXTENSIONS ["json"] {
	SECTION TOKENS {}

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
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got.Rules.Rules), 2, "\n\n"+
		"UT Name:            When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (# Rules): %d\033[0m\n"+
		"\033[31mActual (# Rules):   %d\033[0m\n\n",
		2, len(got.Rules.Rules))

	assert.Equalf(t, got.Rules.Rules[0].Name, "no_trailing_comma", "\n\n"+
		"UT Name:                  When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Name):   %q\033[0m\n\n",
		"no_trailing_comma", got.Rules.Rules[0].Name)

	assert.Equalf(t, got.Rules.Rules[0].Constraint.Kind, ast.ConstraintKindCannotBeFollowedBy, "\n\n"+
		"UT Name:                             When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Constraint kind): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Constraint kind):   %q\033[0m\n\n",
		ast.ConstraintKindCannotBeFollowedBy.String(), got.Rules.Rules[0].Constraint.Kind.String())

	assert.Equalf(t, got.Rules.Rules[0].ErrorMessage, "Trailing commas break standard JSON.", "\n\n"+
		"UT Name:                           When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Error message): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Error message):   %q\033[0m\n\n",
		"Trailing commas break standard JSON.", got.Rules.Rules[0].ErrorMessage)

	assert.Equalf(t, got.Rules.Rules[1].Name, "pair_format", "\n\n"+
		"UT Name:                  When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Name): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Name):   %q\033[0m\n\n",
		"pair_format", got.Rules.Rules[1].Name)

	assert.Equalf(t, got.Rules.Rules[1].Constraint.Kind, ast.ConstraintKindMustBeFollowedBy, "\n\n"+
		"UT Name:                             When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[32mExpected (Rule 0 - Constraint kind): %q\033[0m\n"+
		"\033[31mActual (Rule 0 - Constraint kind):   %q\033[0m\n\n",
		ast.ConstraintKindMustBeFollowedBy.String(), got.Rules.Rules[1].Constraint.Kind.String())

	assert.Equalf(t, got.Rules.Rules[1].ErrorMessage, "JSON keys must be followed by a colon.", "\n\n"+
		"UT Name:                           When parsing multiple rule definitions, parsing succeeds.\n"+
		"\033[31mActual (Rule 0 - Error message):   %q\033[0m\n\n",
		"JSON keys must be followed by a colon.", got.Rules.Rules[1].ErrorMessage)
}

// Verifies that [parser.Parse] tracks the span of the full `LANG` block.
func Test_Parse_TracksLanguageProfileSpan(t *testing.T) {
	t.Parallel()

	const input = `LANG "json" EXTENSIONS ["json"] { SECTION TOKENS { } SECTION RULES { } }`

	// Arrange.
	tokens := scanSource(t, input)

	// Act.
	got, err := parser.Parse(tokens)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:                 When parsing a `LANGUAGE` bloack, the span is correct.\n"+
		"\033[32mExpected (Parser error): <nil>\033[0m\n"+
		"\033[31mActual (Parser error):   %v\033[0m\n\n",
		err)

	want := span(1, 1, 1, 73)

	assert.Equalf(t, got.Span, want, "\n\n"+
		"UT Name:         When parsing a `LANGUAGE` bloack, the span is correct.\n"+
		"\033[32mExpected (Span): %+v\033[0m\n"+
		"\033[31mActual (Span):   %+v\033[0m\n\n",
		want, got.Span)
}
