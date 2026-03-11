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

// Package engine executes compiled Koda language profiles against target source files.
//
// The engine tokenizes target source text using compiled token definitions and then evaluates compiled structural rules
// to produce diagnostics.
package engine

import (
	"github.com/kdeconinck/koda/internal/compiler"
	"github.com/kdeconinck/koda/internal/loc"
)

// Analyze runs the compiled language profile against the provided target source text.
//
// The engine first tokenizes the source using the compiled token definitions and then evaluates the compiled rules over
// the matched tokens.
func Analyze(profile compiler.LanguageProfile, src string) Result {
	tokens, diagnostics := tokenize(profile, src)
	ruleDiagnostics := evaluateRules(profile, tokens)

	return Result{
		Tokens:      tokens,
		Diagnostics: append(diagnostics, ruleDiagnostics...),
	}
}

// A scanner state is used internally while tokenizing target source text.
//
// It tracks the current rune position together with line and column counters so matched runtime tokens and diagnostics
// can report precise spans.
type runtimeScanner struct {
	// The full input source, stored as runes for safe character-by-character scanning.
	src []rune

	// The index of the current rune in the input.
	index int

	// The current 1-based line number.
	line int

	// The current 1-based column number.
	column int
}

// Returns a new runtime scanner for the provided target source text.
// The scanner stores the input as runes and initializes its read position to the first character on line 1, column 1.
func newRuntimeScanner(src string) *runtimeScanner {
	return &runtimeScanner{
		src:    []rune(src),
		index:  0,
		line:   1,
		column: 1,
	}
}

// Tokenizes the target source text using the compiled token definitions.
// Tokens are matched left-to-right, preferring the longest match at each position and breaking ties by declaration
// order.
func tokenize(profile compiler.LanguageProfile, src string) ([]Token, []Diagnostic) {
	s := newRuntimeScanner(src)
	tokens := make([]Token, 0)
	diagnostics := make([]Diagnostic, 0)

	for !s.isAtEnd() {
		matched, consumed, diagnostic, ok := matchAtCurrentPosition(profile, s)

		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}

		if ok {
			tokens = append(tokens, matched...)
			s.consume(consumed)
			continue
		}

		s.advance()
	}

	return tokens, diagnostics
}

// Attempts to match one token at the current source position.
// The longest matching token definition is selected.
// When an enclosed token starts but does not terminate, a diagnostic is produced and the remainder of the source is
// consumed as that token.
func matchAtCurrentPosition(profile compiler.LanguageProfile, s *runtimeScanner) ([]Token, int, *Diagnostic, bool) {
	var bestToken Token
	bestMatched := false
	bestLength := 0

	for _, definition := range profile.TokenDefinitions {
		candidate, matchedLength, diagnostic, ok := tryMatchDefinition(definition, profile, s)
		if diagnostic != nil {
			return []Token{candidate}, matchedLength, diagnostic, true
		}

		if !ok {
			continue
		}

		if !bestMatched || matchedLength > bestLength {
			bestToken = candidate
			bestMatched = true
			bestLength = matchedLength
		}
	}

	if !bestMatched {
		return nil, 0, nil, false
	}

	return []Token{bestToken}, bestLength, nil, true
}

// Attempts to match the provided compiled token definition at the current source position.
// The returned length is measured in runes.
func tryMatchDefinition(definition compiler.TokenDefinition, profile compiler.LanguageProfile, s *runtimeScanner) (Token, int, *Diagnostic, bool) {
	startIndex := s.index
	startPos := s.position()

	switch definition.Kind {
	case compiler.TokenDefinitionKindLiteral:
		length, ok := matchLiteral(definition.Value, s)

		if !ok {
			return Token{}, 0, nil, false
		}

		return Token{
			DefinitionID: definition.ID,
			Name:         definition.Name,
			Lexeme:       string(s.src[startIndex : startIndex+length]),
			Span:         spanFromRunes(s.src, startIndex, length, startPos),
		}, length, nil, true

	case compiler.TokenDefinitionKindSequence:
		length := matchSequence(definition, profile, s)

		if length == 0 {
			return Token{}, 0, nil, false
		}

		return Token{
			DefinitionID: definition.ID,
			Name:         definition.Name,
			Lexeme:       string(s.src[startIndex : startIndex+length]),
			Span:         spanFromRunes(s.src, startIndex, length, startPos),
		}, length, nil, true

	case compiler.TokenDefinitionKindEnclosedBy:
		length, terminated, ok := matchEnclosedBy(definition, s)

		if !ok {
			return Token{}, 0, nil, false
		}

		token := Token{
			DefinitionID: definition.ID,
			Name:         definition.Name,
			Lexeme:       string(s.src[startIndex : startIndex+length]),
			Span:         spanFromRunes(s.src, startIndex, length, startPos),
		}

		if !terminated {
			diagnostic := Diagnostic{
				Message: "Unterminated enclosed token.",
				Span:    token.Span,
			}

			return token, length, &diagnostic, true
		}

		return token, length, nil, true
	}

	return Token{}, 0, nil, false
}

