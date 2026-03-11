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

// Package validator validates parsed Koda configuration models.
//
// The validator performs semantic analysis on the parsed AST.
// It checks that names are unique, references can be resolved, and declarations are structurally meaningful after
// parsing has completed.
package validator

import (
	"fmt"

	"github.com/kdeconinck/koda/internal/ast"
	"github.com/kdeconinck/koda/internal/loc"
)

// Validate verifies that the parsed language profile is semantically valid.
//
// Validation currently checks:
//
//   - Duplicate charset names.
//   - Duplicate token names.
//   - Duplicate rule names.
//   - Valid charset ranges.
//   - Unresolved charset references in sequence token definitions.
//   - Unresolved token references in rules.
func Validate(profile ast.LanguageProfile) error {
	v := &validator{
		charSets: make(map[string]struct{}),
		tokens:   make(map[string]struct{}),
	}

	if err := v.collectAndValidateCharSets(profile); err != nil {
		return err
	}

	if err := v.collectAndValidateTokens(profile); err != nil {
		return err
	}

	if err := v.validateUniqueRuleNames(profile); err != nil {
		return err
	}

	if err := v.validateTokenReferences(profile); err != nil {
		return err
	}

	if err := v.validateRuleReferences(profile); err != nil {
		return err
	}

	return nil
}

// A validator holds the symbol tables required to resolve references across a language profile.
type validator struct {
	// The set of unique charset names declared in the language profile.
	charSets map[string]struct{}

	// The set of unique token names declared in the language profile.
	tokens map[string]struct{}
}

// Checks that charset names are unique and that all character ranges are valid (non-descending).
func (v *validator) collectAndValidateCharSets(profile ast.LanguageProfile) error {
	for _, charSet := range profile.CharSets {
		if _, exists := v.charSets[charSet.Name]; exists {
			return v.newError(fmt.Sprintf("Duplicate charset name `%s`.", charSet.Name), charSet.Span)
		}

		v.charSets[charSet.Name] = struct{}{}

		for _, item := range charSet.Items {
			if item.Kind == ast.CharSetItemKindRange && item.Start > item.End {
				return v.newError(
					fmt.Sprintf("Invalid charset range `%c`..`%c` in charset `%s`.", item.Start, item.End, charSet.Name),
					item.Span,
				)
			}
		}
	}

	return nil
}

// Checks that token names are unique within the token section.
func (v *validator) collectAndValidateTokens(profile ast.LanguageProfile) error {
	for _, definition := range profile.Tokens.Definitions {
		if _, exists := v.tokens[definition.Name]; exists {
			return v.newError(fmt.Sprintf("Duplicate token definition name `%s`.", definition.Name), definition.Span)
		}

		v.tokens[definition.Name] = struct{}{}
	}

	return nil
}

// Checks that rule names are unique within the rule section.
func (v *validator) validateUniqueRuleNames(profile ast.LanguageProfile) error {
	seen := make(map[string]struct{})

	for _, rule := range profile.Rules.Rules {
		if _, exists := seen[rule.Name]; exists {
			return v.newError(fmt.Sprintf("Duplicate rule name `%s`.", rule.Name), rule.Span)
		}

		seen[rule.Name] = struct{}{}
	}

	return nil
}

// Validates that sequence tokens reference charsets that actually exist.
func (v *validator) validateTokenReferences(profile ast.LanguageProfile) error {
	for _, def := range profile.Tokens.Definitions {
		if def.Kind != ast.TokenDefinitionKindSequence {
			continue
		}

		if _, exists := v.charSets[def.CharSetName]; !exists {
			return v.newError(
				fmt.Sprintf("Unknown charset reference `%s` in token definition `%s`.", def.CharSetName, def.Name),
				def.Span,
			)
		}
	}

	return nil
}

// Validates that rules reference tokens that actually exist in both MATCH clauses and constraints.
func (v *validator) validateRuleReferences(profile ast.LanguageProfile) error {
	for _, rule := range profile.Rules.Rules {
		// Validate the target of the MATCH clause.
		if _, exists := v.tokens[rule.MatchTokenName]; !exists {
			return v.newError(
				fmt.Sprintf("Unknown token reference `%s` in MATCH clause of rule `%s`.", rule.MatchTokenName, rule.Name),
				rule.Span,
			)
		}

		// Validate the target of the structural constraint.
		if _, exists := v.tokens[rule.Constraint.TokenName]; !exists {
			return v.newError(
				fmt.Sprintf("Unknown token reference `%s` in constraint of rule `%s`.", rule.Constraint.TokenName, rule.Name),
				rule.Constraint.Span,
			)
		}
	}

	return nil
}

// Creates a validation error with the provided message and span.
func (v *validator) newError(msg string, span loc.Span) Error {
	return Error{
		Message: msg,
		Span:    span,
	}
}
