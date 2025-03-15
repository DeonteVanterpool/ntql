package tbql

import (
	"fmt"
)

type Tokenizer struct {
	Tokens []Token
	Pos    int
	S      string
}

func NewTokenizer(s string) *Tokenizer {
	return &Tokenizer{Tokens: []Token{}, S: s, Pos: 0}
}

// Tokenize takes a string and returns a slice of tokens
// Example: tag.equals(hello OR goodbye) OR (date.before(2024-01-08) AND date.after(2024-01-09))
// tag.equals(hello) AND date.before(2021-01-01) AND title.startswith(("bar" OR "c\"\\runch") AND "foo")
func (l *Tokenizer) Tokenize() ([]Token, error) {
	for {
		if l.atEnd() {
			break
		}
		err := l.ScanToken()
		if err != nil {
			return nil, err
		}
	}
	return l.Tokens, nil
}

func (l *Tokenizer) ScanToken() error {
	c, err := l.advance()
	if err != nil {
		return err
	}
	switch c {
	case '!':
		return l.appendToken(Bang)
	case '(':
		return l.appendToken(LParen)
	case ')':
		return l.appendToken(RParen)
	case '.':
		return l.appendToken(Dot)
	case ':':
		return l.appendToken(Colon)
	case '"':
		return l.Quote()
	case ' ':
		return l.ScanToken()
	}

	if isAlpha(c) {
		return l.Identifier()
	} else if isNum(c) {
		return l.Digit()
	} else {
		return NewTokenizationError(InvalidInput, l.Pos, fmt.Sprintf("Invalid character: %s", string(c)))
	}
}

func (l *Tokenizer) Digit() error {
	d, err := l.current()
	if err != nil {
		return err
	}

	num := string(d)

	isDate := false
	d, err = l.advance()
	if err != nil {
		return err
	}
	for {
		if isNum(d) {
			num += string(d)
		} else if isHyphen(d) {
			isDate = true
			num += string(d)
		} else {
			break
		}

		d, err = l.advance()
		if err != nil {
			return err
		}
	}

	if isDate {
		if !dateRegexp.MatchString(num) {
			return NewTokenizationError(InvalidInput, l.Pos, fmt.Sprintf("Invalid date: %s", num))
		}
		l.appendLiteral(Date, num)
	} else {
		l.appendLiteral(Number, num)
	}
	return nil
}

func (l *Tokenizer) Quote() error {
	s := "" // look at current character

	for {
		c, err := l.advance() // skip over first quote
		if err != nil {
			return err
		}
		matched, err := l.match('"')
		if matched {
			break
		}

		if c == '\\' {

			c, err = l.advance() // skip over escape character
			if err != nil {
				return err
			}

			// NOTE: we may modify for custom behaviour here to allow for more escape characters
			if c == '"' {
			} else if c == '\\' {
			} else {
				return NewTokenizationError(InvalidInput, l.Pos, fmt.Sprintf("Invalid escape character: %s", string(c)))
			}
		}
		s += string(c)
	}

	l.appendLiteral(String, s)

	return nil
}

func (l *Tokenizer) Identifier() error {
	c, err := l.current()
	if err != nil {
		return err
	}
	id := string(c)
	c, err = l.peek()
	if err != nil {
		return err
	}
	for {
		if isAlphaNum(c) || isHyphen(c) || c == '_' {
			id += string(c)
		} else {
			break
		}
		c, err = l.advance()
		if err != nil {
			return err
		}
		c, err = l.peek()
		if err != nil {
			return err
		}
	}

	if keyword(id) {
		if id == "AND" {
			l.appendToken(And)
		} else if id == "OR" {
			l.appendToken(Or)

		} else if id == "NOT" {
			l.appendToken(Bang)
		} else {
			return fmt.Errorf("Unimplemented keyword: %s", id)
		}
	} else {
		l.appendLiteral(Identifier, id)
	}

	return nil
}

func (l *Tokenizer) match(c byte) (bool, error) {
	v, e := l.current()
	return v == c, e
}

func (l *Tokenizer) appendToken(kind TokenType) error {
	l.Tokens = append(l.Tokens, createToken("", kind))
	return nil
}

func (l *Tokenizer) appendLiteral(kind TokenType, literal string) {
	l.Tokens = append(l.Tokens, createToken(literal, kind))
}

func (l *Tokenizer) atEnd() bool {
	return l.Pos >= len(l.S)
}

func (l *Tokenizer) advance() (byte, error) {
	if l.Pos+1 > len(l.S) {
		return '\x00', NewTokenizationError(EndOfInput, l.Pos, "Reached end of input")
	}
	l.Pos += 1
	return l.S[l.Pos-1], nil
}

func (l *Tokenizer) peek() (byte, error) {
	if l.Pos+1 > len(l.S) {
		return '\x00', NewTokenizationError(EndOfInput, l.Pos, "Reached end of input")
	}
	return l.S[l.Pos], nil
}

func (l *Tokenizer) current() (byte, error) {
	if l.atEnd() {
		return '\x00', NewTokenizationError(EndOfInput, l.Pos, "No tokens")
	}
	return l.S[l.Pos-1], nil
}

func (l *Tokenizer) lastToken() (Token, error) {
	if len(l.Tokens) == 0 {
		return Token{}, NewTokenizationError(EndOfInput, l.Pos, "No tokens")
	}
	return l.Tokens[len(l.Tokens)-1], nil
}

func (l *Tokenizer) Error(e error) error {
	return fmt.Errorf("Error: %s", e.Error())
}

type ErrorCode int
const (
    EndOfInput ErrorCode = iota
    InvalidInput
)

type TokenizationError struct {
	Message string
    Position int
    Code ErrorCode
}

func (e ErrorCode) String() string {
    switch e {
    case EndOfInput:
        return "EndOfInput"
    case InvalidInput:
        return "InvalidInput"
    }
    return "UnnamedErrorCode"
}

func NewTokenizationError(code ErrorCode, position int, message string) *TokenizationError {
    return &TokenizationError{
        Code: code,
        Position: position,
        Message: message,
    }
}

func (e *TokenizationError) Error() string {
    return fmt.Sprintf("%s error at position %d: %s", e.Code, e.Position, e.Message)
}


