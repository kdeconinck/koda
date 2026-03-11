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

// Package scanner converts raw `.core` configuration source text into lexical tokens.
//
// The scanner performs lexical analysis only.
// It does not interpret the meaning of tokens or validate their structural relationships.
package scanner

import (
	"fmt"

	"github.com/kdeconinck/koda/internal/loc"
	"github.com/kdeconinck/koda/internal/token"
)

// Contains the reserved words of the Koda configuration language.
// When one of these words appears in identifier form, it is emitted as a keyword token rather than a generic identifier
// token.
var keywords = map[string]struct{}{
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
}

// Scan converts raw `.core` source text into a sequence of lexical tokens.
//
// Scan skips insignificant whitespace and line comments, recognizes the token forms used by the Koda configuration
// language, and records precise source spans for every token it produces.
//
// The scanner recognizes:
//
//   - Reserved keywords such as `LANG`, `DEFINE`, and `SECTION`.
//   - Double-quoted string literals.
//   - Single-quoted character literals.
//   - Identifiers such as `LITERAL`, `SEQUENCE`, and `ENCLOSED_BY`.
//   - Structural symbols such as `{`, `}`, `[`, `]`, and `,`.
//   - The range operator `..`.
//
// Scan returns an error when the source contains an unterminated literal, an invalid range operator, or an unexpected
// character.
func Scan(src string) ([]token.Token, error) {
	return newScanner(src).scanAll()
}

// A scanner holds the state required to tokenize a `.core` source file.
//
// Keeping the implementation stateful makes it easier to track the current rune position, produce accurate spans, and
// split the scanner into small helper methods.
type scanner struct {
	// The full input source, stored as runes for safe character-by-character scanning.
	src []rune

	// The index of the current rune in the input.
	index int

	// The current 1-based line number.
	line int

	// The current 1-based column number.
	column int
}

// Returns a new scanner for the provided source text.
// The scanner stores the input as runes and initializes its read position to the first character on line 1, column 1.
func newScanner(src string) *scanner {
	return &scanner{
		src:    []rune(src),
		index:  0,
		line:   1,
		column: 1,
	}
}

// Scans the full input source and returns the produced tokens.
// Scanning continues until the end of the input is reached or a lexical error is encountered.
func (s *scanner) scanAll() ([]token.Token, error) {
	tokens := make([]token.Token, 0)

	for {
		r, ok := s.peek()

		if !ok {
			return tokens, nil
		}

		switch {
		case isWhitespace(r):
			s.skipWhitespace()

		case r == '#':
			s.skipComment()

		case r == '"':
			tok, err := s.scanStringLiteral()

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, tok)

		case r == '\'':
			tok, err := s.scanCharacterLiteral()

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, tok)

		case isIdentifierStart(r):
			tokens = append(tokens, s.scanIdentifierOrKeyword())

		case isSingleSymbol(r):
			tokens = append(tokens, s.scanSingleSymbol())

		case r == '.':
			tok, err := s.scanRangeOperator()

			if err != nil {
				return nil, err
			}

			tokens = append(tokens, tok)

		default:
			return nil, s.unexpectedCharacterError(r)
		}
	}
}

// Returns the current rune without consuming it.
// The boolean result is false when the scanner has reached the end of the input.
func (s *scanner) peek() (rune, bool) {
	if s.index >= len(s.src) {
		return 0, false
	}

	return s.src[s.index], true
}

// Returns the rune located at the given lookahead offset without consuming it.
// This is primarily used for multi-character operators such as the rangeoperator `..`.
func (s *scanner) peekN(offset int) (rune, bool) {
	idx := s.index + offset

	if idx >= len(s.src) {
		return 0, false
	}

	return s.src[idx], true
}

// Consumes the current rune and advances the scanner by one position.
// Line and column tracking are updated automatically so later tokens and errors can report accurate source locations.
func (s *scanner) advance() (rune, bool) {
	r, ok := s.peek()

	if !ok {
		return 0, false
	}

	s.index++

	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	return r, true
}

// Returns the current scanner position as a source location.
// The returned value always reflects the position of the next rune to be read.
func (s *scanner) position() loc.Position {
	return loc.Position{
		Line:   s.line,
		Column: s.column,
	}
}

// Builds a source span from the provided start position to the current scanner position.
// This is typically used after a token has been fully consumed.
func (s *scanner) spanFrom(start loc.Position) loc.Span {
	return loc.Span{
		Start: start,
		End:   s.position(),
	}
}

// Consumes consecutive whitespace characters and discards them.
// Whitespace does not produce tokens in the Koda configuration language.
func (s *scanner) skipWhitespace() {
	for {
		r, ok := s.peek()

		if !ok || !isWhitespace(r) {
			return
		}

		s.advance()
	}
}

// Consumes a line comment that starts with `#`.
// The scanner stops just before the newline or at the end of the input.
func (s *scanner) skipComment() {
	for {
		r, ok := s.peek()

		if !ok || r == '\n' {
			return
		}

		s.advance()
	}
}

// Scans an identifier-shaped token from the current position.
// The resulting text is classified as a keyword when it matches one of the reserved words of the Koda configuration
// language.
func (s *scanner) scanIdentifierOrKeyword() token.Token {
	start, lexeme := s.position(), make([]rune, 0)

	for {
		r, ok := s.peek()

		if !ok || !isIdentifierContinue(r) {
			break
		}

		lexeme = append(lexeme, r)

		s.advance()
	}

	text := string(lexeme)

	if _, exists := keywords[text]; exists {
		return s.newToken(token.KindKeyword, text, start)
	}

	return s.newToken(token.KindIdentifier, text, start)
}

