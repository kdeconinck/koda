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

// Package parser transforms lexical tokens into parsed Koda configuration models.
//
// The parser performs syntactic analysis only.
// It validates that tokens appear in a structurally correct order and shape, but it does not perform semantic
// validation such as checking for duplicate names or unresolved references.
package parser

import (
	"fmt"
	"slices"

	"github.com/kdeconinck/koda/internal/ast"
	"github.com/kdeconinck/koda/internal/loc"
	"github.com/kdeconinck/koda/internal/token"
)

// Parse transforms a scanned token stream into a parsed Koda language profile.
//
// Parse expects one complete `LANG` block and returns an error when the token stream does not match the grammar of the
// Koda configuration language.
func Parse(tokens []token.Token) (ast.LanguageProfile, error) {
	return newParser(tokens).parseLanguageProfile()
}

// A parser holds the state required to consume a token stream from left to right.
//
// Keeping the implementation stateful makes it easier to express grammar rules as small helper methods and to report
// precise parse errors at the current position.
type parser struct {
	// The full token stream being parsed.
	tokens []token.Token

	// The index of the current token in the input.
	index int
}

// Returns a new parser for the provided token stream.
// Parsing always starts at the first token in the slice.
func newParser(tokens []token.Token) *parser {
	return &parser{
		tokens: tokens,
		index:  0,
	}
}

// Parses a complete language profile from the token stream.
// The current implementation expects exactly one `LANG` block and requires both a `SECTION TOKENS` block and a
// `SECTION RULES` block.
func (p *parser) parseLanguageProfile() (ast.LanguageProfile, error) {
	langToken, err := p.expectKeyword("LANG")

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	nameToken, err := p.expectLiteral()

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	_, err = p.expectKeyword("EXTENSIONS")

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	extensions, err := p.parseExtensions()

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	_, err = p.expectSymbol("{")

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	charSets := make([]ast.CharSet, 0)

	for p.isNextCharsetDeclaration() {
		charSet, err := p.parseCharSet()

		if err != nil {
			return ast.LanguageProfile{}, err
		}

		charSets = append(charSets, charSet)
	}

	tokenSection, err := p.parseTokenSection()

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	ruleSection, err := p.parseRuleSection()

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	closingBrace, err := p.expectSymbol("}")

	if err != nil {
		return ast.LanguageProfile{}, err
	}

	if !p.isAtEnd() {
		tok, _ := p.peek()

		return ast.LanguageProfile{}, p.newError("Unexpected tokens after the end of the `LANG` block.", tok.Span)
	}

	return ast.LanguageProfile{
		Name:       nameToken.Lexeme,
		Extensions: extensions,
		CharSets:   charSets,
		Tokens:     tokenSection,
		Rules:      ruleSection,
		Span: loc.Span{
			Start: langToken.Span.Start,
			End:   closingBrace.Span.End,
		},
	}, nil
}

// Parses the `EXTENSIONS [...]` list of a language profile.
// The list may be empty or contain one or more string literals separated by commas.
func (p *parser) parseExtensions() ([]string, error) {
	_, err := p.expectSymbol("[")

	if err != nil {
		return nil, err
	}

	extensions := make([]string, 0)

	if p.checkSymbol("]") {
		p.advance()

		return extensions, nil
	}

	for {
		literal, err := p.expectLiteral()

		if err != nil {
			return nil, err
		}

		extensions = append(extensions, literal.Lexeme)

		if p.checkSymbol(",") {
			p.advance()

			continue
		}

		break
	}

	_, err = p.expectSymbol("]")

	if err != nil {
		return nil, err
	}

	return extensions, nil
}

// Reports whether the current token begins a charset declaration.
// A charset declaration starts with `DEFINE CHARSET`.
func (p *parser) isNextCharsetDeclaration() bool {
	first, ok := p.peek()

	if !ok {
		return false
	}

	second, ok := p.peekN(1)

	if !ok {
		return false
	}

	return first.Kind == token.KindKeyword &&
		first.Lexeme == "DEFINE" &&
		second.Kind == token.KindKeyword &&
		second.Lexeme == "CHARSET"
}

