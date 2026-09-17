package ntql

import (
	"strings"
)

// Data types for any objects passed as an argument to a function in the NTQL query
type DType int

const (
	DTypeString DType = iota
	DTypeInt
	DTypeDate
	DTypeTag
	DTypeDateTime
)

// TODO: Sanitize input
// BNF Grammar:
// query = expr
// expr = or_expr
// or_expr = and_expr ("OR" and_expr)*
// and_expr = not_expr ("AND" not_expr)*
// not_expr = ["!"] term
// term = func_call | "(" expr ")"
// func_call = subject "." verb "(" value_expr ")" # subject from list of subjects, verb from subject verbs
// value_expr = value_or
// value_or = value_and ("OR" value_and)*
// value_and = value_not ("AND" value_not)*
// value_not = ["!"] value_term
// value_term = "(" value_expr ")" | value
// value = object # type belonging to current verb
// object = NUMBER | STRING | DATE | TAG
type Parser struct {
	Tokens []Token
	Pos    int
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

type Subject struct {
	Name       string
	Aliases    []string
	ValidVerbs []Verb
	ValidTypes []DType
}

type Verb struct {
	Name    string
	Aliases []string
}

var validSubjects = []Subject{
	{
		Name:    "title",
		Aliases: []string{"name"},
		ValidVerbs: []Verb{
			{Name: "startswith", Aliases: []string{}},
			{Name: "endswith", Aliases: []string{}},
			{Name: "contains", Aliases: []string{}},
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []DType{DTypeString},
	},
	{
		Name:    "due",
		Aliases: []string{"deadline"},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []DType{DTypeDate, DTypeDateTime},
	},
	{
		Name:    "status",
		Aliases: []string{"state"},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []DType{DTypeString},
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
		ValidTypes: []DType{DTypeInt},
	},
	{
		Name:    "project",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []DType{DTypeString},
	},
	{
		Name:    "createdAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []DType{DTypeDate, DTypeDateTime},
	},
	{
		Name:    "updatedAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []DType{DTypeDate, DTypeDateTime},
	},
	{
		Name:    "completedAt",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "before", Aliases: []string{}},
			{Name: "after", Aliases: []string{}},
			{Name: "equals", Aliases: []string{}},
		},
		ValidTypes: []DType{DTypeDate, DTypeDateTime},
	},
	{
		Name:    "createdBy",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []DType{DTypeString},
	},
	{
		Name:    "tag",
		Aliases: []string{},
		ValidVerbs: []Verb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []DType{DTypeTag},
	},
}

func (p *Parser) Parse() (QueryExpr, error) {
	return p.Query()
}

func NewParserError(message string, t Token) *ParserError {
	return &ParserError{
		Message: message,
		Token:   t,
	}
}

type ValueExpr interface {
	Transform(subject string, verb string) (QueryExpr, error)
}

func (v *Value) Transform(subject string, verb string) (QueryExpr, error) {
	return NewQueryCondition(subject, verb, v.Value)
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
	return &QueryBinaryOp{Left: left, Right: right, Operator: v.Operator}, nil
}

func (v *ValueUnaryOp) Transform(subject string, verb string) (QueryExpr, error) {
	operand, err := v.Operand.Transform(subject, verb)
	if err != nil {
		return nil, err
	}
	return &QueryUnaryOp{Operand: operand, Operator: v.Operator}, nil
}

func (p *Parser) match(t TokenType) bool {
	if p.Pos >= len(p.Tokens) {
		return false
	}
	if p.Tokens[p.Pos].Kind != t {
		return false
	}
	p.advance()
	return true
}

func (p *Parser) advance() {
	p.Pos++
}

func (p *Parser) previous() Token {
	return p.Tokens[p.Pos-1]
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
		expr = NewQueryOr(expr, right)
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
		expr = NewQueryAnd(expr, right)
	}
	return expr, nil
}

func (p *Parser) NotExpr() (QueryExpr, error) {
	if p.match(TokenBang) {
		expr, err := p.Term()
		if err != nil {
			return nil, err
		}
		return NewQueryNot(expr), nil
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
			return nil, NewParserError("Expected closing parenthesis", p.previous())
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
		return nil, NewParserError("Expected dot", p.previous())
	}

	verb, err := p.Verb(subject)
	if err != nil {
		return nil, err
	}

	if !p.match(TokenLParen) {
		return nil, NewParserError("Expected opening parenthesis", p.previous())
	}

	valueExpr, err := p.ValueExpr()
	if err != nil {
		return nil, err
	}

	if !p.match(TokenRParen) {
		return nil, NewParserError("Expected closing parenthesis", p.previous())
	}

	return valueExpr.Transform(subject, verb)
}

func toLowerCase(s string) string {
	if s == "" {
		return s
	}
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, "_", "")
}

func (p *Parser) Subject() (string, error) {
	if p.match(TokenSubject) {
		subject := p.previous().Literal
		s, err := getSubject(subject)
		if err != nil {
			return "", NewParserError("Invalid subject: "+subject, p.previous())
		}
		return s.Name, nil
	} else {
		return "", NewParserError("Expected subject", p.previous())
	}
}

func (p *Parser) Verb(subject string) (string, error) {
	for _, s := range validSubjects {
		if s.Name == subject {
			if p.match(TokenVerb) {
				verb := p.previous().Literal
				for _, v := range s.ValidVerbs {
					if toLowerCase(v.Name) == toLowerCase(verb) {
						return v.Name, nil
					}
					for _, alias := range v.Aliases {
						if toLowerCase(alias) == toLowerCase(verb) {
							return v.Name, nil
						}
					}
				}
				return "", NewParserError("Invalid verb: "+verb, p.previous())
			} else {
				return "", NewParserError("Expected verb", p.previous())
			}
		}
	}
	return "", NewParserError("Invalid subject: "+subject, p.previous())
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
	if p.match(TokenBang) {
		valueTerm, err := p.ValueTerm()
		if err != nil {
			return nil, err
		}
		return &ValueUnaryOp{Operator: "NOT", Operand: valueTerm}, nil
	} else {
		return p.ValueTerm()
	}
}

func (p *Parser) ValueTerm() (ValueExpr, error) {
	if p.match(TokenLParen) {
		valueExpr, err := p.ValueExpr()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRParen) {
			return nil, NewParserError("Expected closing parenthesis", p.Tokens[p.Pos])
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
	if p.match(TokenString) || p.match(TokenDate) || p.match(TokenTag) || p.match(TokenInt) {
		return &Value{Value: p.previous().Literal}, nil
	} else {
		return nil, NewParserError("Expected value. Got: "+p.Tokens[p.Pos].Literal, p.Tokens[p.Pos])
	}
}
