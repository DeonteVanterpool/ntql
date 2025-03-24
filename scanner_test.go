package tbql

import (
    "testing"
)

func TestScanner_ScanLexeme(t *testing.T) {
    s := NewScanner("tag.equals(hello OR goodbye) OR (date.before(2024-01-08) AND date.after(2024-01-09))")

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