// Parses a charset declaration of the form `DEFINE CHARSET "name" VALUES ['a'..'z', '_']`.
// NOTE: This function assumes that the initial `DEFINE CHARSET` keywords have already been matched, not consumed.
func (p *parser) parseCharSet() (ast.CharSet, error) {
	defineToken := p.advance()

	p.advance()

	nameToken, err := p.expectLiteral()

	if err != nil {
		return ast.CharSet{}, err
	}

	_, err = p.expectKeyword("VALUES")

	if err != nil {
		return ast.CharSet{}, err
	}

	_, err = p.expectSymbol("[")

	if err != nil {
		return ast.CharSet{}, err
	}

	items := make([]ast.CharSetItem, 0)

	if !p.checkSymbol("]") {
		for {
			item, err := p.parseCharSetItem()

			if err != nil {
				return ast.CharSet{}, err
			}

			items = append(items, item)

			if p.checkSymbol(",") {
				p.advance()

				continue
			}

			break
		}
	}

	closingBracket, err := p.expectSymbol("]")

	if err != nil {
		return ast.CharSet{}, err
	}

	return ast.CharSet{
		Name:  nameToken.Lexeme,
		Items: items,
		Span: loc.Span{
			Start: defineToken.Span.Start,
			End:   closingBracket.Span.End,
		},
	}, nil
}

// Parses a single charset item.
// Supported forms are a single character literal or a character range.
func (p *parser) parseCharSetItem() (ast.CharSetItem, error) {
	firstToken, err := p.expectCharacter()

	if err != nil {
		return ast.CharSetItem{}, err
	}

	firstRune, err := decodeCharacterLexeme(firstToken.Lexeme)

	if err != nil {
		return ast.CharSetItem{}, p.newError(err.Error(), firstToken.Span)
	}

	if p.checkSymbol("..") {
		p.advance()

		lastToken, err := p.expectCharacter()

		if err != nil {
			return ast.CharSetItem{}, err
		}

		lastRune, err := decodeCharacterLexeme(lastToken.Lexeme)

		if err != nil {
			return ast.CharSetItem{}, p.newError(err.Error(), lastToken.Span)
		}

		return ast.CharSetItem{
			Kind:  ast.CharSetItemKindRange,
			Start: firstRune,
			End:   lastRune,
			Span: loc.Span{
				Start: firstToken.Span.Start,
				End:   lastToken.Span.End,
			},
		}, nil
	}

	return ast.CharSetItem{
		Kind:  ast.CharSetItemKindSingle,
		Value: firstRune,
		Span:  firstToken.Span,
	}, nil
}

// Parses a `SECTION TOKENS` block.
// Token definitions are parsed until the closing brace of the section is reached.
func (p *parser) parseTokenSection() (ast.TokenSection, error) {
	sectionToken, err := p.expectKeyword("SECTION")

	if err != nil {
		return ast.TokenSection{}, err
	}

	_, err = p.expectKeyword("TOKENS")

	if err != nil {
		return ast.TokenSection{}, err
	}

	_, err = p.expectSymbol("{")

	if err != nil {
		return ast.TokenSection{}, err
	}

	definitions := make([]ast.TokenDefinition, 0)

	for !p.checkSymbol("}") {
		definition, err := p.parseTokenDefinition()

		if err != nil {
			return ast.TokenSection{}, err
		}

		definitions = append(definitions, definition)
	}

	closingBrace, _ := p.expectSymbol("}")

	return ast.TokenSection{
		Definitions: definitions,
		Span: loc.Span{
			Start: sectionToken.Span.Start,
			End:   closingBrace.Span.End,
		},
	}, nil
}

