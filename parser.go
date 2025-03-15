package tbql

import "fmt"

type Type int

const (
    TypeString Type = iota + 1
    TypeInt
    TypeDate
)

type Subject struct {
    Name string
    Aliases []string
    ValidVerbs []Verb
    ValidTypes []Type
}

type Verb struct {
    Name string
    Aliases []string
}

var validSubjects = []Subject{
    {
        Name: "name",
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
        Name: "due",
        Aliases: []string{"deadline"},
        ValidVerbs: []Verb{
            {Name: "before", Aliases: []string{}},
            {Name: "after", Aliases: []string{}},
            {Name: "equals", Aliases: []string{}},
        },
        ValidTypes: []Type{TypeDate},
    },
    {
        Name: "status",
        Aliases: []string{"state"},
        ValidVerbs: []Verb{
            {Name: "equals", Aliases: []string{"eq"}},
        },
        ValidTypes: []Type{TypeString},
    },
    {
        Name: "priority",
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
        Name: "project",
        Aliases: []string{},
        ValidVerbs: []Verb{
            {Name: "equals", Aliases: []string{"eq"}},
        },
        ValidTypes: []Type{TypeString},
    },
    {
        Name: "createdAt",
        Aliases: []string{},
        ValidVerbs: []Verb{
            {Name: "before", Aliases: []string{}},
            {Name: "after", Aliases: []string{}},
            {Name: "equals", Aliases: []string{}},
        },
        ValidTypes: []Type{TypeDate},
    },
    {
        Name: "updatedAt",
        Aliases: []string{},
        ValidVerbs: []Verb{
            {Name: "before", Aliases: []string{}},
            {Name: "after", Aliases: []string{}},
            {Name: "equals", Aliases: []string{}},
        },
        ValidTypes: []Type{TypeDate},
    },
    {
        Name: "completedAt",
        Aliases: []string{},
        ValidVerbs: []Verb{
            {Name: "before", Aliases: []string{}},
            {Name: "after", Aliases: []string{}},
            {Name: "equals", Aliases: []string{}},
        },
        ValidTypes: []Type{TypeDate},
    },
    {
        Name: "createdBy",
        Aliases: []string{},
        ValidVerbs: []Verb{
            {Name: "equals", Aliases: []string{"eq"}},
        },
        ValidTypes: []Type{TypeString},
    },
}

func QueryAnd(left QueryExpr, right QueryExpr) QueryExpr {
    return &QueryBinaryOp{Left: left, Right: right, Op: "AND"}
}

func QueryOr(left QueryExpr, right QueryExpr) QueryExpr {
    return &QueryBinaryOp{Left: left, Right: right, Op: "OR"}
}

func QueryNot(expr QueryExpr) QueryExpr {
    return &QueryUnaryOp{Operand: expr, Operator: "NOT"}
}

type BinaryExpression struct {
    Left *BinaryExpression
    Right *BinaryExpression
    Op string
}

type ParserError struct {
    Message string
    Position int
    Code ErrorCode
}

func (e *ParserError) Error() string {
    return fmt.Sprintf("%s error at position %d: %s", e.Code, e.Position, e.Message)
}

func NewParserError(code ErrorCode, position int, message string) *ParserError {
    return &ParserError{
        Code: code,
        Position: position,
        Message: message,
    }
}

type Parser struct {
    Tokens []Token
    Pos    int
}

func NewParser(tokens []Token) *Parser {
    return &Parser{Tokens: tokens, Pos: 0}
}

func (p *Parser) Query() (*Query, error) {
    if (p.match(TokenLParen)) {
        expr, err := p.Expression()
        if err != nil {
            return nil, err
        }

        if !p.match(TokenRParen) {
            return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
        }

        return &expr, nil
    } else {
        expr, err := p.Expression()
        if err != nil {
            return nil, err
        }

        return expr, nil
    }
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

    verb, err := p.Verb()
    if err != nil {
        return nil, err
    }

    if !p.match(TokenLParen) {
        return nil, NewParserError(InvalidInput, p.Pos, "Expected opening parenthesis")
    }

    valueExpr, err := p.Value()
    if err != nil {
        return nil, err
    }

    if !p.match(TokenRParen) {
        return nil, NewParserError(InvalidInput, p.Pos, "Expected closing parenthesis")
    }

    return &QueryFuncCall{Subject: subject, Verb: verb, Value: valueExpr}, nil
}

func transformValueExpr(valueExpr QueryExpr) (string, error) {

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

