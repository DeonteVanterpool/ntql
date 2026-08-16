package ntql

import "testing"

func TestCrossTableJoinTagEqualsString(t *testing.T) {
	lexer := NewLexer(`tag.equals(work)`)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	parser := NewParser(tokens)
	expr, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sql, err := expr.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL() failed: %v", err)
	}

	expected := "tag_id = (SELECT id FROM atomic_tags WHERE title = 'work')"
	if sql != expected {
		t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
	}
}

func TestCrossTableJoinTagEqualsNumericID(t *testing.T) {
	lexer := NewLexer(`tag.equals(42)`)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	parser := NewParser(tokens)
	expr, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sql, err := expr.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL() failed: %v", err)
	}

	expected := "tag_id = (SELECT id FROM atomic_tags WHERE id = 42)"
	if sql != expected {
		t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
	}
}

func TestCrossTableJoinComposesWithMainTableFilter(t *testing.T) {
	lexer := NewLexer(`tag.equals(work) AND title.contains("roadmap")`)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	parser := NewParser(tokens)
	expr, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sql, err := expr.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL() failed: %v", err)
	}

	expected := "(tag_id = (SELECT id FROM atomic_tags WHERE title = 'work') AND title LIKE '%roadmap%')"
	if sql != expected {
		t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
	}
}
