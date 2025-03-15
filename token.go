package tbql

import "fmt"

type TokenType int

const (
	// Single-character tokens
	Bang TokenType = iota + 1
	Dot
	Colon

    // Grouping
	LParen
	RParen

	// Literals
	Identifier
	String
	Number
	Date

	// Operators
	And
	Or
)

func (t TokenType) String() string {
	switch t {
	case Bang:
		return "!"
	case Dot:
		return "."
	case Colon:
		return ":"
	case LParen:
		return "("
	case RParen:
		return ")"
	case Identifier:
		return "Identifier"
	case String:
		return "String"
	case Number:
		return "Number"
	case Date:
		return "Date"
	case And:
		return "AND"
	case Or:
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

