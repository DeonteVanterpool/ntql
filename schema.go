package ntql

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed schema.yaml
var schemaFS embed.FS

var (
	validSubjects = []Subject{}

	date_types    = []string{}
	bool_types    = []string{}
	numeric_types = []string{}
	string_types  = []string{}
)

type schemaConfig struct {
	Subjects   []schemaSubject  `yaml:"subjects"`
	FieldTypes schemaFieldTypes `yaml:"fieldTypes"`
}

type schemaSubject struct {
	Name       string       `yaml:"name"`
	Aliases    []string     `yaml:"aliases"`
	ValidVerbs []schemaVerb `yaml:"validVerbs"`
	ValidTypes []string     `yaml:"validTypes"`
}

type schemaVerb struct {
	Name    string   `yaml:"name"`
	Aliases []string `yaml:"aliases"`
}

type schemaFieldTypes struct {
	DateTypes    []string `yaml:"dateTypes"`
	BoolTypes    []string `yaml:"boolTypes"`
	NumericTypes []string `yaml:"numericTypes"`
	StringTypes  []string `yaml:"stringTypes"`
}

type LoadedSchema struct {
	ValidSubjects []Subject
	DateTypes     []string
	BoolTypes     []string
	NumericTypes  []string
	StringTypes   []string
}

func init() {
	if err := LoadEmbeddedSchema(); err != nil {
		panic(fmt.Sprintf("failed to load schema.yaml: %v", err))
	}
}

func LoadEmbeddedSchema() error {
	schemaData, err := schemaFS.ReadFile("schema.yaml")
	if err != nil {
		return err
	}

	cfg, err := loadSchemaConfigFromYAML(schemaData)
	if err != nil {
		return err
	}

	return applySchemaConfig(cfg)
}

func LoadSchemaFromFile(path string) error {
	schemaData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	cfg, err := loadSchemaConfigFromYAML(schemaData)
	if err != nil {
		return err
	}

	return applySchemaConfig(cfg)
}

func GetLoadedSchema() LoadedSchema {
	return LoadedSchema{
		ValidSubjects: copySubjects(validSubjects),
		DateTypes:     append([]string{}, date_types...),
		BoolTypes:     append([]string{}, bool_types...),
		NumericTypes:  append([]string{}, numeric_types...),
		StringTypes:   append([]string{}, string_types...),
	}
}