// Scans a double-quoted string literal from the current position.
// The surrounding quotes are removed from the token lexeme, while escape sequences are preserved exactly as written.
func (s *scanner) scanStringLiteral() (token.Token, error) {
	start := s.position()

	s.advance() // Consume the opening quote.

	lexeme := make([]rune, 0)

	for {
		r, ok := s.peek()

		if !ok {
			return token.Token{}, s.newError("Unterminated string literal.", start)
		}

		if r == '"' {
			s.advance()

			return s.newToken(token.KindLiteral, string(lexeme), start), nil
		}

		if r == '\\' {
			escaped, err := s.scanEscapeSequence()

			if err != nil {
				return token.Token{}, s.newError("Unterminated string literal.", start)
			}

			lexeme = append(lexeme, escaped...)

			continue
		}

		lexeme = append(lexeme, r)

		s.advance()
	}
}

// Scans a single-quoted character literal from the current position.
// The scanner accepts either a single rune or an escaped rune between the quotes and preserves the raw lexeme exactly
// as written.
func (s *scanner) scanCharacterLiteral() (token.Token, error) {
	start := s.position()

	s.advance() // Consume the opening quote.

	r, ok := s.peek()

	if !ok {
		return token.Token{}, s.newError("Unterminated character literal.", start)
	}

	if r == '\'' {
		return token.Token{}, s.newError("Character literal cannot be empty.", start)
	}

	lexeme := make([]rune, 0)

	if r == '\\' {
		escaped, err := s.scanEscapeSequence()

		if err != nil {
			return token.Token{}, s.newError("Unterminated character literal.", start)
		}

		lexeme = append(lexeme, escaped...)
	} else {
		lexeme = append(lexeme, r)

		s.advance()
	}

	closing, ok := s.peek()

	if !ok || closing != '\'' {
		return token.Token{}, s.newError("Unterminated character literal.", start)
	}

	s.advance()

	return s.newToken(token.KindCharacter, string(lexeme), start), nil
}

// Scans the current single-character structural symbol and returns it as a token.
// This is used for symbols such as `{`, `}`, `[`, `]`, and `,`.
func (s *scanner) scanSingleSymbol() token.Token {
	start := s.position()
	r, _ := s.advance()

	return s.newToken(token.KindSymbol, string(r), start)
}

// Consumes an escape sequence starting from the current backslash.
// Returns the backslash and escaped character, or an error if no character follows.
func (s *scanner) scanEscapeSequence() ([]rune, error) {
	backslash, _ := s.advance() // Consume the backslash.

	escaped, ok := s.peek()

	if !ok {
		return nil, fmt.Errorf("unterminated")
	}

	s.advance() // Consume the escaped character.

	return []rune{backslash, escaped}, nil
}

// Scans the range operator `..` from the current position.
// A single dot is rejected because it does not represent a valid token in the Koda configuration language.
func (s *scanner) scanRangeOperator() (token.Token, error) {
	start := s.position()

	first, okFirst := s.peekN(0)
	second, okSecond := s.peekN(1)

	if okFirst && okSecond && first == '.' && second == '.' {
		s.advance()
		s.advance()

		return s.newToken(token.KindSymbol, "..", start), nil
	}

	return token.Token{}, Error{
		Message: "Unexpected `.`. Expected the range operator `..`.",
		Span: loc.Span{
			Start: start,
			End: loc.Position{
				Line:   start.Line,
				Column: start.Column + 1,
			},
		},
	}
}

// Creates a new token with the given kind, lexeme and start position.
func (s *scanner) newToken(kind token.Kind, lexeme string, start loc.Position) token.Token {
	return token.Token{
		Kind:   kind,
		Lexeme: lexeme,
		Span:   s.spanFrom(start),
	}
}

// Creates an error with the provided message and a span from the given start position to the current scanner position.
func (s *scanner) newError(msg string, start loc.Position) Error {
	return Error{
		Message: msg,
		Span:    s.spanFrom(start),
	}
}

// Builds a lexical error for a rune that cannot begin any valid token in the language.
// The returned error span covers exactly that unexpected rune.
func (s *scanner) unexpectedCharacterError(r rune) Error {
	start := s.position()

	return Error{
		Message: fmt.Sprintf("Unexpected character `%c`.", r),
		Span: loc.Span{
			Start: start,
			End: loc.Position{
				Line:   start.Line,
				Column: start.Column + 1,
			},
		},
	}
}

// Reports whether the provided rune is treated as whitespace by the scanner.
// The recognized whitespace characters are space, tab, carriage return, and newline.
func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true

	default:
		return false
	}
}

// Reports whether the provided rune can begin an identifier.
// Identifiers may start with an ASCII letter or an underscore.
func isIdentifierStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// Reports whether the provided rune can continue an identifier.
// Identifier bodies may contain ASCII letters, digits, and underscores.
func isIdentifierContinue(r rune) bool {
	return isIdentifierStart(r) || (r >= '0' && r <= '9')
}

// Reports whether the provided rune is one of the recognized single-character
// structural symbols in the Koda configuration language.
func isSingleSymbol(r rune) bool {
	switch r {
	case '{', '}', '[', ']', ',':
		return true

	default:
		return false
	}
}
