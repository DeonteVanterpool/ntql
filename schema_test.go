package ntql

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedSchemaIsLoaded(t *testing.T) {
	schema := GetLoadedSchema()

	if len(schema.ValidSubjects) == 0 {
		t.Fatalf("expected subjects to be loaded")
	}
	if len(schema.DateTypes) == 0 || len(schema.BoolTypes) == 0 || len(schema.NumericTypes) == 0 || len(schema.StringTypes) == 0 {
		t.Fatalf("expected all field type arrays to be loaded")
	}

	subject, err := getSubject("deadline")
	if err != nil {
		t.Fatalf("expected alias lookup to work, got error: %v", err)
	}
	if subject.Name != "due" {
		t.Fatalf("expected alias 'deadline' to map to due, got %s", subject.Name)
	}
}

func TestLoadSchemaConfigFromYAMLValidation(t *testing.T) {
	_, err := loadSchemaConfigFromYAML([]byte(`
subjects:
  - name: title
    aliases: [name]
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [unknownType]
fieldTypes:
  dateTypes: [due_date]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [title]
`))

	if err == nil {
		t.Fatalf("expected invalid type schema to fail validation")
	}
}

func TestLoadSchemaConfigFromYAMLDuplicateAliasValidation(t *testing.T) {
	_, err := loadSchemaConfigFromYAML([]byte(`
subjects:
  - name: title
    aliases: [name]
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
  - name: due
    aliases: [name]
    validVerbs:
      - name: equals
        aliases: []
    validTypes: [date]
fieldTypes:
  dateTypes: [due_date]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [title]
`))

	if err == nil {
		t.Fatalf("expected duplicate alias schema to fail validation")
	}
}

func TestLoadSchemaFromFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "schema.yaml")
	err := os.WriteFile(path, []byte(`
subjects:
  - name: only
    aliases: [single]
    validVerbs:
      - name: equals
        aliases: [eq]
    validTypes: [string]
fieldTypes:
  dateTypes: [due_date]
  boolTypes: [completed]
  numericTypes: [priority]
  stringTypes: [title]
`), 0o600)
	if err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	if err := LoadSchemaFromFile(path); err != nil {
		t.Fatalf("expected schema file to load, got error: %v", err)
	}
	t.Cleanup(func() {
		if err := LoadEmbeddedSchema(); err != nil {
			t.Fatalf("failed to restore embedded schema: %v", err)
		}
	})

	subject, err := getSubject("single")
	if err != nil {
		t.Fatalf("expected alias lookup to work after loading file: %v", err)
	}
	if subject.Name != "only" {
		t.Fatalf("expected alias to map to loaded subject, got %s", subject.Name)
	}
}
