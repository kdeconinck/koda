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

// Verify the public API of the scanner package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package scanner_test

import (
	"testing"

	"github.com/kdeconinck/koda/internal/assert"
	"github.com/kdeconinck/koda/internal/loc"
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

// Verifies that [scanner.Scan] recognizes all reserved keywords as keyword tokens.
func Test_Scan_RecognizesReservedKeywords(t *testing.T) {
	t.Parallel()

	for keyword := range map[string]struct{}{
		"LANG":                  {},
		"EXTENSIONS":            {},
		"DEFINE":                {},
		"CHARSET":               {},
		"VALUES":                {},
		"SECTION":               {},
		"TOKENS":                {},
		"RULES":                 {},
		"RULE":                  {},
		"MATCH":                 {},
		"ERROR":                 {},
		"MUST_BE_FOLLOWED_BY":   {},
		"CANNOT_BE_FOLLOWED_BY": {},
		"LITERAL":               {},
		"SEQUENCE":              {},
		"ENCLOSED_BY":           {},
	} {
		// Act.
		got, err := scanner.Scan(keyword)

		// Assert.
		assert.Nilf(t, err, "\n\n"+
			"UT Name:          When scanning the keyword '%s', the keyword is recognized.\n"+
			"\033[32mExpected (Error): <nil>\033[0m\n"+
			"\033[31mActual (Error):   %v\033[0m\n\n",
			keyword, err)

		assert.Equalf(t, len(got), 1, "\n\n"+
			"UT Name:          When scanning the keyword '%s', the keyword is recognized.\n"+
			"\033[32mExpected (# tokens): %d\033[0m\n"+
			"\033[31mActual (# tokens):   %d\033[0m\n\n",
			keyword, 1, len(got))

		assert.Equalf(t, got[0].Kind, token.KindKeyword, "\n\n"+
			"UT Name:          When scanning the keyword '%s', the keyword is recognized.\n"+
			"\033[32mExpected (Kind): %s\033[0m\n"+
			"\033[31mActual (Kind):   %s\033[0m\n\n",
			keyword, token.KindKeyword.String(), got[0].Kind.String())

		assert.Equalf(t, got[0].Lexeme, keyword, "\n\n"+
			"UT Name:          When scanning the keyword '%s', the keyword is recognized.\n"+
			"\033[32mExpected (Lexeme): %s\033[0m\n"+
			"\033[31mActual (Lexeme):   %s\033[0m\n\n",
			keyword, keyword, got[0].Lexeme)
	}
}

// Verifies that [scanner.Scan] recognizes identifier tokens that are not reserved keywords.
func Test_Scan_RecognizesIdentifiers(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input string
		want  string
	}{
		"When the source contains an identifier, it is recognized as an identifier.": {
			input: "abc",
			want:  "abc",
		},
		"When the source contains a mixed identifier with digits and underscores, it is recognized as an identifier.": {
			input: "alpha_2",
			want:  "alpha_2",
		},
		"When the source contains an underscore-only start, it is recognized as an identifier.": {
			input: "_hidden",
			want:  "_hidden",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Nilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): <nil>\033[0m\n"+
				"\033[31mActual (Error):   %v\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, len(got), 1, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 1, len(got))

			assert.Equalf(t, got[0].Kind, token.KindIdentifier, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Kind): %s\033[0m\n"+
				"\033[31mActual (Kind):   %s\033[0m\n\n",
				tcName, token.KindIdentifier.String(), got[0].Kind.String())

			assert.Equalf(t, got[0].Lexeme, tc.want, "\n\n"+
				"UT Name:           %s\n"+
				"\033[32mExpected (Lexeme): %s\033[0m\n"+
				"\033[31mActual (Lexeme):   %s\033[0m\n\n",
				tcName, tc.want, got[0].Lexeme)
		})
	}
}

