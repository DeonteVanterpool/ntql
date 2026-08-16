package ntql

import (
	"os"
	"path/filepath"
	"testing"
)

const validationTestSchemaYAML = `
subjects:
  - name: due
    aliases: [deadline]
    validVerbs:
      - name: equals
        aliases: [eq]
      - name: before
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
fieldTypes:
  dateTypes: [due_date, assigned_at]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [assignee_name, title]
tables:
  - name: tasks
    primaryKey: id
  - name: task_assignments
    primaryKey: id
joins:
  - fromTable: tasks
    toTable: task_assignments
    fromKey: id
    toKey: task_id
`

func TestSchemaTableDefinition(t *testing.T) {
	cfg, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed loading schema config: %v", err)
	}
	if len(cfg.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(cfg.Tables))
	}
	if cfg.Tables[0].Name != "tasks" || cfg.Tables[0].PrimaryKey != "id" {
		t.Fatalf("unexpected first table definition: %+v", cfg.Tables[0])
	}
}

func TestSchemaTableMapping(t *testing.T) {
	cfg, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed loading schema config: %v", err)
	}
	if err := applySchemaConfig(cfg); err != nil {
		t.Fatalf("failed applying schema config: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	subject, err := getSubject("owner")
	if err != nil {
		t.Fatalf("expected owner alias to resolve, got: %v", err)
	}
	if subject.Table != "task_assignments" {
		t.Fatalf("expected subject table to be task_assignments, got %s", subject.Table)
	}
}

func TestSchemaValidation(t *testing.T) {
	if _, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML)); err != nil {
		t.Fatalf("expected valid schema, got: %v", err)
	}
}

func TestSchemaFieldTypeConsistency(t *testing.T) {
	cfg, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed loading schema config: %v", err)
	}
	if err := applySchemaConfig(cfg); err != nil {
		t.Fatalf("failed applying schema config: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	loaded := GetLoadedSchema()
	if len(loaded.DateTypes) == 0 || len(loaded.BoolTypes) == 0 || len(loaded.NumericTypes) == 0 || len(loaded.StringTypes) == 0 {
		t.Fatalf("expected all field type arrays to be populated")
	}
}

func TestSchemaSubjectAliases(t *testing.T) {
	cfg, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed loading schema config: %v", err)
	}
	if err := applySchemaConfig(cfg); err != nil {
		t.Fatalf("failed applying schema config: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	subject, err := getSubject("deadline")
	if err != nil {
		t.Fatalf("expected alias deadline to resolve, got: %v", err)
	}
	if subject.Name != "due" {
		t.Fatalf("expected deadline alias to map to due, got %s", subject.Name)
	}
}

func TestSchemaVerbAliases(t *testing.T) {
	cfg, err := loadSchemaConfigFromYAML([]byte(validationTestSchemaYAML))
	if err != nil {
		t.Fatalf("failed loading schema config: %v", err)
	}
	if err := applySchemaConfig(cfg); err != nil {
		t.Fatalf("failed applying schema config: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	parser := NewParser([]Token{
		{Kind: TokenSubject, Literal: "due"},
		{Kind: TokenDot, Literal: ""},
		{Kind: TokenVerb, Literal: "eq"},
		{Kind: TokenLParen, Literal: ""},
		{Kind: TokenDate, Literal: "2026-01-01"},
		{Kind: TokenRParen, Literal: ""},
	})
	expr, err := parser.Parse()
	if err != nil {
		t.Fatalf("expected parser to resolve verb alias eq, got: %v", err)
	}

	condition, ok := expr.(*QueryCondition)
	if !ok {
		t.Fatalf("expected parsed expression to be a condition, got %T", expr)
	}
	if condition.Operator != OperatorEq {
		t.Fatalf("expected eq alias to map to equals, got %s", condition.Operator)
	}
}

func TestSchemaLoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "schema.yaml")
	if err := os.WriteFile(path, []byte(validationTestSchemaYAML), 0o600); err != nil {
		t.Fatalf("failed writing schema file: %v", err)
	}

	if err := LoadSchemaFromFile(path); err != nil {
		t.Fatalf("expected file schema to load, got: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	schema := GetLoadedSchema()
	if len(schema.Tables) == 0 {
		t.Fatalf("expected tables to be loaded from file schema")
	}
}

func TestSchemaDuplicateSubject(t *testing.T) {
	_, err := loadSchemaConfigFromYAML([]byte(`
subjects:
  - name: due
    aliases: [deadline]
    validVerbs:
      - name: equals
        aliases: []
    validTypes: [date]
    table: tasks
    column: due_date
  - name: due
    aliases: []
    validVerbs:
      - name: equals
        aliases: []
    validTypes: [date]
    table: tasks
    column: due_date
fieldTypes:
  dateTypes: [due_date]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [title]
tables:
  - name: tasks
    primaryKey: id
`))
	if err == nil {
		t.Fatalf("expected duplicate subjects to be rejected")
	}
}

func TestSchemaForeignKeyDefinition(t *testing.T) {
	t.Skip("placeholder for future foreign key validation tests")
}

func TestSchemaInvalidTableReference(t *testing.T) {
	t.Skip("placeholder for future invalid table reference tests")
}

func TestSchemaJoinPathResolution(t *testing.T) {
	t.Skip("placeholder for future join path resolution tests")
}

func TestSchemaCircularDependencies(t *testing.T) {
	t.Skip("placeholder for future circular dependency tests")
}

func TestSchemaMultipleJoinPaths(t *testing.T) {
	t.Skip("placeholder for future multiple join path tests")
}
