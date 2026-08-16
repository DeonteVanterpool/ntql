package ntql

import (
	"strings"
	"testing"
)

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

const joinTestSchemaYAML = `
subjects:
  - name: title
    aliases: []
    validVerbs:
      - name: equals
        aliases: [eq]
      - name: contains
        aliases: []
    validTypes: [string]
    table: tasks
    column: title
  - name: status
    aliases: [state]
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
    table: tasks
    column: status
  - name: due
    aliases: [deadline]
    validVerbs:
      - name: after
        aliases: []
      - name: before
        aliases: []
      - name: equals
        aliases: []
    validTypes: [date]
    table: tasks
    column: due_date
  - name: assignee
    aliases: [owner]
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
    table: task_assignments
    column: assignee_name
  - name: assignmentStatus
    aliases: []
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
    table: task_assignments
    column: assignment_status
  - name: assignedAt
    aliases: []
    validVerbs:
      - name: after
        aliases: []
      - name: before
        aliases: []
    validTypes: [date]
    table: task_assignments
    column: assigned_at
  - name: project
    aliases: []
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
    table: projects
    column: name
fieldTypes:
  dateTypes: [due_date, assigned_at]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [title, status, assignee_name, assignment_status, name]
tables:
  - name: tasks
    primaryKey: id
  - name: task_assignments
    primaryKey: id
  - name: projects
    primaryKey: id
joins:
  - fromTable: tasks
    toTable: task_assignments
    fromKey: id
    toKey: task_id
  - fromTable: tasks
    toTable: projects
    fromKey: project_id
    toKey: id
`

func loadJoinTestSchema(t *testing.T) {
	t.Helper()
	cfg, err := loadSchemaConfigFromYAML([]byte(joinTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed to load join test schema: %v", err)
	}
	if err := applySchemaConfig(cfg); err != nil {
		t.Fatalf("failed to apply join test schema: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})
}

func mustBuildJoinSQL(t *testing.T, expr QueryExpr) string {
	t.Helper()
	sql, err := BuildSQLJoinQuery(expr, JoinQueryOptions{})
	if err != nil {
		t.Fatalf("BuildSQLJoinQuery failed: %v", err)
	}
	return sql
}

func TestJoinSimpleCrossTable(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		&QueryCondition{Field: "title", Operator: OperatorEq, Value: "Bug"},
		&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Alice"},
	))

	assertStringContainsAll(t, sql,
		"FROM tasks t0",
		"LEFT JOIN task_assignments t1 ON t0.id = t1.task_id",
		"(t0.title = 'Bug' AND t1.assignee_name = 'Alice')",
	)
}

func TestJoinMultipleFields(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		NewQueryAnd(
			&QueryCondition{Field: "title", Operator: OperatorEq, Value: "Issue"},
			&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Bob"},
		),
		&QueryCondition{Field: "assignmentStatus", Operator: OperatorEq, Value: "active"},
	))

	assertStringContainsAll(t, sql,
		"LEFT JOIN task_assignments t1 ON t0.id = t1.task_id",
		"t0.title = 'Issue'",
		"t1.assignee_name = 'Bob'",
		"t1.assignment_status = 'active'",
	)
}

func TestJoinWithOR(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryOr(
		&QueryCondition{Field: "title", Operator: OperatorEq, Value: "Hotfix"},
		&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Cara"},
	))

	assertStringContainsAll(t, sql, "(t0.title = 'Hotfix' OR t1.assignee_name = 'Cara')")
}

func TestJoinWithNegation(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryNot(&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Dave"}))
	assertStringContainsAll(t, sql, "(NOT (t1.assignee_name = 'Dave'))")
}

func TestJoinDeduplication(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Eve"},
		&QueryCondition{Field: "assignmentStatus", Operator: OperatorEq, Value: "pending"},
	))

	if countOccurrences(sql, "LEFT JOIN task_assignments") != 1 {
		t.Fatalf("expected exactly one join to task_assignments, got SQL: %s", sql)
	}
}

func TestJoinThreeTables(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Frank"},
		&QueryCondition{Field: "project", Operator: OperatorEq, Value: "Apollo"},
	))

	assertStringContainsAll(t, sql,
		"LEFT JOIN task_assignments t1 ON t0.id = t1.task_id",
		"LEFT JOIN projects t2 ON t0.project_id = t2.id",
		"(t1.assignee_name = 'Frank' AND t2.name = 'Apollo')",
	)
}

func TestJoinAliasing(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		&QueryCondition{Field: "title", Operator: OperatorEq, Value: "Roadmap"},
		&QueryCondition{Field: "project", Operator: OperatorEq, Value: "Platform"},
	))

	assertStringContainsAll(t, sql, "FROM tasks t0", "LEFT JOIN projects t1 ON t0.project_id = t1.id", "t0.title", "t1.name")
}

func TestJoinComplexExpression(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		NewQueryOr(
			&QueryCondition{Field: "title", Operator: OperatorCnt, Value: "bug"},
			&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Grace"},
		),
		NewQueryNot(&QueryCondition{Field: "project", Operator: OperatorEq, Value: "Legacy"}),
	))

	assertStringContainsAll(t, sql,
		"((t0.title LIKE '%bug%' OR t1.assignee_name = 'Grace') AND (NOT (t2.name = 'Legacy')))",
	)
}

func TestJoinSelectDistinct(t *testing.T) {
	loadJoinTestSchema(t)

	sql, err := BuildSQLJoinQuery(
		&QueryCondition{Field: "assignee", Operator: OperatorEq, Value: "Hank"},
		JoinQueryOptions{Distinct: true},
	)
	if err != nil {
		t.Fatalf("BuildSQLJoinQuery failed: %v", err)
	}
	assertStringContainsAll(t, sql, "SELECT DISTINCT t0.*")
}

func TestJoinDateRangeAcrossTables(t *testing.T) {
	loadJoinTestSchema(t)

	sql := mustBuildJoinSQL(t, NewQueryAnd(
		&QueryCondition{Field: "due", Operator: OperatorGte, Value: "2026-01-01"},
		&QueryCondition{Field: "assignedAt", Operator: OperatorLte, Value: "2026-12-31"},
	))

	assertStringContainsAll(t, sql,
		"t0.due_date >= '2026-01-01'",
		"t1.assigned_at <= '2026-12-31'",
	)
}

func assertStringContainsAll(t *testing.T, input string, expected ...string) {
	t.Helper()
	for _, item := range expected {
		if !strings.Contains(input, item) {
			t.Fatalf("expected SQL to contain %q, got: %s", item, input)
		}
	}
}

func countOccurrences(input, fragment string) int {
	return strings.Count(input, fragment)
}
