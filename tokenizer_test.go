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
        {Type: Bang, Literal: ""},
        {Type: Identifier, Literal: "tag"},
        {Type: Dot, Literal: ""},
        {Type: Identifier, Literal: "equals"},
        {Type: LParen, Literal: ""},
        {Type: Identifier, Literal: "hello"},
        {Type: RParen, Literal: ""},
        {Type: And, Literal: ""},
        {Type: Identifier, Literal: "date"},
        {Type: Dot, Literal: ""},
        {Type: Identifier, Literal: "before"},
        {Type: LParen, Literal: ""},
        {Type: Date, Literal: "2021-01-01"},
        {Type: And, Literal: ""},
        {Type: Identifier, Literal: "title"},
        {Type: Dot, Literal: ""},
        {Type: Identifier, Literal: "startswith"},
        {Type: LParen, Literal: ""},
        {Type: String, Literal: "bar"},
        {Type: Or, Literal: ""},
        {Type: String, Literal: "c\"\\runch"},
        {Type: RParen, Literal: ""},
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


