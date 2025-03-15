package tbql

import (
	"encoding/json"
	"testing"
)

func TestQueryExpr(t *testing.T) {
	q := &QueryBinaryOp{
		Left: &QueryUnaryOp{
			Operator: "!",
			Operand: &QueryBinaryOp{
				Left:  &QueryBinaryOp{Left: &QueryCondition{Field: "due_date", Value: "2021-01-01", Operator: "="}, Right: &QueryCondition{Field: "completed", Value: "true", Operator: "="}, Op: "&"},
				Right: &QueryCondition{Field: "priority", Value: "1", Operator: "="},
				Op:    "|",
			},
		},
		Right: &QueryCondition{Field: "title", Value: "test", Operator: "="},
		Op:    "&",
	}
    q_json, err := json.Marshal(q)
    if err != nil {
        t.Fatalf("json.Marshal() failed: %v", err)
    }
    t.Logf("Query: %s", q_json)
	sql, err := q.ToSQL()
	if err != nil {
		t.Fatalf("ToSQL() failed: %v", err)
	}
	expected := "((NOT (((due_date = '2021-01-01' AND completed = true) OR priority = 1))) AND title = 'test')"
	if sql != expected {
		t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
	}
}

func TestQueryExprTag1(t *testing.T) {
    q := &QueryCondition{Field: "tag", Value: "work", Operator: "="}
    sql, err := q.ToSQL()
    if err != nil {
        t.Fatalf("ToSQL() failed: %v", err)
    }
    q_json, err := json.Marshal(q)
    if err != nil {
        t.Fatalf("json.Marshal() failed: %v", err)
    }
    t.Logf("Query: %s", q_json)
    expected := "tag_id = (SELECT id FROM atomic_tags WHERE title = 'work')"
    if sql != expected {
        t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
    }
}

func TestQueryExprTag2(t *testing.T) {
    q := &QueryCondition{Field: "tag", Value: "1", Operator: "="}
    sql, err := q.ToSQL()
    if err != nil {
        t.Fatalf("ToSQL() failed: %v", err)
    }
    expected := "tag_id = (SELECT id FROM atomic_tags WHERE id = 1)"
    if sql != expected {
        t.Fatalf("ToSQL() returned %q, expected %q", sql, expected)
    }
}


