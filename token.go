package ntql

import "fmt"

type TokenType int

const (
	TokenSubject TokenType = iota
	TokenBang
	TokenVerb
	TokenString
	TokenTag
	TokenDate
	TokenDateTime
	TokenInt
	TokenBool
	TokenDot
	TokenLParen
	TokenRParen
	TokenAnd
	TokenOr
)

// type TokenType int

func (t TokenType) String() string {
	switch t {
	case TokenBang:
		return "!"
	case TokenDot:
		return "."
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenSubject:
		return "Subject"
	case TokenTag:
		return "Tag"
	case TokenVerb:
		return "Verb"
	case TokenBool:
		return "Bool"
	case TokenString:
		return "String"
	case TokenInt:
		return "Number"
	case TokenDate:
		return "Date"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	}
	return "UnnamedTokenType"
}

// Example: tag.equals(hello OR goodbye)
// date.before(2024-01-08) AND date.after(2024-01-09)
type Token struct {
	Kind     TokenType
	Literal  string
	Position int
}

func createToken(literal string, token TokenType, position int) Token {
	return Token{Kind: token, Literal: literal, Position: position}
}

func (t Token) String() string {
	if t.Literal == "" {
		return fmt.Sprintf("\"%s\"", t.Kind)
	} else {
		return fmt.Sprintf("\"%s: %s\"", t.Kind, t.Literal)

	}
}