// Parses a token definition inside `SECTION TOKENS`.
// Supported forms are `LITERAL`, `SEQUENCE`, and `ENCLOSED_BY`.
func (p *parser) parseTokenDefinition() (ast.TokenDefinition, error) {
	defineToken, err := p.expectKeyword("DEFINE")

	if err != nil {
		return ast.TokenDefinition{}, err
	}

	nameToken, err := p.expectLiteral()

	if err != nil {
		return ast.TokenDefinition{}, err
	}

	kindToken, err := p.expectKeywordOneOf("LITERAL", "SEQUENCE", "ENCLOSED_BY")

	switch kindToken.Lexeme {
	case "LITERAL":
		valueToken, err := p.expectLiteral()

		if err != nil {
			return ast.TokenDefinition{}, err
		}

		return ast.TokenDefinition{
			Kind:  ast.TokenDefinitionKindLiteral,
			Name:  nameToken.Lexeme,
			Value: valueToken.Lexeme,
			Span: loc.Span{
				Start: defineToken.Span.Start,
				End:   valueToken.Span.End,
			},
		}, nil

	case "SEQUENCE":
		charSetToken, err := p.expectLiteral()

		if err != nil {
			return ast.TokenDefinition{}, err
		}

		return ast.TokenDefinition{
			Kind:        ast.TokenDefinitionKindSequence,
			Name:        nameToken.Lexeme,
			CharSetName: charSetToken.Lexeme,
			Span: loc.Span{
				Start: defineToken.Span.Start,
				End:   charSetToken.Span.End,
			},
		}, nil

	case "ENCLOSED_BY":
		startToken, err := p.expectCharacter()

		if err != nil {
			return ast.TokenDefinition{}, err
		}

		endToken, err := p.expectCharacter()

		if err != nil {
			return ast.TokenDefinition{}, err
		}

		startRune, err := decodeCharacterLexeme(startToken.Lexeme)

		if err != nil {
			return ast.TokenDefinition{}, p.newError(err.Error(), startToken.Span)
		}

		endRune, err := decodeCharacterLexeme(endToken.Lexeme)

		if err != nil {
			return ast.TokenDefinition{}, p.newError(err.Error(), endToken.Span)
		}

		return ast.TokenDefinition{
			Kind:  ast.TokenDefinitionKindEnclosedBy,
			Name:  nameToken.Lexeme,
			Start: string(startRune),
			End:   string(endRune),
			Span: loc.Span{
				Start: defineToken.Span.Start,
				End:   endToken.Span.End,
			},
		}, nil
	}

	return ast.TokenDefinition{}, err
}

// Parses a `SECTION RULES` block.
// Rules are parsed until the closing brace of the section is reached.
func (p *parser) parseRuleSection() (ast.RuleSection, error) {
	sectionToken, err := p.expectKeyword("SECTION")

	if err != nil {
		return ast.RuleSection{}, err
	}

	_, err = p.expectKeyword("RULES")

	if err != nil {
		return ast.RuleSection{}, err
	}

	_, err = p.expectSymbol("{")

	if err != nil {
		return ast.RuleSection{}, err
	}

	rules := make([]ast.Rule, 0)

	for !p.checkSymbol("}") {
		rule, err := p.parseRule()

		if err != nil {
			return ast.RuleSection{}, err
		}

		rules = append(rules, rule)
	}

	closingBrace := p.advance()

	return ast.RuleSection{
		Rules: rules,
		Span: loc.Span{
			Start: sectionToken.Span.Start,
			End:   closingBrace.Span.End,
		},
	}, nil
}

