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

// Package compiler defines the compiled runtime model of the Koda configuration language.
//
// The types in this package represent a validated language profile in a runtime-friendly form.
// References are resolved, character sets are normalized, and token definitions and rules are prepared for efficient
// execution by the engine.
package compiler

import "github.com/kdeconinck/koda/internal/ast"

// Compile transforms a validated AST language profile into a compiled runtime language profile.
//
// Compilation resolves all name-based references, assigns stable identifiers, and expands charset declarations into
// explicit character sets.
func Compile(profile ast.LanguageProfile) LanguageProfile {
	compiledCharSets, charSetIndex := compileCharSets(profile)
	compiledTokenDefinitions, tokenIndex := compileTokenDefinitions(profile, charSetIndex)
	compiledRules := compileRules(profile, tokenIndex)

	return LanguageProfile{
		Name:             profile.Name,
		Extensions:       append([]string(nil), profile.Extensions...),
		CharSets:         compiledCharSets,
		TokenDefinitions: compiledTokenDefinitions,
		Rules:            compiledRules,
		Span:             profile.Span,
	}
}

// Compiles parsed charset declarations into normalized runtime charsets.
// Each compiled charset receives a stable ID based on declaration order.
func compileCharSets(profile ast.LanguageProfile) ([]CharSet, map[string]CharSetID) {
	compiled := make([]CharSet, 0, len(profile.CharSets))
	index := make(map[string]CharSetID, len(profile.CharSets))

	for idx, charSet := range profile.CharSets {
		id := CharSetID(idx)
		index[charSet.Name] = id

		characters := make(map[rune]struct{})

		for _, item := range charSet.Items {
			switch item.Kind {
			case ast.CharSetItemKindSingle:
				characters[item.Value] = struct{}{}

			case ast.CharSetItemKindRange:
				for ch := item.Start; ch <= item.End; ch++ {
					characters[ch] = struct{}{}
				}
			}
		}

		compiled = append(compiled, CharSet{
			ID:         id,
			Name:       charSet.Name,
			Characters: characters,
			Span:       charSet.Span,
		})
	}

	return compiled, index
}

// Compiles parsed token definitions into runtime token definitions.
// Sequence definitions are resolved to compiled charset IDs.
func compileTokenDefinitions(profile ast.LanguageProfile, charSetIndex map[string]CharSetID) ([]TokenDefinition, map[string]TokenDefinitionID) {
	compiled := make([]TokenDefinition, 0, len(profile.Tokens.Definitions))
	index := make(map[string]TokenDefinitionID, len(profile.Tokens.Definitions))

	for idx, definition := range profile.Tokens.Definitions {
		id := TokenDefinitionID(idx)
		index[definition.Name] = id

		compiledDefinition := TokenDefinition{
			ID:   id,
			Name: definition.Name,
			Span: definition.Span,
		}

		switch definition.Kind {
		case ast.TokenDefinitionKindLiteral:
			compiledDefinition.Kind = TokenDefinitionKindLiteral
			compiledDefinition.Value = definition.Value
		case ast.TokenDefinitionKindSequence:
			charSetID, _ := charSetIndex[definition.CharSetName]

			compiledDefinition.Kind = TokenDefinitionKindSequence
			compiledDefinition.CharSetID = charSetID
		default:
			compiledDefinition.Kind = TokenDefinitionKindEnclosedBy
			compiledDefinition.Start = definition.Start
			compiledDefinition.End = definition.End
		}

		compiled = append(compiled, compiledDefinition)
	}

	return compiled, index
}

// Compiles parsed rules into runtime rules.
// Rule references are resolved to compiled token-definition IDs.
func compileRules(profile ast.LanguageProfile, tokenIndex map[string]TokenDefinitionID) []Rule {
	compiled := make([]Rule, 0, len(profile.Rules.Rules))

	for _, rule := range profile.Rules.Rules {
		matchTokenID := tokenIndex[rule.MatchTokenName]

		constraintTokenID := tokenIndex[rule.Constraint.TokenName]

		compiledConstraint := Constraint{
			TokenID: constraintTokenID,
			Span:    rule.Constraint.Span,
		}

		if rule.Constraint.Kind == ast.ConstraintKindMustBeFollowedBy {
			compiledConstraint.Kind = ConstraintKindMustBeFollowedBy
		} else {
			compiledConstraint.Kind = ConstraintKindCannotBeFollowedBy
		}

		compiled = append(compiled, Rule{
			Name:         rule.Name,
			MatchTokenID: matchTokenID,
			Constraint:   compiledConstraint,
			ErrorMessage: rule.ErrorMessage,
			Span:         rule.Span,
		})
	}

	return compiled
}
