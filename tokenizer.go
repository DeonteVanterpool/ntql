package tbql

import (
	"fmt"
	"strings"
)

type Tokenizer struct {
	Tokens         []Token
	Scanner        *Scanner
	InnerDepth     int
	ExpectedTokens []TokenType
	lastTokenVerb  bool
}

type TokenizationError struct {
	Message  error
	Position int
}

func newTokenizationError(message error, position int) TokenizationError {
	return TokenizationError{Message: message, Position: position}
}

func (e TokenizationError) Error() string {
	return fmt.Sprintf("Error: %s at position %d", e.Message.Error(), e.Position)
}

type ErrEndOfInput struct{}

type ErrInvalidCharacter struct {
	Input byte
}

func (e ErrInvalidCharacter) Error() string {
	return fmt.Sprintf("Invalid character: %s", string(e.Input))
}

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
	return fmt.Sprintf("Invalid type: expected type from [%s]", expected)
}

type ErrInvalidToken struct {
	Expected []TokenType
}

var connectorTypes = []TokenType{TokenAnd, TokenOr}

func (e ErrEndOfInput) Error() string {
	return fmt.Sprintf("End of input")
}

func NewTokenizer(s string) *Tokenizer {
	return &Tokenizer{Tokens: []Token{}, Scanner: NewScanner(s), InnerDepth: 0, ExpectedTokens: []TokenType{TokenLParen, TokenBang, TokenSubject}}
}

// Tokenize takes a string and returns a slice of tokens
// Example: tag.equals(hello OR goodbye) OR (date.before(2024-01-08) AND date.after(2024-01-09))
// tag.equals(hello) AND date.before(2021-01-01) AND title.startswith(("bar" OR "c\"\\runch") AND "foo")
func (t *Tokenizer) Tokenize() ([]Token, error) {
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

func (t *Tokenizer) ScanToken() error {
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
		case TokenTag:
			res, err := t.matchTag(lexeme)
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
		case TokenSubject:
			res, err := t.matchSubject(lexeme)
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
		case TokenDateTime:
			res, err := t.matchDateTime(lexeme)
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
		case TokenNumber:
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
		case TokenDot:
			res, err := t.matchDot(lexeme)
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		default:
			return newTokenizationError(ErrInvalidToken{Expected: t.ExpectedTokens}, t.Scanner.Pos)
		}
	}
	return nil
}

func (t *Tokenizer) LastTokenComplete() (bool, error) {
	lastChar := t.Scanner.S[len(t.Scanner.S)-1]
	symbols := []byte{'!', '(', ')', '.', ' '}

	lastToken, err := t.lastToken()
	if err != nil {
		return false, err
	}

	if lastToken.Kind == TokenString && lastChar != '"' { // incomplete strings should never be complete even if they end with a symbol
		return false, nil
	}

	if strings.Contains(string(symbols), string(lastChar)) {
		return true, nil
	}

	return false, nil
}

func (t *Tokenizer) lastToken() (Token, error) {
	if len(t.Tokens) == 0 {
		return Token{}, fmt.Errorf("No tokens")
	}
	return t.Tokens[len(t.Tokens)-1], nil
}

func (t *Tokenizer) matchSubject(lexeme Lexeme) (bool, error) {
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenSubject, lexeme)
	t.ExpectedTokens = []TokenType{TokenDot}
	return true, nil
}

func (t *Tokenizer) matchTag(lexeme Lexeme) (bool, error) {
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenTag, lexeme)
	t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
	return true, nil
}

func (t *Tokenizer) matchBool(lexeme Lexeme) (bool, error) {
	if lexeme == "true" || lexeme == "false" {
		t.appendToken(TokenBool, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDot(lexeme Lexeme) (bool, error) {
	if lexeme == "." {
		t.appendToken(TokenDot, lexeme)
		t.ExpectedTokens = []TokenType{TokenVerb}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchVerb(lexeme Lexeme) (bool, error) {
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenVerb, lexeme)
	t.ExpectedTokens = []TokenType{TokenLParen}
	t.lastTokenVerb = true
	return true, nil
}

func (t *Tokenizer) matchLParen(lexeme Lexeme) (bool, error) {
	if lexeme == "(" {
		prev, _ := t.lastToken()
		if t.InnerDepth != 0 || prev.Kind == TokenVerb { // if we are in a method
			t.InnerDepth++
			t.ExpectedTokens = append([]TokenType{TokenLParen, TokenBang}, objectTypes...)
			t.ExpectedTokens = append(t.ExpectedTokens, TokenBang)
		} else { // keep same fsm state

		}
		err := t.appendToken(TokenLParen, lexeme)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchRParen(lexeme Lexeme) (bool, error) {
	if lexeme == ")" {
		t.appendToken(TokenRParen, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		if t.InnerDepth != 0 { // if we are not in a method
			t.InnerDepth--
		} else {
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchAnd(lexeme Lexeme) (bool, error) {
	if lexeme == "AND" {
		t.appendToken(TokenAnd, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = append([]TokenType{TokenLParen, TokenBang}, objectTypes...)
		} else {
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang, TokenSubject}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchOr(lexeme Lexeme) (bool, error) {
	if lexeme == "OR" {
		t.appendToken(TokenOr, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = append([]TokenType{TokenLParen, TokenBang}, objectTypes...)
		} else {
			t.ExpectedTokens = []TokenType{TokenLParen, TokenBang, TokenSubject}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchBang(lexeme Lexeme) (bool, error) {
	if lexeme == "!" {
		t.appendToken(TokenBang, lexeme)
		// keep current fsm state
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDate(lexeme Lexeme) (bool, error) {
	if dateRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDate, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDateTime(lexeme Lexeme) (bool, error) {
	if dateTimeRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDateTime, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchString(lexeme Lexeme) (bool, error) {
	if stringRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenString, Lexeme(string(lexeme)[1:len(string(lexeme))-1])) // remove quotes
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDigit(lexeme Lexeme) (bool, error) {
	if numRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenNumber, lexeme)
		t.ExpectedTokens = append(connectorTypes, TokenRParen)
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) GetPosition() int {
	return t.Scanner.Pos
}

func (t *Tokenizer) appendToken(kind TokenType, lexeme Lexeme) error {
	t.Tokens = append(t.Tokens, createToken(string(lexeme), kind))
	return nil
}

func (t *Tokenizer) atEnd() bool {
	return t.Scanner.atEnd()
}

func (t *Tokenizer) Error(e error) error {
	return fmt.Errorf("Error: %s", e.Error())
}

func (t *Tokenizer) LastLexeme() (Lexeme, error) {
	return t.Scanner.LastLexeme()
}