// Parses a single rule block inside `SECTION RULES`.
func (p *parser) parseRule() (ast.Rule, error) {
	ruleToken, err := p.expectKeyword("RULE")

	if err != nil {
		return ast.Rule{}, err
	}

	nameToken, err := p.expectLiteral()

	if err != nil {
		return ast.Rule{}, err
	}

	_, err = p.expectSymbol("{")

	if err != nil {
		return ast.Rule{}, err
	}

	_, err = p.expectKeyword("MATCH")

	if err != nil {
		return ast.Rule{}, err
	}

	matchToken, err := p.expectLiteral()

	if err != nil {
		return ast.Rule{}, err
	}

	constraint, err := p.parseConstraint()

	if err != nil {
		return ast.Rule{}, err
	}

	_, err = p.expectKeyword("ERROR")

	if err != nil {
		return ast.Rule{}, err
	}

	errorToken, err := p.expectLiteral()

	if err != nil {
		return ast.Rule{}, err
	}

	closingBrace, err := p.expectSymbol("}")

	if err != nil {
		return ast.Rule{}, err
	}

	return ast.Rule{
		Name:           nameToken.Lexeme,
		MatchTokenName: matchToken.Lexeme,
		Constraint:     constraint,
		ErrorMessage:   errorToken.Lexeme,
		Span: loc.Span{
			Start: ruleToken.Span.Start,
			End:   closingBrace.Span.End,
		},
	}, nil
}

// Parses the constraint declaration inside a rule block.
// Supported forms are `MUST_BE_FOLLOWED_BY` and `CANNOT_BE_FOLLOWED_BY`.
func (p *parser) parseConstraint() (ast.Constraint, error) {
	keywordToken, err := p.expectKeywordOneOf("MUST_BE_FOLLOWED_BY", "CANNOT_BE_FOLLOWED_BY")

	if err != nil {
		return ast.Constraint{}, err
	}

	targetToken, err := p.expectLiteral()

	if err != nil {
		return ast.Constraint{}, err
	}

	kind := ast.ConstraintKindMustBeFollowedBy

	if keywordToken.Lexeme == "CANNOT_BE_FOLLOWED_BY" {
		kind = ast.ConstraintKindCannotBeFollowedBy
	}

	return ast.Constraint{
		Kind:      kind,
		TokenName: targetToken.Lexeme,
		Span: loc.Span{
			Start: keywordToken.Span.Start,
			End:   targetToken.Span.End,
		},
	}, nil
}

// Returns the current token without consuming it.
// The boolean result is false when the parser has reached the end of the input.
func (p *parser) peek() (token.Token, bool) {
	if p.index >= len(p.tokens) {
		return token.Token{}, false
	}

	return p.tokens[p.index], true
}

// Returns the token at the given lookahead offset without consuming it.
// The boolean result is false when the requested lookahead is out of range.
func (p *parser) peekN(offset int) (token.Token, bool) {
	index := p.index + offset

	if index >= len(p.tokens) {
		return token.Token{}, false
	}

	return p.tokens[index], true
}

// Consumes the current token and advances the parser by one position.
func (p *parser) advance() token.Token {
	tok, _ := p.peek()

	p.index++

	return tok
}

// Reports whether the parser has consumed the complete token stream.
func (p *parser) isAtEnd() bool {
	return p.index >= len(p.tokens)
}

// Reports whether the current token is the expected symbol.
func (p *parser) checkSymbol(want string) bool {
	tok, ok := p.peek()

	if !ok {
		return false
	}

	return tok.Kind == token.KindSymbol && tok.Lexeme == want
}

// Returns the current token when it is the expected keyword.
// A parser error is returned when the token stream ends or when another token kind or lexeme appears instead.
func (p *parser) expectKeyword(want string) (token.Token, error) {
	tok, ok := p.peek()

	if !ok {
		return token.Token{}, p.newError(fmt.Sprintf("Expected keyword `%s`, but reached the end of input.", want), eofSpan(p.tokens))
	}

	if tok.Kind != token.KindKeyword || tok.Lexeme != want {
		return token.Token{}, p.newError(fmt.Sprintf("Expected keyword `%s`.", want), tok.Span)
	}

	return p.advance(), nil
}