// Evaluates compiled structural rules against the matched runtime token stream.
// A diagnostic is produced each time a rule is violated.
func evaluateRules(profile compiler.LanguageProfile, tokens []Token) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)

	for idx, tok := range tokens {
		for _, rule := range profile.Rules {
			if tok.DefinitionID != rule.MatchTokenID {
				continue
			}

			nextToken, hasNext := nextToken(tokens, idx)

			switch rule.Constraint.Kind {
			case compiler.ConstraintKindMustBeFollowedBy:
				if !hasNext || nextToken.DefinitionID != rule.Constraint.TokenID {
					diagnostics = append(diagnostics, Diagnostic{
						Message: rule.ErrorMessage,
						Span:    tok.Span,
					})
				}

			case compiler.ConstraintKindCannotBeFollowedBy:
				if hasNext && nextToken.DefinitionID == rule.Constraint.TokenID {
					diagnostics = append(diagnostics, Diagnostic{
						Message: rule.ErrorMessage,
						Span:    tok.Span,
					})
				}
			}
		}
	}

	return diagnostics
}

// Returns the next token after the current index, when one exists.
func nextToken(tokens []Token, current int) (Token, bool) {
	next := current + 1

	if next >= len(tokens) {
		return Token{}, false
	}

	return tokens[next], true
}

// Matches a literal token definition against the current scanner position.
// The returned length is measured in runes.
func matchLiteral(value string, s *runtimeScanner) (int, bool) {
	want := []rune(value)

	if len(want) == 0 {
		return 0, false
	}

	if s.index+len(want) > len(s.src) {
		return 0, false
	}

	for idx := range want {
		if s.src[s.index+idx] != want[idx] {
			return 0, false
		}
	}

	return len(want), true
}

// Matches a sequence token definition against the current scanner position.
// The returned length is the number of consecutive runes that belong to the compiled charset.
func matchSequence(definition compiler.TokenDefinition, profile compiler.LanguageProfile, s *runtimeScanner) int {
	charSetIndex := int(definition.CharSetID)

	if charSetIndex < 0 || charSetIndex >= len(profile.CharSets) {
		return 0
	}

	characters := profile.CharSets[charSetIndex].Characters
	length := 0

	for idx := s.index; idx < len(s.src); idx++ {
		if _, ok := characters[s.src[idx]]; !ok {
			break
		}

		length++
	}

	return length
}

// Matches an enclosed token definition against the current scanner position.
// The returned length is measured in runes. The second result reports whether the closing delimiter was found.
func matchEnclosedBy(definition compiler.TokenDefinition, s *runtimeScanner) (int, bool, bool) {
	startRunes := []rune(definition.Start)
	endRunes := []rune(definition.End)

	if len(startRunes) == 0 || len(endRunes) == 0 {
		return 0, false, false
	}

	startLength, ok := matchLiteral(definition.Start, s)

	if !ok {
		return 0, false, false
	}

	searchIndex := s.index + startLength

	for searchIndex <= len(s.src)-len(endRunes) {
		if hasRunePrefixAt(s.src, searchIndex, endRunes) {
			totalLength := (searchIndex - s.index) + len(endRunes)

			return totalLength, true, true
		}

		searchIndex++
	}

	return len(s.src) - s.index, false, true
}

// Reports whether the provided rune slice has the given prefix at the provided
// starting index.
func hasRunePrefixAt(src []rune, start int, want []rune) bool {
	if start+len(want) > len(src) {
		return false
	}

	for idx := range want {
		if src[start+idx] != want[idx] {
			return false
		}
	}

	return true
}

// Reports whether the runtime scanner has reached the end of the input.
func (s *runtimeScanner) isAtEnd() bool {
	return s.index >= len(s.src)
}

// Consumes the requested number of runes from the current scanner position.
// Line and column tracking are updated automatically.
func (s *runtimeScanner) consume(length int) {
	for idx := 0; idx < length; idx++ {
		s.advance()
	}
}

// Consumes one rune from the current scanner position.
// Line and column tracking are updated automatically.
func (s *runtimeScanner) advance() {
	if s.isAtEnd() {
		return
	}

	r := s.src[s.index]
	s.index++

	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}
}

// Returns the current scanner position as a source location.
func (s *runtimeScanner) position() loc.Position {
	return loc.Position{
		Line:   s.line,
		Column: s.column,
	}
}

// Builds a source span for the provided rune range starting at the provided
// source position.
func spanFromRunes(src []rune, startIndex, length int, startPos loc.Position) loc.Span {
	line := startPos.Line
	column := startPos.Column

	for idx := startIndex; idx < startIndex+length; idx++ {
		if src[idx] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}

	return loc.Span{
		Start: startPos,
		End: loc.Position{
			Line:   line,
			Column: column,
		},
	}
}
