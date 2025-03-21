package tbql

import "fmt"

type TokenType int

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

func (t TokenType) String() string {
	switch t {
	case TokenBang:
		return "!"
	case TokenDot:
		return "."
	case TokenColon:
		return ":"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenIdentifier:
		return "Identifier"
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
	Type    TokenType
	Literal string
}

func createToken(literal string, token TokenType) Token {
	return Token{Type: token, Literal: literal}
}

func (t Token) String() string {
	if t.Literal == "" {
		return fmt.Sprintf("\"%s\"", t.Type)
	} else {
        return fmt.Sprintf("\"%s: %s\"", t.Type, t.Literal)

	}
}