// Returns the current token when it matches one of the expected keywords.
// A parser error is returned when the token stream ends or when another token kind or lexeme appears instead.
func (p *parser) expectKeywordOneOf(wants ...string) (token.Token, error) {
	tok, ok := p.peek()

	if !ok {
		return token.Token{}, p.newError(fmt.Sprintf("Expected one of %s, but reached the end of input.", formatExpectedKeywords(wants)), eofSpan(p.tokens))
	}

	if tok.Kind != token.KindKeyword {
		return token.Token{}, p.newError(fmt.Sprintf("Expected one of %s.", formatExpectedKeywords(wants)), tok.Span)
	}

	if slices.Contains(wants, tok.Lexeme) {
		return p.advance(), nil
	}

	return token.Token{}, p.newError(fmt.Sprintf("Expected one of %s.", formatExpectedKeywords(wants)), tok.Span)
}

// Returns the current token when it is a string literal.
// A parser error is returned when the token stream ends or when another token kind appears instead.
func (p *parser) expectLiteral() (token.Token, error) {
	tok, ok := p.peek()

	if !ok {
		return token.Token{}, p.newError("Expected literal, but reached the end of input.", eofSpan(p.tokens))
	}

	if tok.Kind != token.KindLiteral {
		return token.Token{}, p.newError("Expected literal.", tok.Span)
	}

	return p.advance(), nil
}

// Returns the current token when it is a character literal.
// A parser error is returned when the token stream ends or when another token kind appears instead.
func (p *parser) expectCharacter() (token.Token, error) {
	tok, ok := p.peek()

	if !ok {
		return token.Token{}, p.newError("Expected character literal, but reached the end of input.", eofSpan(p.tokens))
	}

	if tok.Kind != token.KindCharacter {
		return token.Token{}, p.newError("Expected character literal.", tok.Span)
	}

	return p.advance(), nil
}

// Returns the current token when it is the expected symbol.
// A parser error is returned when the token stream ends or when another token kind or lexeme appears instead.
func (p *parser) expectSymbol(want string) (token.Token, error) {
	tok, ok := p.peek()

	if !ok {
		return token.Token{}, p.newError(fmt.Sprintf("Expected symbol `%s`, but reached the end of input.", want), eofSpan(p.tokens))
	}

	if tok.Kind != token.KindSymbol || tok.Lexeme != want {
		return token.Token{}, p.newError(fmt.Sprintf("Expected symbol `%s`.", want), tok.Span)
	}

	return p.advance(), nil
}

// Creates an error with the provided message and span.
func (p *parser) newError(msg string, span loc.Span) Error {
	return Error{
		Message: msg,
		Span:    span,
	}
}

// Returns a fallback span representing the logical end of the token stream.
// When the input is empty, the span points to line 1, column 1.
func eofSpan(tokens []token.Token) loc.Span {
	if len(tokens) == 0 {
		pos := loc.Position{
			Line:   1,
			Column: 1,
		}

		return loc.Span{Start: pos, End: pos}
	}

	last := tokens[len(tokens)-1]

	return loc.Span{
		Start: last.Span.End,
		End:   last.Span.End,
	}
}

// Formats a list of expected keyword values for use in parser error messages.
func formatExpectedKeywords(wants []string) string {
	result := ""

	for idx, want := range wants {
		if idx > 0 {
			if idx == len(wants)-1 {
				result += " or "
			} else {
				result += ", "
			}
		}

		result += fmt.Sprintf("`%s`", want)
	}

	return result
}

// Decodes the lexeme of a scanned character token into its actual rune value.
// Supported escape sequences match the scanner's accepted character-literal forms.
func decodeCharacterLexeme(lexeme string) (rune, error) {
	runes := []rune(lexeme)

	if len(runes) == 1 {
		return runes[0], nil
	}

	switch runes[1] {
	case '\\':
		return '\\', nil

	case '\'':
		return '\'', nil

	case '"':
		return '"', nil

	case 'n':
		return '\n', nil

	case 'r':
		return '\r', nil

	case 't':
		return '\t', nil

	default:
		return 0, fmt.Errorf("Unsupported character escape %q", lexeme)
	}
}
