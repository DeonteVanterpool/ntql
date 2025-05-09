package ntql

import "fmt"

type ScannerError struct {
	Message error
}

type ErrInvalidLexeme struct {
	Input byte
}

type ErrInvalidSubject struct {
	Position int
	Lexeme   Lexeme
}

type ErrEndOfInput struct{}

type ErrInvalidToken struct {
	Expected []TokenType
	Position int
	Lexeme   Lexeme
}

type ParserError struct {
	Message string
	Token   Token
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("Scanner Error: %s", e.Message.Error())
}

func (e ErrInvalidLexeme) Error() string {
	return fmt.Sprintf("Invalid character, %v", e.Input)
}

func (e ErrInvalidSubject) Error() string {
	return fmt.Sprintf("Invalid subject: %s", e.Lexeme)
}

func (e ErrEndOfInput) Error() string {
	return fmt.Sprintf("End of input")
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

	return fmt.Sprintf("Invalid lexeme '%s': expected type from [%s]", e.Lexeme, expected)
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("Error occurred at position %d: %s", e.Token.Position, e.Message)
}
