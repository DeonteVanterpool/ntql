package ntql

import (
	"testing"
)

// TODO: Test positions of tokens
func TestLexer(t *testing.T) {
	lexer := NewLexer(`!tag.equals(hello) AND due.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")`)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Tokens: %s", tokens)

	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: crunch" ")"]
	expected := []Token{
		{Kind: TokenBang, Literal: "!"},
		{Kind: TokenSubject, Literal: "tag"},
		{Kind: TokenDot, Literal: "."},
		{Kind: TokenVerb, Literal: "equals"},
		{Kind: TokenLParen, Literal: "("},
		{Kind: TokenTag, Literal: "hello"},
		{Kind: TokenRParen, Literal: ")"},
		{Kind: TokenAnd, Literal: "AND"},
		{Kind: TokenSubject, Literal: "due"},
		{Kind: TokenDot, Literal: "."},
		{Kind: TokenVerb, Literal: "before"},
		{Kind: TokenLParen, Literal: "("},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ")"},
		{Kind: TokenAnd, Literal: "AND"},
		{Kind: TokenSubject, Literal: "title"},
		{Kind: TokenDot, Literal: "."},
		{Kind: TokenVerb, Literal: "startswith"},
		{Kind: TokenLParen, Literal: "("},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: "OR"},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ")"},
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Kind != expected[i].Kind {
			t.Errorf("Expected token type %s, got %s", expected[i].Kind, tok.Kind)
		}
		if tok.Literal != expected[i].Literal {
			t.Errorf("Expected token literal %s, got %s", expected[i].Literal, tok.Literal)
		}
	}
}
