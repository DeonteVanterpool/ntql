package ntql

type Lexeme string

type Scanner struct {
	Lexemes []Lexeme
	Pos     int
	S       string
}

func NewScanner(s string) *Scanner {
	return &Scanner{Lexemes: []Lexeme{}, S: s, Pos: 0}
}

func (s *Scanner) ScanLexeme() (Lexeme, error) {

	if s.atEnd() {
		return "", ErrEndOfInput{}
	}

	l, err := s.match()

	if err != nil {
		return "", err
	}

	s.skipWhitespace()

	return l, nil
}

func (s *Scanner) LastLexeme() (Lexeme, error) {
	if len(s.Lexemes) == 0 {
		return "", ErrEndOfInput{}
	}

	return s.Lexemes[len(s.Lexemes)-1], nil
}

func (s *Scanner) HasNext() bool {
	return !s.atEnd()
}

func (s *Scanner) GetPosition() int {
	return s.Pos
}

func (s *Scanner) skipWhitespace() error {
	for !s.atEnd() {
		if !s.matchWhitespace() {
			break
		}

		_, err := s.advance()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Scanner) atEnd() bool {
	return s.Pos >= len(s.S)
}

func (s *Scanner) advance() (byte, error) {
	if s.atEnd() {
		return '\x00', ErrEndOfInput{}
	}
	s.Pos += 1
	return s.S[s.Pos-1], nil
}

func (s *Scanner) appendLexeme(l string) string {
	s.Lexemes = append(s.Lexemes, Lexeme(l))
	return l
}

func (s *Scanner) current() (byte, error) {
	if s.atEnd() {
		return '\x00', ErrEndOfInput{}
	}

	return s.S[s.Pos], nil
}

func (s *Scanner) match() (Lexeme, error) {
	if s.matchSymbol() {
		return s.consumeSymbol(), nil
	} else if s.matchQuote() {
		return s.consumeQuote(), nil
	} else if s.matchWhitespace() {
		s.skipWhitespace()
		return s.ScanLexeme()
	} else if s.matchAlphaNum() {
		return s.consumeAlphaNum(), nil
	} else {
		return "", ErrInvalidLexeme{Input: s.S[s.Pos]}
	}
}

func (s *Scanner) matchSymbol() bool {
	c, _ := s.current()

	return isSymbol(c)
}

func (s *Scanner) consumeSymbol() Lexeme {
	c, err := s.advance()
	if err != nil {
		panic(err)
	}

	return Lexeme(s.appendLexeme(string(c)))
}

func (s *Scanner) matchQuote() bool {
	c, err := s.current()
	if err != nil {
		panic(err)
	}

	return c == '"'
}

func (s *Scanner) consumeQuote() Lexeme {
	var l string
	escaped := false

	c, err := s.advance()
	if err != nil {
		panic(err)
	}

	l += string(c)

	for !s.atEnd() {
		c, err := s.advance()
		if err != nil {
			panic(err)
		}

		if c == '\\' && !escaped {
			escaped = true
			continue // skip adding this backslash character
		}

		if c == '"' && !escaped {
			l += string(c)
			break
		}

		escaped = false

		l += string(c)
	}

	return Lexeme(s.appendLexeme(l))
}

func (s *Scanner) matchWhitespace() bool {
	c, err := s.current()

	if err != nil {
		panic(err)
	}

	return c == ' '
}

func (s *Scanner) matchAlphaNum() bool {
	c, err := s.current()
	if err != nil {
		panic(err)
	}

	return alphaNumRegexp.MatchString(string(c))
}

func (s *Scanner) consumeAlphaNum() Lexeme {
	var l string

	for !s.atEnd() {
		if s.matchWhitespace() || s.matchSymbol() {
			break
		}

		c, err := s.advance()
		if err != nil {
			panic(err)
		}
		l += string(c)
	}

	return Lexeme(s.appendLexeme(l))
}
