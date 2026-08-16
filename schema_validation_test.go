package ntql

import "testing"

func validSchemaConfigForValidationTests() *schemaConfig {
	return &schemaConfig{
		Subjects: []schemaSubject{
			{
				Name:    "title",
				Aliases: []string{"name"},
				ValidVerbs: []schemaVerb{
					{Name: "equals", Aliases: []string{"eq"}},
				},
				ValidTypes: []string{"string"},
			},
		},
		FieldTypes: schemaFieldTypes{
			DateTypes:    []string{"due_date"},
			BoolTypes:    []string{"completed"},
			NumericTypes: []string{"priority"},
			StringTypes:  []string{"title"},
		},
	}
}

func TestValidateSchemaConfigRequiresNonNilConfig(t *testing.T) {
	err := validateSchemaConfig(nil)
	if err == nil {
		t.Fatalf("expected nil schema config to fail validation")
	}
}

func TestValidateSchemaConfigRequiresSubjects(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.Subjects = nil

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected schema with no subjects to fail validation")
	}
}

func TestValidateSchemaConfigRejectsDuplicateSubjectAlias(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.Subjects = append(cfg.Subjects, schemaSubject{
		Name:    "due",
		Aliases: []string{"name"},
		ValidVerbs: []schemaVerb{
			{Name: "equals", Aliases: []string{"eq"}},
		},
		ValidTypes: []string{"date"},
	})

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected duplicate subject alias to fail validation")
	}
}

func TestValidateSchemaConfigRejectsDuplicateVerbAlias(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.Subjects[0].ValidVerbs = []schemaVerb{
		{Name: "equals", Aliases: []string{"eq"}},
		{Name: "contains", Aliases: []string{"eq"}},
	}

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected duplicate verb alias to fail validation")
	}
}

func TestValidateSchemaConfigRejectsUnknownValidType(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.Subjects[0].ValidTypes = []string{"unknown_type"}

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected unknown valid type to fail validation")
	}
}

func TestValidateSchemaConfigRejectsDuplicateFieldTypeValue(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.FieldTypes.StringTypes = []string{"title", "title"}

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected duplicate field type values to fail validation")
	}
}

func TestValidateSchemaConfigRequiresFieldTypeValues(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()
	cfg.FieldTypes.BoolTypes = []string{}

	err := validateSchemaConfig(cfg)
	if err == nil {
		t.Fatalf("expected empty field type array to fail validation")
	}
}

func TestValidateSchemaConfigAcceptsValidSchema(t *testing.T) {
	cfg := validSchemaConfigForValidationTests()

	err := validateSchemaConfig(cfg)
	if err != nil {
		t.Fatalf("expected valid schema config, got error: %v", err)
	}
}
