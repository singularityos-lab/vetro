package vetro

import (
	"testing"
)

func TestLexer_Tokenize(t *testing.T) {
	input := `Window(id: "main", title: "Test") {
		Label("Hello")
	}`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that we have the expected tokens
	expectedTypes := []TokenType{
		TokenIdent,  // Window
		TokenLParen, // (
		TokenIdent,  // id
		TokenColon,  // :
		TokenString, // "main"
		TokenComma,  // ,
		TokenIdent,  // title
		TokenColon,  // :
		TokenString, // "Test"
		TokenRParen, // )
		TokenLBrace, // {
		TokenIdent,  // Label
		TokenLParen, // (
		TokenString, // "Hello"
		TokenRParen, // )
		TokenRBrace, // }
		TokenEOF,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expectedTypes[i] {
			t.Errorf("token %d: expected type %v, got %v (literal: %q)", i, expectedTypes[i], tok.Type, tok.Literal)
		}
	}
}

func TestLexer_TokenLiterals(t *testing.T) {
	input := `Window(id: "test_string", 123)`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find specific tokens and check their literals
	checks := map[string]TokenType{
		"Window":      TokenIdent,
		"id":          TokenIdent,
		"test_string": TokenString,
		"123":         TokenNumber,
	}

	for expected, tokType := range checks {
		found := false
		for _, tok := range tokens {
			if tok.Type == tokType && tok.Literal == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find token %s of type %v", expected, tokType)
		}
	}
}

func TestLexer_Position(t *testing.T) {
	input := `Window
	(id: 1)`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that line numbers are tracked
	// Window should be on line 1
	// ( should be on line 2
	if tokens[0].Line != 1 {
		t.Errorf("Window should be on line 1, got %d", tokens[0].Line)
	}
	if tokens[1].Line != 2 {
		t.Errorf("( should be on line 2, got %d", tokens[1].Line)
	}
}

func TestLexer_Modifier(t *testing.T) {
	input := `Label("test").cssClass("title")`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that we have dot and modifier tokens
	foundDot := false
	foundCSSClass := false
	for _, tok := range tokens {
		if tok.Type == TokenDot {
			foundDot = true
		}
		if tok.Type == TokenIdent && tok.Literal == "cssClass" {
			foundCSSClass = true
		}
	}
	if !foundDot {
		t.Error("expected to find dot token")
	}
	if !foundCSSClass {
		t.Error("expected to find cssClass identifier")
	}
}

func TestLexer_BooleanIdentifiers(t *testing.T) {
	input := `Switch(active: true, visible: false)`
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that true and false are parsed as identifiers
	foundTrue := false
	foundFalse := false
	for _, tok := range tokens {
		if tok.Type == TokenIdent && tok.Literal == "true" {
			foundTrue = true
		}
		if tok.Type == TokenIdent && tok.Literal == "false" {
			foundFalse = true
		}
	}
	if !foundTrue {
		t.Error("expected to find 'true' identifier")
	}
	if !foundFalse {
		t.Error("expected to find 'false' identifier")
	}
}
