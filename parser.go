package tbql

import "fmt"

type Type int

const (
	TypeString Type = iota
	TypeInt
	TypeDate
)

type Subject struct {
	Name       string
	Aliases    []string
	ValidVerbs []Verb
	ValidTypes []Type
}

type Verb struct {
	Name    string
	Aliases []string
}

var validSubjects = []Subject{
	{
		Name:    "name",
		Aliases: []string{"title"},
		ValidVerbs: []Verb{
			{Name: "startswith", Aliases: []string{}},
			{Name: "endswith", Aliases: []string{}},
			{Name: "contains", Aliases: []string{}},
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []Type{TypeString},
	},
	{
		Name:    "due",
		Aliases: []string{"deadline"},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []Type{TypeDate},
	},
	{
		Name:    "status",
		Aliases: []string{"state"},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []Type{TypeString},
	},
	{
		Name:    "priority",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
			{Name: "lessthan", Aliases: []string{"lt"}},
			{Name: "greaterthan", Aliases: []string{"gt"}},
			{Name: "lessthanorequal", Aliases: []string{"lte"}},
			{Name: "greaterthanorequal", Aliases: []string{"gte"}},
		},
		ValidTypes: []Type{TypeInt},
	},
	{
		Name:    "project",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []Type{TypeString},
	},
	{
		Name:    "createdAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []Type{TypeDate},
	},
	{
		Name:    "updatedAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []Type{TypeDate},
	},
	{
		Name:    "completedAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []Type{TypeDate},
	},
	{
		Name:    "createdBy",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []Type{TypeString},
	},
}

func QueryAnd(left QueryExpr, right QueryExpr) *QueryBinaryOp {
	return &QueryBinaryOp{Left: left, Right: right, Op: "AND"}
}

func QueryOr(left QueryExpr, right QueryExpr) *QueryBinaryOp {
	return &QueryBinaryOp{Left: left, Right: right, Op: "OR"}
}

func QueryNot(expr QueryExpr) *QueryUnaryOp {
	return &QueryUnaryOp{Operand: expr, Operator: "NOT"}
}

func QueryFuncCall(subject string, verb string, value string) (*QueryCondition, error) {
	op, err := NewOperator(verb)
	if err != nil {
		return nil, err
	}
	return &QueryCondition{Field: subject, Operator: op, Value: value}, nil
}

type BinaryExpression struct {
	Left  *BinaryExpression
	Right *BinaryExpression
	Op    string
}

type ParserError struct {
	Message  string
	Position int
	Code     ErrorCode
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("%s error at position %d: %s", e.Code, e.Position, e.Message)
}

func NewParserError(code ErrorCode, position int, message string) *ParserError {
	return &ParserError{
		Code:     code,
		Position: position,
		Message:  message,
	}
}

type Parser struct {
	Tokens []Token
	Pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{Tokens: tokens, Pos: 0}
}

func (p *Parser) Query() (QueryExpr, error) {
	return p.Expression()
}

func (p *Parser) Expression() (QueryExpr, error) {
	return p.OrExpr()
}

func (p *Parser) OrExpr() (QueryExpr, error) {
	expr, err := p.AndExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenOr) {
		right, err := p.AndExpr()
		if err != nil {
			return nil, err
		}
		expr = QueryOr(expr, right)
	}
	return expr, nil
}

func (p *Parser) AndExpr() (QueryExpr, error) {
	expr, err := p.NotExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenAnd) {
		right, err := p.NotExpr()
		if err != nil {
			return nil, err
		}
		expr = QueryAnd(expr, right)
	}
	return expr, nil
}

func (p *Parser) NotExpr() (QueryExpr, error) {
	if p.match(TokenBang) {
		expr, err := p.Term()
		if err != nil {
			return nil, err
		}
		return QueryNot(expr), nil
	}
	return p.Term()
}

func (p *Parser) Term() (QueryExpr, error) {
	if p.match(TokenLParen) {
		expr, err := p.Expression()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRParen) {
			return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
		}
		return expr, nil
	} else {
		funcCall, err := p.FunctionCall()
		if err != nil {
			return nil, err
		}

		return funcCall, nil
	}
}

func (p *Parser) FunctionCall() (QueryExpr, error) {
	subject, err := p.Subject()
	if err != nil {
		return nil, err

	}

	if !p.match(TokenDot) {
		return nil, NewParserError(InvalidInput, p.Pos, "Expected dot")
	}

	verb, err := p.Verb(subject)
	if err != nil {
		return nil, err
	}

	if !p.match(TokenLParen) {
		return nil, NewParserError(InvalidInput, p.Pos, "Expected opening parenthesis")
	}

	valueExpr, err := p.ValueExpr()
	if err != nil {
		return nil, err
	}

	if !p.match(TokenRParen) {
		return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
	}

	return valueExpr.Transform(subject, verb)
}