// Verifies that [scanner.Scan] recognizes string literals and preserves their raw escape content in the produced
// lexeme.
func Test_Scan_RecognizesStringLiterals(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input string
		want  string
	}{
		"When the source contains a simple string literal, the inner text is returned as the lexeme.": {
			input: `"json"`,
			want:  "json",
		},
		"When the source contains an empty string literal, an empty lexeme is returned.": {
			input: `""`,
			want:  "",
		},
		"When the source contains an escaped quote inside a string literal, the raw escape is preserved.": {
			input: `"\""`,
			want:  `\"`,
		},
		"When the source contains an escaped newline marker inside a string literal, the raw escape is preserved.": {
			input: `"line\nbreak"`,
			want:  `line\nbreak`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Nilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): <nil>\033[0m\n"+
				"\033[31mActual (Error):   %v\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, len(got), 1, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 1, len(got))

			assert.Equalf(t, got[0].Kind, token.KindLiteral, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Kind): %s\033[0m\n"+
				"\033[31mActual (Kind):   %s\033[0m\n\n",
				tcName, token.KindLiteral.String(), got[0].Kind.String())

			assert.Equalf(t, got[0].Lexeme, tc.want, "\n\n"+
				"UT Name:           %s\n"+
				"\033[32mExpected (Lexeme): %s\033[0m\n"+
				"\033[31mActual (Lexeme):   %s\033[0m\n\n",
				tcName, tc.want, got[0].Lexeme)
		})
	}
}

// Verifies that [scanner.Scan] returns an error when a string literal is not terminated.
func Test_Scan_ReturnsErrorForUnterminatedStringLiteral(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input       string
		wantMessage string
		wantSpan    loc.Span
	}{
		"When the closing quote is missing, an unterminated string literal error is returned.": {
			input:       `"json`,
			wantMessage: "Unterminated string literal.",
			wantSpan:    span(1, 1, 1, 6),
		},
		"When the source ends immediately after a backslash in a string literal, an unterminated string literal error is returned.": {
			input:       `"abc\`,
			wantMessage: "Unterminated string literal.",
			wantSpan:    span(1, 1, 1, 6),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Equalf(t, len(got), 0, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 0, len(got))

			assert.NotNilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): NOT <nil>\033[0m\n"+
				"\033[31mActual (Error):   <nil>\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, err.Error(), tc.wantMessage, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): %q\033[0m\n"+
				"\033[31mActual (Error):   %q\033[0m\n\n",
				tcName, tc.wantMessage, err.Error())

			scanErr := err.(scanner.Error)

			assert.Equalf(t, scanErr.Span, tc.wantSpan, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Span): %+v\033[0m\n"+
				"\033[31mActual (Span):   %+v\033[0m\n\n",
				tcName, tc.wantSpan, scanErr.Span)
		})
	}
}

// Verifies that [scanner.Scan] recognizes character literals and preserves their raw escape content in the produced
// lexeme.
func Test_Scan_RecognizesCharacterLiterals(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input string
		want  string
	}{
		"When the source contains a simple character literal, the inner character is returned as the lexeme.": {
			input: `'0'`,
			want:  "0",
		},
		"When the source contains an underscore character literal, the inner character is returned as the lexeme.": {
			input: `'_'`,
			want:  "_",
		},
		"When the source contains an escaped single quote, the raw escape is preserved.": {
			input: `'\''`,
			want:  `\'`,
		},
		"When the source contains an escaped double quote, the raw escape is preserved.": {
			input: `'"'`,
			want:  `"`,
		},
		"When the source contains an escaped newline marker, the raw escape is preserved.": {
			input: `'\n'`,
			want:  `\n`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Nilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): <nil>\033[0m\n"+
				"\033[31mActual (Error):   %v\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, len(got), 1, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 1, len(got))

			assert.Equalf(t, got[0].Kind, token.KindCharacter, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Kind): %s\033[0m\n"+
				"\033[31mActual (Kind):   %s\033[0m\n\n",
				tcName, token.KindCharacter.String(), got[0].Kind.String())

			assert.Equalf(t, got[0].Lexeme, tc.want, "\n\n"+
				"UT Name:           %s\n"+
				"\033[32mExpected (Lexeme): %s\033[0m\n"+
				"\033[31mActual (Lexeme):   %s\033[0m\n\n",
				tcName, tc.want, got[0].Lexeme)
		})
	}
}

