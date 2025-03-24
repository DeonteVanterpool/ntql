package tbql

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
    TokenNumber
    TokenBool
    TokenDot
    TokenLParen
    TokenRParen
    TokenAnd
    TokenOr
)

// type TokenType int

/*
const (
	// Single-character tokens
	TokenBang TokenType = iota
	TokenDot
	TokenColon

    // Grouping
	TokenLParen
	TokenRParen

	// Literals
	TokenIdentifier
	TokenString
	TokenNumber
	TokenDate

	// Operators
	TokenAnd
	TokenOr
)
*/

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
	case TokenNumber:
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
	Kind    TokenType
	Literal string
}

func createToken(literal string, token TokenType) Token {
	return Token{Kind: token, Literal: literal}
}

func (t Token) String() string {
	if t.Literal == "" {
		return fmt.Sprintf("\"%s\"", t.Kind)
	} else {
        return fmt.Sprintf("\"%s: %s\"", t.Kind, t.Literal)

	}
}