func (p *Parser) Subject() (string, error) {
    if p.match(TokenIdentifier) {
        subject := p.previous().Literal
        for _, s := range validSubjects {
            if s.Name == subject {
                return subject, nil
            }
            for _, alias := range s.Aliases {
                if alias == subject {
                    return s.Name, nil
                }
            }
        }
        return "", NewParserError(InvalidInput, p.Pos, "Invalid subject")
    } else {
        return "", NewParserError(InvalidInput, p.Pos, "Expected subject")
    }
}

func (p *Parser) Verb(subject string) (string, error) {
    for _, s := range validSubjects {
        if s.Name == subject {
            if p.match(TokenIdentifier) {
                verb := p.previous().Literal
                for _, v := range s.ValidVerbs {
                    if v.Name == verb {
                        return verb, nil
                    }
                    for _, alias := range v.Aliases {
                        if alias == verb {
                            return v.Name, nil
                        }
                    }
                }
                return "", NewParserError(InvalidInput, p.Pos, "Invalid verb")
            } else {
                return "", NewParserError(InvalidInput, p.Pos, "Expected verb")
            }
        }
    }
    return "", NewParserError(InvalidInput, p.Pos, "Invalid subject")
}

type ValueExpr interface {
	Transform(subject string, verb string) (QueryExpr, error)
}

type ValueBinaryOp struct {
	Operator Operator
	Left     ValueExpr
	Right    ValueExpr
}

type ValueUnaryOp struct {
	Operator Operator
	Operand  ValueExpr
}

type Value struct {
	Value string
}

func (v *Value) Transform(subject string, verb string) (QueryExpr, error) {
	return QueryFuncCall(subject, verb, v.Value)
}

func (v *ValueBinaryOp) Transform(subject string, verb string) (QueryExpr, error) {
	left, err := v.Left.Transform(subject, verb)
	if err != nil {
		return nil, err
	}
	right, err := v.Right.Transform(subject, verb)
	if err != nil {
		return nil, err
	}
	return &QueryBinaryOp{Left: left, Right: right, Op: v.Operator}, nil
}

func (v *ValueUnaryOp) Transform(subject string, verb string) (QueryExpr, error) {
	operand, err := v.Operand.Transform(subject, verb)
	if err != nil {
		return nil, err
	}
	return &QueryUnaryOp{Operand: operand, Operator: v.Operator}, nil
}

func (p *Parser) ValueExpr() (ValueExpr, error) {
	valueOrExpr, err := p.ValueOr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenOr) {
		right, err := p.ValueOr()
		if err != nil {
			return nil, err
		}
		valueOrExpr = &ValueBinaryOp{Operator: "OR", Left: valueOrExpr, Right: right}
	}
	return valueOrExpr, nil
}

func (p *Parser) ValueOr() (ValueExpr, error) {
	valueAndExpr, err := p.ValueAnd()
	if err != nil {
		return nil, err
	}

	for p.match(TokenAnd) {
		right, err := p.ValueAnd()
		if err != nil {
			return nil, err
		}
		valueAndExpr = &ValueBinaryOp{Operator: "AND", Left: valueAndExpr, Right: right}
	}
	return valueAndExpr, nil
}

func (p *Parser) ValueAnd() (ValueExpr, error) {
	valueNotExpr, err := p.ValueNot()
	if err != nil {
		return nil, err
	}

	if p.match(TokenBang) {
		return &ValueUnaryOp{Operator: "NOT", Operand: valueNotExpr}, nil
	}
	return valueNotExpr, nil
}

func (p *Parser) ValueNot() (ValueExpr, error) {
	if p.match(TokenLParen) {
		valueExpr, err := p.ValueExpr()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRParen) {
			return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
		}
		return valueExpr, nil
	} else {
		value, err := p.ValueTerm()
		if err != nil {
			return nil, err
		}
		return value, nil
	}
}

func (p *Parser) ValueTerm() (ValueExpr, error) {
    if p.match(TokenLParen) {
        valueExpr, err := p.ValueExpr()
        if err != nil {
            return nil, err
        }
        if !p.match(TokenRParen) {
            return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
        }
        return valueExpr, nil
    } else {
        value, err := p.ValueObject()
        if err != nil {
            return nil, err
        }
        return value, nil
    }
}

func (p *Parser) ValueObject() (ValueExpr, error) {
	if p.match(TokenString) || p.match(TokenDate) || p.match(TokenIdentifier) || p.match(TokenNumber) {
		return &Value{Value: p.previous().Literal}, nil
	} else {
		return nil, NewParserError(InvalidInput, p.Pos, "Expected term")
	}
}

func (p *Parser) match(t TokenType) bool {
	if p.Pos >= len(p.Tokens) {
		return false
	}
	if p.Tokens[p.Pos].Type != t {
		return false
	}
	p.advance()
	return true
}

func (p *Parser) advance() {
	p.Pos++
}

func (p *Parser) peek() Token {
	if p.Pos >= len(p.Tokens) {
		return Token{}
	}
	return p.Tokens[p.Pos]
}

func (p *Parser) isAtEnd() bool {
	return p.Pos >= len(p.Tokens)
}

func (p *Parser) previous() Token {
	return p.Tokens[p.Pos-1]
}