// Verifies that [scanner.Scan] returns an error when a character literal is invalid.
func Test_Scan_ReturnsErrorForInvalidCharacterLiteral(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input       string
		wantMessage string
		wantSpan    loc.Span
	}{
		"When the character literal is empty, an explicit empty-character error is returned.": {
			input:       `''`,
			wantMessage: "Character literal cannot be empty.",
			wantSpan:    span(1, 1, 1, 2),
		},
		"When the character literal end immediately after the opening quote, an explicit empty-character error is returned.": {
			input:       `'`,
			wantMessage: "Unterminated character literal.",
			wantSpan:    span(1, 1, 1, 2),
		},
		"When the closing quote is missing, an unterminated character literal error is returned.": {
			input:       `'a`,
			wantMessage: "Unterminated character literal.",
			wantSpan:    span(1, 1, 1, 3),
		},
		"When the source ends immediately after a backslash in a character literal, an unterminated character literal error is returned.": {
			input:       `'\`,
			wantMessage: "Unterminated character literal.",
			wantSpan:    span(1, 1, 1, 3),
		},
		"When the closing quote is missing after an escaped character, an unterminated character literal error is returned.": {
			input:       `'\n`,
			wantMessage: "Unterminated character literal.",
			wantSpan:    span(1, 1, 1, 4),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Equalf(t, len(got), 0, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 0, len(got))

			assert.NotNilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): NOT <nil>\033[0m\n"+
				"\033[31mActual (Error):   <nil>\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, err.Error(), tc.wantMessage, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): %q\033[0m\n"+
				"\033[31mActual (Error):   %q\033[0m\n\n",
				tcName, tc.wantMessage, err.Error())

			scanErr := err.(scanner.Error)

			assert.Equalf(t, scanErr.Span, tc.wantSpan, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Span): %+v\033[0m\n"+
				"\033[31mActual (Span):   %+v\033[0m\n\n",
				tcName, tc.wantSpan, scanErr.Span)
		})
	}
}

// Verifies that [scanner.Scan] recognizes all single-character symbols and the range operator.
func Test_Scan_RecognizesSymbols(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input string
		want  string
	}{
		"When the source contains an opening brace, it is recognized as a symbol.": {
			input: "{",
			want:  "{",
		},
		"When the source contains a closing brace, it is recognized as a symbol.": {
			input: "}",
			want:  "}",
		},
		"When the source contains an opening bracket, it is recognized as a symbol.": {
			input: "[",
			want:  "[",
		},
		"When the source contains a closing bracket, it is recognized as a symbol.": {
			input: "]",
			want:  "]",
		},
		"When the source contains a comma, it is recognized as a symbol.": {
			input: ",",
			want:  ",",
		},
		"When the source contains the range operator, it is recognized as a symbol.": {
			input: "..",
			want:  "..",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Nilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): <nil>\033[0m\n"+
				"\033[31mActual (Error):   %v\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, len(got), 1, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 1, len(got))

			assert.Equalf(t, got[0].Kind, token.KindSymbol, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Kind): %s\033[0m\n"+
				"\033[31mActual (Kind):   %s\033[0m\n\n",
				tcName, token.KindSymbol.String(), got[0].Kind.String())

			assert.Equalf(t, got[0].Lexeme, tc.want, "\n\n"+
				"UT Name:           %s\n"+
				"\033[32mExpected (Lexeme): %s\033[0m\n"+
				"\033[31mActual (Lexeme):   %s\033[0m\n\n",
				tcName, tc.want, got[0].Lexeme)
		})
	}
}

// Verifies that [scanner.Scan] returns an error when a single dot does not form the range operator.
func Test_Scan_ReturnsErrorForSingleDot(t *testing.T) {
	t.Parallel()

	// Act.
	got, err := scanner.Scan(".")

	// Assert.
	assert.Equalf(t, len(got), 0, "\n\n"+
		"UT Name:             Scan fails for a single dot.\n"+
		"\033[32mExpected (# tokens): %d\033[0m\n"+
		"\033[31mActual (# tokens):   %d\033[0m\n\n",
		0, len(got))

	assert.NotNilf(t, err, "\n\n"+
		"UT Name:             Scan fails for a single dot.\n"+
		"\033[32mExpected (Error): NOT <nil>\033[0m\n"+
		"\033[31mActual (Error):   <nil>\033[0m\n\n",
		err)

	assert.Equalf(t, err.Error(), "Unexpected `.`. Expected the range operator `..`.", "\n\n"+
		"UT Name:             Scan fails for a single dot.\n"+
		"\033[32mExpected (Error): %q\033[0m\n"+
		"\033[31mActual (Error):   %q\033[0m\n\n",
		"Unexpected `.`. Expected the range operator `..`.", err.Error())

	scanErr := err.(scanner.Error)
	wantSpan := span(1, 1, 1, 2)

	assert.Equalf(t, scanErr.Span, wantSpan, "\n\n"+
		"UT Name:             Scan fails for a single dot.\n"+
		"\033[32mExpected (Span): %+v\033[0m\n"+
		"\033[31mActual (Span):   %+v\033[0m\n\n",
		wantSpan, scanErr.Span)
}