func loadSchemaConfigFromYAML(data []byte) (*schemaConfig, error) {
	var cfg schemaConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := validateSchemaConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validateSchemaConfig(cfg *schemaConfig) error {
	if cfg == nil {
		return errors.New("schema config cannot be nil")
	}
	if len(cfg.Subjects) == 0 {
		return errors.New("schema must define at least one subject")
	}

	dtypeMap := map[string]DType{
		"string":   DTypeString,
		"int":      DTypeInt,
		"date":     DTypeDate,
		"tag":      DTypeTag,
		"datetime": DTypeDateTime,
	}

	seenSubjects := map[string]struct{}{}
	for _, subject := range cfg.Subjects {
		if subject.Name == "" {
			return errors.New("subject name cannot be empty")
		}
		key := toLowerCase(subject.Name)
		if _, exists := seenSubjects[key]; exists {
			return fmt.Errorf("duplicate subject or alias: %s", subject.Name)
		}
		seenSubjects[key] = struct{}{}

		if len(subject.ValidVerbs) == 0 {
			return fmt.Errorf("subject %s must define at least one verb", subject.Name)
		}
		if len(subject.ValidTypes) == 0 {
			return fmt.Errorf("subject %s must define at least one valid type", subject.Name)
		}

		for _, alias := range subject.Aliases {
			if alias == "" {
				return fmt.Errorf("subject %s contains empty alias", subject.Name)
			}
			aliasKey := toLowerCase(alias)
			if _, exists := seenSubjects[aliasKey]; exists {
				return fmt.Errorf("duplicate subject or alias: %s", alias)
			}
			seenSubjects[aliasKey] = struct{}{}
		}

		seenVerbs := map[string]struct{}{}
		for _, verb := range subject.ValidVerbs {
			if verb.Name == "" {
				return fmt.Errorf("subject %s contains empty verb name", subject.Name)
			}
			verbKey := toLowerCase(verb.Name)
			if _, exists := seenVerbs[verbKey]; exists {
				return fmt.Errorf("subject %s has duplicate verb or alias: %s", subject.Name, verb.Name)
			}
			seenVerbs[verbKey] = struct{}{}
			for _, alias := range verb.Aliases {
				if alias == "" {
					return fmt.Errorf("subject %s verb %s contains empty alias", subject.Name, verb.Name)
				}
				aliasKey := toLowerCase(alias)
				if _, exists := seenVerbs[aliasKey]; exists {
					return fmt.Errorf("subject %s has duplicate verb or alias: %s", subject.Name, alias)
				}
				seenVerbs[aliasKey] = struct{}{}
			}
		}

		for _, dtype := range subject.ValidTypes {
			if _, ok := dtypeMap[toLowerCase(dtype)]; !ok {
				return fmt.Errorf("subject %s contains unknown valid type: %s", subject.Name, dtype)
			}
		}
	}

	if err := validateFieldTypeArray("dateTypes", cfg.FieldTypes.DateTypes); err != nil {
		return err
	}
	if err := validateFieldTypeArray("boolTypes", cfg.FieldTypes.BoolTypes); err != nil {
		return err
	}
	if err := validateFieldTypeArray("numericTypes", cfg.FieldTypes.NumericTypes); err != nil {
		return err
	}
	if err := validateFieldTypeArray("stringTypes", cfg.FieldTypes.StringTypes); err != nil {
		return err
	}

	return nil
}

func validateFieldTypeArray(name string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s must define at least one value", name)
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("%s cannot contain empty values", name)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s contains duplicate value: %s", name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func applySchemaConfig(cfg *schemaConfig) error {
	dtypeMap := map[string]DType{
		"string":   DTypeString,
		"int":      DTypeInt,
		"date":     DTypeDate,
		"tag":      DTypeTag,
		"datetime": DTypeDateTime,
	}

	subjects := make([]Subject, 0, len(cfg.Subjects))
	for _, subject := range cfg.Subjects {
		validTypes := make([]DType, 0, len(subject.ValidTypes))
		for _, dtype := range subject.ValidTypes {
			mapped, ok := dtypeMap[toLowerCase(dtype)]
			if !ok {
				return fmt.Errorf("subject %s contains unknown valid type: %s", subject.Name, dtype)
			}
			validTypes = append(validTypes, mapped)
		}

		validVerbs := make([]Verb, 0, len(subject.ValidVerbs))
		for _, verb := range subject.ValidVerbs {
			validVerbs = append(validVerbs, Verb{Name: verb.Name, Aliases: append([]string{}, verb.Aliases...)})
		}

		subjects = append(subjects, Subject{
			Name:       subject.Name,
			Aliases:    append([]string{}, subject.Aliases...),
			ValidVerbs: validVerbs,
			ValidTypes: validTypes,
		})
	}

	validSubjects = subjects
	date_types = append([]string{}, cfg.FieldTypes.DateTypes...)
	bool_types = append([]string{}, cfg.FieldTypes.BoolTypes...)
	numeric_types = append([]string{}, cfg.FieldTypes.NumericTypes...)
	string_types = append([]string{}, cfg.FieldTypes.StringTypes...)

	return nil
}

func copySubjects(subjects []Subject) []Subject {
	copied := make([]Subject, 0, len(subjects))
	for _, subject := range subjects {
		validVerbs := make([]Verb, 0, len(subject.ValidVerbs))
		for _, verb := range subject.ValidVerbs {
			validVerbs = append(validVerbs, Verb{Name: verb.Name, Aliases: append([]string{}, verb.Aliases...)})
		}
		copied = append(copied, Subject{
			Name:       subject.Name,
			Aliases:    append([]string{}, subject.Aliases...),
			ValidVerbs: validVerbs,
			ValidTypes: append([]DType{}, subject.ValidTypes...),
		})
	}
	return copied
}
