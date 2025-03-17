package tbql

import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	// !tag.equals(hello) AND date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: crunch" ")"]
	tokens := []Token{
		{Type: TokenBang, Literal: ""},
		{Type: TokenIdentifier, Literal: "tag"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "equals"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "hello"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenAnd, Literal: ""},
		{Type: TokenIdentifier, Literal: "due"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "before"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenDate, Literal: "2021-01-01"},
		{Type: TokenRParen, Literal: ""},
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

	parser := NewParser(tokens)

	q, err := parser.Parse()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Query: %s", q)

	expected := &QueryBinaryOp{
		Operator: OperatorAnd,
		Left: &QueryBinaryOp{
			Operator: OperatorAnd,
			Left: &QueryUnaryOp{
				Operator: OperatorNot,
				Operand: &QueryCondition{
					Field:    "tag",
					Value:    "hello",
					Operator: OperatorEq,
				},
			},
			Right: &QueryCondition{
				Field:    "due",
				Value:    "2021-01-01",
				Operator: OperatorLt,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorOr,
			Left: &QueryCondition{
				Field:    "title",
				Value:    "bar",
				Operator: OperatorSw,
			},
			Right: &QueryCondition{
				Field:    "title",
				Value:    "c\"\\runch",
				Operator: OperatorSw,
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}