// Verifies that [scanner.Scan] skips insignificant whitespace and line comments.
func Test_Scan_SkipsWhitespaceAndComments(t *testing.T) {
	t.Parallel()

	const input = `
# language profile
LANG "json" EXTENSIONS ["json"] {
	# token section
}
`

	// Act.
	got, err := scanner.Scan(input)

	// Assert.
	assert.Nilf(t, err, "\n\n"+
		"UT Name:          When scanning, whitespace and comments are skipped.\n"+
		"\033[32mExpected (Error): <nil>\033[0m\n"+
		"\033[31mActual (Error):   %v\033[0m\n\n",
		err)

	assert.Equalf(t, len(got), 8, "\n\n"+
		"UT Name:             When scanning, whitespace and comments are skipped.\n"+
		"\033[32mExpected (# tokens): %d\033[0m\n"+
		"\033[31mActual (# tokens):   %d\033[0m\n\n",
		8, len(got))

	wantLexemes := []string{"LANG", "json", "EXTENSIONS", "[", "json", "]", "{", "}"}

	for idx := range wantLexemes {
		assert.Equalf(t, got[idx].Lexeme, wantLexemes[idx], "\n\n"+
			"UT Name:             When scanning, whitespace and comments are skipped.\n"+
			"\033[32mExpected (Lexeme %d): %s\033[0m\n"+
			"\033[31mActual (Lexeme %d):   %s\033[0m\n\n",
			idx, wantLexemes[idx], idx, got[idx].Lexeme)
	}
}

// Verifies that [scanner.Scan] tracks accurate spans for tokens, including tokens that appear after newlines and
// comments.
func Test_Scan_TracksTokenSpans(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input    string
		wantSpan loc.Span
	}{
		"When the token is at the start of the file, its span is tracked correctly.": {
			input:    "LANG",
			wantSpan: span(1, 1, 1, 5),
		},
		"When the token appears after a comment and newline, its span is tracked correctly.": {
			input:    "# comment\nLANG",
			wantSpan: span(2, 1, 2, 5),
		},
		"When a string literal is scanned, its span includes both surrounding quotes.": {
			input:    `"json"`,
			wantSpan: span(1, 1, 1, 7),
		},
		"When a character literal is scanned, its span includes both surrounding quotes.": {
			input:    `'0'`,
			wantSpan: span(1, 1, 1, 4),
		},
		"When the range operator is scanned, its span covers both dots.": {
			input:    "..",
			wantSpan: span(1, 1, 1, 3),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Nilf(t, err, "\n\n"+
				"UT Name:  %s\n"+
				"\033[32mExpected: <nil>\033[0m\n"+
				"\033[31mActual:   %v\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, got[0].Span, tc.wantSpan, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Span): %+v\033[0m\n"+
				"\033[31mActual (Span):   %+v\033[0m\n\n",
				tcName, tc.wantSpan, got[0].Span)
		})
	}
}

// Verifies that [scanner.Scan] returns an error when an unexpected rune cannot begin any valid token.
func Test_Scan_ReturnsErrorForUnexpectedCharacter(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		input       string
		wantMessage string
		wantSpan    loc.Span
	}{
		"When the source contains an unsupported symbol, an unexpected-character error is returned.": {
			input:       "@",
			wantMessage: "Unexpected character `@`.",
			wantSpan:    span(1, 1, 1, 2),
		},
		"When the source begins with a digit, an unexpected-character error is returned.": {
			input:       "1abc",
			wantMessage: "Unexpected character `1`.",
			wantSpan:    span(1, 1, 1, 2),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got, err := scanner.Scan(tc.input)

			// Assert.
			assert.Equalf(t, len(got), 0, "\n\n"+
				"UT Name:             %s\n"+
				"\033[32mExpected (# tokens): %d\033[0m\n"+
				"\033[31mActual (# tokens):   %d\033[0m\n\n",
				tcName, 0, len(got))

			assert.NotNilf(t, err, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): NOT <nil>\033[0m\n"+
				"\033[31mActual (Error):   <nil>\033[0m\n\n",
				tcName, err)

			assert.Equalf(t, err.Error(), tc.wantMessage, "\n\n"+
				"UT Name:          %s\n"+
				"\033[32mExpected (Error): %q\033[0m\n"+
				"\033[31mActual (Error):   %q\033[0m\n\n",
				tcName, tc.wantMessage, err.Error())

			scanErr := err.(scanner.Error)

			assert.Equalf(t, scanErr.Span, tc.wantSpan, "\n\n"+
				"UT Name:         %s\n"+
				"\033[32mExpected (Span): %+v\033[0m\n"+
				"\033[31mActual (Span):   %+v\033[0m\n\n",
				tcName, tc.wantSpan, scanErr.Span)
		})
	}
}
