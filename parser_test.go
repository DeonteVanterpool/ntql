package tbql

/*
import (
	"encoding/json"
	"testing"
)

func TestParser(t *testing.T) {
	// !tag.equals(hello) AND date.before(2021-01-01) AND title.startswith("bar" OR "c\"\\runch")
	// ["!" "Identifier: tag" "." "Identifier: equals" "(" "Identifier: hello" ")" "AND" "Identifier: date" "." "Identifier: before" "(" "Date: 2021-01-01" "AND" "Identifier: title" "." "Identifier: startswith" "(" "String: bar" "OR" "String: c\"\\runch" ")"]
	tokens := []Token{
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "startswith"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
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
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "startswith"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
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
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "startsWith"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
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
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "starts_with"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenRParen, Literal: ""},
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
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "starts_With"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenString, Literal: "foo"},
		{Kind: TokenRParen, Literal: ""},
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
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "tag"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "equals"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenIdentifier, Literal: "hello"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenOr, Literal: ""},
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenIdentifier, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "before"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2021-01-01"},
		{Kind: TokenRParen, Literal: ""},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenIdentifier, Literal: "title"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenIdentifier, Literal: "startswith"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenBang, Literal: ""},
		{Kind: TokenString, Literal: "bar"},
		{Kind: TokenAnd, Literal: ""},
		{Kind: TokenString, Literal: "c\"\\runch"},
		{Kind: TokenRParen, Literal: ""},
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
        {Kind: TokenBang, Literal: ""},
        {Kind: TokenIdentifier, Literal: "tag"},
        {Kind: TokenDot, Literal: ""},
        {Kind: TokenIdentifier, Literal: "equals"},
        {Kind: TokenLParen, Literal: ""},
        {Kind: TokenIdentifier, Literal: "hello"},
        {Kind: TokenRParen, Literal: ""},
        {Kind: TokenOr, Literal: ""},
        {Kind: TokenBang, Literal: ""},
        {Kind: TokenIdentifier, Literal: "due"},
        {Kind: TokenDot, Literal: ""},
        {Kind: TokenIdentifier, Literal: "before"},
        {Kind: TokenLParen, Literal: ""},
        {Kind: TokenDate, Literal: "2021-01-01"},
        {Kind: TokenRParen, Literal: ""},
        {Kind: TokenAnd, Literal: ""},
        {Kind: TokenIdentifier, Literal: "title"},
        {Kind: TokenDot, Literal: ""},
        {Kind: TokenIdentifier, Literal: "startswith"},
        {Kind: TokenLParen, Literal: ""},
        {Kind: TokenBang, Literal: ""},
        {Kind: TokenLParen, Literal: ""},
        {Kind: TokenString, Literal: "bar"},
        {Kind: TokenAnd, Literal: ""},
        {Kind: TokenString, Literal: "c\"\\runch"},
        {Kind: TokenRParen, Literal: ""},
        {Kind: TokenOr, Literal: ""},
        {Kind: TokenString, Literal: "foo"},
        {Kind: TokenRParen, Literal: ""},
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
*/
