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
	return fmt.Sprintf("Error: %s at position %d", e.Message, e.Position)
}

type ErrEndOfInput struct {
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

func (e ErrEndOfInput) Error() string {
    return fmt.Sprintf("End of input")
}

func NewTokenizer(s string) *Tokenizer {
	return &Tokenizer{Tokens: []Token{}, Scanner: NewScanner(s), InnerDepth: 0, ExpectedTokens: []TokenType{TokenSubject, TokenLParen}}
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
	for _, expected := range t.ExpectedTokens {
		switch expected {
		case TokenTag:
			res, err := t.matchTag()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenBool:
			res, err := t.matchBool()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenSubject:
			res, err := t.matchSubject()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenLParen:
			res, err := t.matchLParen()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenRParen:
			res, err := t.matchRParen()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenAnd:
			res, err := t.matchAnd()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenOr:
			res, err := t.matchOr()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenBang:
			res, err := t.matchBang()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDate:
			res, err := t.matchDate()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDateTime:
			res, err := t.matchDateTime()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenString:
			res, err := t.matchString()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenNumber:
			res, err := t.matchDigit()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenVerb:
			res, err := t.matchVerb()
			if err != nil {
				return err
			}
			if res {
				return nil
			}
		case TokenDot:
			res, err := t.matchDot()
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
	return newTokenizationError(ErrEndOfInput{}, t.Scanner.Pos)
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

func (t *Tokenizer) matchSubject() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenSubject, lexeme)
	t.ExpectedTokens = []TokenType{TokenDot}
	return true, nil
}

func (t *Tokenizer) matchTag() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenTag, lexeme)
	t.ExpectedTokens = []TokenType{TokenDot}
	return true, nil
}

func (t *Tokenizer) matchBool() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "true" || lexeme == "false" {
		t.appendToken(TokenBool, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDot() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "." {
		t.appendToken(TokenDot, lexeme)
		t.ExpectedTokens = []TokenType{TokenVerb}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchVerb() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "" {
		return false, nil
	}
	t.appendToken(TokenVerb, lexeme)
	t.ExpectedTokens = []TokenType{TokenLParen}
	t.lastTokenVerb = true
	return true, nil
}

func (t *Tokenizer) matchLParen() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "(" {
		t.appendToken(TokenLParen, lexeme)
		if err != nil {
			return false, err
		}
		if t.InnerDepth != 0 || t.lastTokenVerb { // if we are in a method
			t.InnerDepth++
			t.ExpectedTokens = objectTypes
			t.lastTokenVerb = false
		} else {
			t.ExpectedTokens = []TokenType{TokenSubject, TokenLParen}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchRParen() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == ")" {
		t.appendToken(TokenRParen, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = []TokenType{TokenAnd, TokenOr}
			t.InnerDepth--
		} else {
			t.ExpectedTokens = []TokenType{TokenAnd, TokenOr}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchAnd() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "AND" {
		t.appendToken(TokenAnd, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = append(objectTypes, TokenLParen)
		} else {
			t.ExpectedTokens = []TokenType{TokenSubject, TokenLParen}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchOr() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "OR" {
		t.appendToken(TokenOr, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = append(objectTypes, TokenLParen)
		} else {
			t.ExpectedTokens = []TokenType{TokenSubject, TokenLParen}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchBang() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if lexeme == "!" {
		t.appendToken(TokenBang, lexeme)
		if t.InnerDepth != 0 { // if we are in a method
			t.ExpectedTokens = append(objectTypes, TokenLParen)
		} else {
			t.ExpectedTokens = []TokenType{TokenSubject, TokenLParen}
		}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDate() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if dateRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDate, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDateTime() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if dateTimeRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenDateTime, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchString() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if stringRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenString, Lexeme(string(lexeme)[1:len(string(lexeme))-1])) // remove quotes
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
}

func (t *Tokenizer) matchDigit() (bool, error) {
	lexeme, err := t.Scanner.ScanLexeme()
	if err != nil {
		return false, err
	}
	if numRegexp.MatchString(string(lexeme)) {
		t.appendToken(TokenNumber, lexeme)
		t.ExpectedTokens = []TokenType{TokenAnd, TokenOr, TokenRParen}
		return true, nil
	}
	return false, nil
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
