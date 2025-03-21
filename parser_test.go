package tbql

import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	// !tag.equals(hello) AND date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: c\"\\runch" ")"]
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
				Operator: OperatorLT,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorOr,
			Left: &QueryCondition{
				Field:    "title",
				Value:    "bar",
				Operator: OperatorSW,
			},
			Right: &QueryCondition{
				Field:    "title",
				Value:    "c\"\\runch",
				Operator: OperatorSW,
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}

func TestParserOpPrecedenceAnd(t *testing.T) {
	// !tag.equals(hello) OR date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "OR" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: c\"\\runch" ")"]
	tokens := []Token{
		{Type: TokenBang, Literal: ""},
		{Type: TokenIdentifier, Literal: "tag"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "equals"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "hello"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenOr, Literal: ""},
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
			Operator: OperatorOr,
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
				Operator: OperatorLT,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorOr,
			Left: &QueryCondition{
				Field:    "title",
				Value:    "bar",
				Operator: OperatorSW,
			},
			Right: &QueryCondition{
				Field:    "title",
				Value:    "c\"\\runch",
				Operator: OperatorSW,
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}

func TestParserOpPrecedenceOr(t *testing.T) {
	// !tag.equals(hello) AND date.before(2021-01-01) OR title.startswith("bar" OR "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "OR" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: c\"\\runch" ")"]
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
		{Type: TokenOr, Literal: ""},
		{Type: TokenIdentifier, Literal: "title"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "startsWith"},
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
		Operator: OperatorOr,
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
				Operator: OperatorLT,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorOr,
			Left: &QueryCondition{
				Field:    "title",
				Value:    "bar",
				Operator: OperatorSW,
			},
			Right: &QueryCondition{
				Field:    "title",
				Value:    "c\"\\runch",
				Operator: OperatorSW,
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}

func TestOuterNestedParens(t *testing.T) {
	// !tag.equals(hello) OR (date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch"))
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "OR" "(" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: c\"\\runch" ")" ")"]
	tokens := []Token{
		{Type: TokenBang, Literal: ""},
		{Type: TokenIdentifier, Literal: "tag"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "equals"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "hello"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenOr, Literal: ""},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "due"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "before"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenDate, Literal: "2021-01-01"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenAnd, Literal: ""},
		{Type: TokenIdentifier, Literal: "title"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "starts_with"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenString, Literal: "bar"},
		{Type: TokenOr, Literal: ""},
		{Type: TokenString, Literal: "c\"\\runch"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenRParen, Literal: ""},
	}

	parser := NewParser(tokens)

	q, err := parser.Parse()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Query: %s", q)

	expected := &QueryBinaryOp{
		Operator: OperatorOr,
		Left: &QueryUnaryOp{
			Operator: OperatorNot,
			Operand: &QueryCondition{
				Field:    "tag",
				Value:    "hello",
				Operator: OperatorEq,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorAnd,
			Left: &QueryCondition{
				Field:    "due",
				Value:    "2021-01-01",
				Operator: OperatorLT,
			},
			Right: &QueryBinaryOp{
				Operator: OperatorOr,
				Left: &QueryCondition{
					Field:    "title",
					Value:    "bar",
					Operator: OperatorSW,
				},
				Right: &QueryCondition{
					Field:    "title",
					Value:    "c\"\\runch",
					Operator: OperatorSW,
				},
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}

func TestInnerNestedParens(t *testing.T) {
	// !tag.equals(hello) OR date.before(2021-01-01) AND title.startswith(("bar" OR "c\"\\runch") AND "foo")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "OR" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "AND" "Identifier: title" "." "Identifier: startswith" "(" "(" "String: bar" "OR" "String: c\"\\runch" ")" "AND" "String: foo" ")"]

	tokens := []Token{
		{Type: TokenBang, Literal: ""},
		{Type: TokenIdentifier, Literal: "tag"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "equals"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "hello"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenOr, Literal: ""},
		{Type: TokenIdentifier, Literal: "due"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "before"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenDate, Literal: "2021-01-01"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenAnd, Literal: ""},
		{Type: TokenIdentifier, Literal: "title"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "starts_With"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenString, Literal: "bar"},
		{Type: TokenOr, Literal: ""},
		{Type: TokenString, Literal: "c\"\\runch"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenAnd, Literal: ""},
		{Type: TokenString, Literal: "foo"},
		{Type: TokenRParen, Literal: ""},
	}

	parser := NewParser(tokens)

	q, err := parser.Parse()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Query: %s", q)

	expected := &QueryBinaryOp{
		Operator: OperatorOr,
		Left: &QueryUnaryOp{
			Operator: OperatorNot,
			Operand: &QueryCondition{
				Field:    "tag",
				Value:    "hello",
				Operator: OperatorEq,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorAnd,
			Left: &QueryCondition{
				Field:    "due",
				Value:    "2021-01-01",
				Operator: OperatorLT,
			},
			Right: &QueryBinaryOp{
				Operator: OperatorAnd,
				Left: &QueryBinaryOp{
					Operator: OperatorOr,
					Left: &QueryCondition{
						Field:    "title",
						Value:    "bar",
						Operator: OperatorSW,
					},
					Right: &QueryCondition{
						Field:    "title",
						Value:    "c\"\\runch",
						Operator: OperatorSW,
					},
				},
				Right: &QueryCondition{
					Field:    "title",
					Value:    "foo",
					Operator: OperatorSW,
				},
			},
		},
	}

	if q.String() != expected.String() {
		expectedStr, _ := json.MarshalIndent(expected, "", "    ")
		qStr, _ := json.MarshalIndent(q, "", "    ")
		t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
	}
}

func TestNestedNot(t *testing.T) {
	// !tag.equals(hello) OR !date.before(2021-01-01) AND title.startswith(!"bar" AND "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "OR" "!" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "AND" "Identifier: title" "." "Identifier: startswith" "(" "!" "String: bar" "AND" "String: c\"\\runch" ")"]
	tokens := []Token{
		{Type: TokenBang, Literal: ""},
		{Type: TokenIdentifier, Literal: "tag"},
		{Type: TokenDot, Literal: ""},
		{Type: TokenIdentifier, Literal: "equals"},
		{Type: TokenLParen, Literal: ""},
		{Type: TokenIdentifier, Literal: "hello"},
		{Type: TokenRParen, Literal: ""},
		{Type: TokenOr, Literal: ""},
		{Type: TokenBang, Literal: ""},
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
		{Type: TokenBang, Literal: ""},
		{Type: TokenString, Literal: "bar"},
		{Type: TokenAnd, Literal: ""},
		{Type: TokenString, Literal: "c\"\\runch"},
		{Type: TokenRParen, Literal: ""},
	}

	p := NewParser(tokens)

	q, err := p.Parse()

	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Query: %s", q)

	expected := &QueryBinaryOp{
		Operator: OperatorOr,
		Left: &QueryUnaryOp{
			Operator: OperatorNot,
			Operand: &QueryCondition{
				Field:    "tag",
				Value:    "hello",
				Operator: OperatorEq,
			},
		},
		Right: &QueryBinaryOp{
			Operator: OperatorAnd,
			Left: &QueryUnaryOp{
				Operator: OperatorNot,
				Operand: &QueryCondition{
					Field:    "due",
					Value:    "2021-01-01",
					Operator: OperatorLT,
				},
			},
			Right: &QueryBinaryOp{
				Operator: OperatorAnd,
				Left: &QueryUnaryOp{
					Operator: OperatorNot,
					Operand: &QueryCondition{
						Field:    "title",
						Value:    "bar",
						Operator: OperatorSW,
					},
				},
				Right: &QueryCondition{
					Field:    "title",
					Value:    "c\"\\runch",
					Operator: OperatorSW,
				},
			},
		},
	}

    if q.String() != expected.String() {
        expectedStr, _ := json.MarshalIndent(expected, "", "    ")
        qStr, _ := json.MarshalIndent(q, "", "    ")
        t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
    }
}

func TestNestedNot2(t *testing.T) {
    // !tag.equals(hello) OR !date.before(2021-01-01) AND title.startswith(!("bar" AND "c\"\\runch") OR "foo")
    // ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "OR" "!" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" ")" "AND" "Identifier: title" "." "Identifier: startswith" "(" "!" "(" "String: bar" "AND" "String: c\"\\runch" ")" "OR" "String: foo" ")"]
    tokens := []Token{
        {Type: TokenBang, Literal: ""},
        {Type: TokenIdentifier, Literal: "tag"},
        {Type: TokenDot, Literal: ""},
        {Type: TokenIdentifier, Literal: "equals"},
        {Type: TokenLParen, Literal: ""},
        {Type: TokenIdentifier, Literal: "hello"},
        {Type: TokenRParen, Literal: ""},
        {Type: TokenOr, Literal: ""},
        {Type: TokenBang, Literal: ""},
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
        {Type: TokenBang, Literal: ""},
        {Type: TokenLParen, Literal: ""},
        {Type: TokenString, Literal: "bar"},
        {Type: TokenAnd, Literal: ""},
        {Type: TokenString, Literal: "c\"\\runch"},
        {Type: TokenRParen, Literal: ""},
        {Type: TokenOr, Literal: ""},
        {Type: TokenString, Literal: "foo"},
        {Type: TokenRParen, Literal: ""},
    }

    p := NewParser(tokens)

    q, err := p.Parse()

    if err != nil {
        t.Errorf("Error: %s", err.Error())
    }

    t.Logf("Query: %s", q)

    expected := &QueryBinaryOp{
        Operator: OperatorOr,
        Left: &QueryUnaryOp{
            Operator: OperatorNot,
            Operand: &QueryCondition{
                Field:    "tag",
                Value:    "hello",
                Operator: OperatorEq,
            },
        },
        Right: &QueryBinaryOp{
            Operator: OperatorAnd,
            Left: &QueryUnaryOp{
                Operator: OperatorNot,
                Operand: &QueryCondition{
                    Field:    "due",
                    Value:    "2021-01-01",
                    Operator: OperatorLT,
                },
            },
            Right: &QueryBinaryOp{
                Operator: OperatorOr,
                Left: &QueryUnaryOp{
                    Operator: OperatorNot,
                    Operand: &QueryBinaryOp{
                        Operator: OperatorAnd,
                        Left: &QueryCondition{
                            Field:    "title",
                            Value:    "bar",
                            Operator: OperatorSW,
                        },
                        Right: &QueryCondition{
                            Field:    "title",
                            Value:    "c\"\\runch",
                            Operator: OperatorSW,
                        },
                    },
                },
                Right: &QueryCondition{
                    Field:    "title",
                    Value:    "foo",
                    Operator: OperatorSW,
                },
            },
        },
    }
    
    if q.String() != expected.String() {
        expectedStr, _ := json.MarshalIndent(expected, "", "    ")
        qStr, _ := json.MarshalIndent(q, "", "    ")
        t.Errorf("Expected: %s\n Got: %s", expectedStr, qStr)
    }
}
