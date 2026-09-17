package ntql

import (
	"testing"
)

func TestScanLexeme(t *testing.T) {
	s := NewScanner("tag.equals(hello OR goodbye) OR (date.before(2024-01-08) AND date.after(2024-01-09)) ")

	expected := []Lexeme{"tag", ".", "equals", "(", "hello", "OR", "goodbye", ")", "OR", "(", "date", ".", "before", "(", "2024-01-08", ")", "AND", "date", ".", "after", "(", "2024-01-09", ")", ")"}

	for _, e := range expected {
		v, err := s.ScanLexeme()
		if err != nil {
			break
		}
		if v != e {
			t.Errorf("Expected %s, got %s", e, v)
		}
	}
}

func TestScanString(t *testing.T) {
	s := NewScanner(`"hello" OR "goodbye" `)

	expected := []Lexeme{"\"hello\"", "OR", "\"goodbye\""}

	for _, e := range expected {
		v, err := s.ScanLexeme()
		if err != nil {
			break
		}
		if v != e {
			t.Errorf("Expected %s, got %s", e, v)
		}
	}
}

func TestScanEscapedString(t *testing.T) {
	s := NewScanner(`"hello \"world\"" OR "good\"bye" `)

	expected := []Lexeme{"\"hello \"world\"\"", "OR", "\"good\"bye\""}

	for _, e := range expected {
		v, err := s.ScanLexeme()
		if err != nil {
			break
		}
		if v != e {
			t.Errorf("Expected %s, got %s", e, v)
		}
	}
}
