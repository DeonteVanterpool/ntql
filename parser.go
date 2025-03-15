package tbql

import "fmt"

var validSubjects = map[string][]string {
    "name": {"startswith", "endswith", "contains", "equals"},
    "due": {"before", "after", "equals"},
    "status": {"equals"},
    "priority": {"equals", "lessthan", "greaterthan", "lessthanorequal", "greaterthanorequal"},
    "project": {"equals"},
    "createdAt": {"before", "after", "equals"},
    "updatedAt": {"before", "after", "equals"},
    "completedAt": {"before", "after", "equals"},
    "createdBy": {"equals"},
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
    if (p.match(LParen)) {
        expr, err := p.Expression()
        if err != nil {
            return nil, err
        }

        if !p.match(RParen) {
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
    
    for p.match(Or) {
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

    for p.match(And) {
        right, err := p.NotExpr()
        if err != nil {
            return nil, err
        }
        expr = QueryAnd(expr, right)
    }
    return expr, nil
}

func (p *Parser) NotExpr() (QueryExpr, error) {
    if p.match(Bang) {
        expr, err := p.Term()
        if err != nil {
            return nil, err
        }
        return QueryNot(expr), nil
    }
    return p.Term()
}

func (p *Parser) Term() (QueryExpr, error) {
    if p.match(LParen) {
        expr, err := p.Expression()
        if err != nil {
            return nil, err
        }
        if !p.match(RParen) {
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

    if !p.match(Dot) {
        return nil, NewParserError(InvalidInput, p.Pos, "Expected dot")
    }

    verb, err := p.Verb()
    if err != nil {
        return nil, err
    }

    if !p.match(LParen) {
        return nil, NewParserError(InvalidInput, p.Pos, "Expected opening parenthesis")
    }

    valueExpr, err := p.Value()
    if err != nil {
        return nil, err
    }

    if !p.match(RParen) {
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


