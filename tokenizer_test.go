package tbql

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
    Tokenizer := NewTokenizer(`!tag.equals(hello) AND date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")`)
	tokens, err := Tokenizer.Tokenize()
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

    t.Logf("Tokens: %s", tokens)

    // ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: crunch" ")"]
    expected := []Token{
        {Type: TokenBang, Literal: ""},
        {Type: TokenIdentifier, Literal: "tag"},
        {Type: TokenDot, Literal: ""},
        {Type: TokenIdentifier, Literal: "equals"},
        {Type: TokenLParen, Literal: ""},
        {Type: TokenIdentifier, Literal: "hello"},
        {Type: TokenRParen, Literal: ""},
        {Type: TokenAnd, Literal: ""},
        {Type: TokenIdentifier, Literal: "date"},
        {Type: TokenDot, Literal: ""},
        {Type: TokenIdentifier, Literal: "before"},
        {Type: TokenLParen, Literal: ""},
        {Type: TokenDate, Literal: "2021-01-01"},
        {Type: TokenAnd, Literal: ""},
        {Type: TokenIdentifier, Literal: "title"},
        {Type: TokenDot, Literal: ""},
        {Type: TokenIdentifier, Literal: "startswith"},
        {Type: TokenLParen, Literal: ""},
        {Type: TokenString, Literal: "bar"},
        {Type: TokenOr, Literal: ""},
        {Type: TokenString, Literal: "c\"\\runch"},
        {Type: TokenRParen, Literal: ""},
    }

    if len(tokens) != len(expected) {
        t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
    }

    for i, tok := range tokens {
        if tok.Type != expected[i].Type {
            t.Errorf("Expected token type %s, got %s", expected[i].Type, tok.Type)
        }
        if tok.Literal != expected[i].Literal {
            t.Errorf("Expected token literal %s, got %s", expected[i].Literal, tok.Literal)
        }
    }
}

