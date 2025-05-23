package ntql

import (
	"fmt"
)

type Lexer struct {
	Tokens            []Token
	Scanner           *Scanner
	InnerDepth        int
	ExpectedTokens    []TokenType
	lastTokenVerb     bool
	ExpectedDataTypes []DType
}

type ErrInvalidSubject struct {
	Position int
	Lexeme   Lexeme
}

func (e ErrInvalidSubject) Error() string {
	return fmt.Sprintf("Invalid subject: %s", e.Lexeme)
}

type ErrEndOfInput struct{}

func (e ErrInvalidToken) Error() string {
	// convert expected tokens to string
	var expected string
	for i, t := range e.Expected {
		if i == 0 {
			expected = t.String()
		} else {
			expected = fmt.Sprintf("%s, %s", expected, t.String())
		}
	}

	return fmt.Sprintf("Invalid lexme '%s': expected type from [%s]", e.Lexeme, expected)
}

type ErrInvalidToken struct {
	Expected []TokenType
	Position int
	Lexeme   Lexeme
}

var connectorTypes = []TokenType{TokenAnd, TokenOr}

func (e ErrEndOfInput) Error() string {
	return fmt.Sprintf("End of input")
}

func NewLexer(s string) *Lexer {
	return &Lexer{Tokens: []Token{}, Scanner: NewScanner(s), InnerDepth: 0, ExpectedTokens: []TokenType{TokenLParen, TokenBang, TokenSubject}}
}

func (l *Lexer) insideMethodCall() bool {
	return l.InnerDepth > 0
}

// Lex takes a string and returns a slice of tokens
// Example: tag.equals(hello OR goodbye) OR (date.before(2024-01-08) AND date.after(2024-01-09))
// tag.equals(hello) AND date.before(2021-01-01) AND title.startswith(("bar" OR "c\"\\runch") AND "foo")
func (t *Lexer) Lex() ([]Token, error) {
	for {
		if t.atEnd() {
			break
		}
		err := t.ScanToken()
		if err != nil {
			return nil, err
		}
	}
	return t.Tokens, nil
}

func (t *Lexer) ScanToken() error {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return err
	}
	for _, expected := range t.ExpectedTokens {
		switch expected {
		case TokenBang:
			res, err := t.matchBang(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenLParen:
			res, err := t.matchLParen(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenRParen:
			res, err := t.matchRParen(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDot:
			res, err := t.matchDot(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenAnd:
			res, err := t.matchAnd(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenOr:
			res, err := t.matchOr(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenSubject:
			res, err := t.matchSubject(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenBool:
			res, err := t.matchBool(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDateTime:
			res, err := t.matchDateTime(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDate:
			res, err := t.matchDate(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenString:
			res, err := t.matchString(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenInt:
			res, err := t.matchDigit(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenVerb:
			res, err := t.matchVerb(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenTag:
			res, err := t.matchTag(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		default:
			return ErrInvalidToken{Expected: t.ExpectedTokens, Position: t.Scanner.Pos, Lexeme: lexeme}
		}
	}
	return nil
}

func (t *Lexer) lastToken() (Token, error) {
	if len(t.Tokens) == 0 {
		return Token{}, ErrEndOfInput{}
	}
	return t.Tokens[len(t.Tokens)-1], nil
}

func (t *Lexer) matchSubject(lexeme Lexeme) (bool, error) {
	t.appendToken(TokenSubject, lexeme)
	t.ExpectedTokens = []TokenType{TokenDot}
	subj, err := getSubject(string(lexeme))
	if err != nil {
		t.ExpectedDataTypes = []DType{}
		return false, ErrInvalidSubject{Position: t.Scanner.Pos, Lexeme: lexeme}
	}

	t.ExpectedDataTypes = subj.ValidTypes

	return true, nil
}

func (t *Lexer) matchTag(lexeme Lexeme) (bool, error) {
	t.appendToken(TokenTag, lexeme)
	t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
	return true, nil
}

func (t *Lexer) matchBool(lexeme Lexeme) (bool, error) {
	if lexeme == "true" || lexeme == "false" {
		t.appendToken(TokenBool, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchDot(lexeme Lexeme) (bool, error) {
	if lexeme == "." {
		t.appendToken(TokenDot, lexeme)
		t.ExpectedTokens = []TokenType{TokenVerb}
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchVerb(lexeme Lexeme) (bool, error) {
	t.appendToken(TokenVerb, lexeme)
	t.ExpectedTokens = []TokenType{TokenLParen}
	t.lastTokenVerb = true
	return true, nil
}

func (t *Lexer) matchLParen(lexeme Lexeme) (bool, error) {
	if lexeme == "(" {
		prev, _ := t.lastToken()                         // if there are no tokens, we get an error, but we can ignore it here
		if t.InnerDepth != 0 || prev.Kind == TokenVerb { // if we are in a method
			t.InnerDepth++
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang}
			for _, dtype := range t.ExpectedDataTypes {
				switch dtype {
				case DTypeString:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenString)
				case DTypeInt:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenInt)
				case DTypeDateTime:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDateTime)
				case DTypeDate:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDate)
				case DTypeTag:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenTag)
				}
			}
			t.ExpectedTokens = append(t.ExpectedTokens, TokenBang)
		} else { // keep same fsm state

		}
		t.appendToken(TokenLParen, lexeme)
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchRParen(lexeme Lexeme) (bool, error) {
	if lexeme == ")" {
		t.appendToken(TokenRParen, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		if t.InnerDepth != 0 { // if we are in a method
			t.InnerDepth--
		} else {
		}
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchAnd(lexeme Lexeme) (bool, error) {
	if toLowerCase(string(lexeme)) == "and" {
		t.appendToken(TokenAnd, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang}
			for _, dtype := range t.ExpectedDataTypes {
				switch dtype {
				case DTypeString:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenString)
				case DTypeInt:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenInt)
				case DTypeDateTime:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDateTime)
				case DTypeDate:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDate)
				case DTypeTag:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenTag)
				}
			}
		} else {
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang, TokenSubject}
		}
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchOr(lexeme Lexeme) (bool, error) {
	if toLowerCase(string(lexeme)) == "or" {
		t.appendToken(TokenOr, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang}
			for _, dtype := range t.ExpectedDataTypes {
				switch dtype {
				case DTypeString:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenString)
				case DTypeInt:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenInt)
				case DTypeDateTime:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDateTime)
				case DTypeDate:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenDate)
				case DTypeTag:
					t.ExpectedTokens = append(t.ExpectedTokens, TokenTag)
				}
			}
		} else {
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang, TokenSubject}
		}
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchBang(lexeme Lexeme) (bool, error) {
	if lexeme == "!" {
		t.appendToken(TokenBang, lexeme)
		// keep current fsm state
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchDate(lexeme Lexeme) (bool, error) {
	if dateRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDate, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchDateTime(lexeme Lexeme) (bool, error) {
	if dateTimeRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDateTime, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchString(lexeme Lexeme) (bool, error) {
	if stringRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenString, Lexeme(string(lexeme)[1:len(string(lexeme))-1])) // remove quotes
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Lexer) matchDigit(lexeme Lexeme) (bool, error) {
	if numRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenInt, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Lexer) GetPosition() int {
	return t.Scanner.Pos
}

func (t *Lexer) appendToken(kind TokenType, lexeme Lexeme) {
	t.Tokens = append(t.Tokens, createToken(string(lexeme), kind, t.GetPosition()))
}

func (t *Lexer) atEnd() bool {
	return t.Scanner.atEnd()
}

func (t *Lexer) LastLexeme() (Lexeme, error) {
	return t.Scanner.LastLexeme()
}
