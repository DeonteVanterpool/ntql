package ntql

import (
	"testing"
)

func TestCompletion1(t *testing.T) {
	engine := NewAutocompleteEngine([]string{"school", "work", "projects"})
	suggestions, err := engine.Suggest("!ta")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	t.Logf("Suggestions: %s", suggestions)

	expected := []string{"tag", "state", "status"}
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d", len(expected), len(suggestions))
	}

	for i, suggestion := range suggestions {
		if suggestion != expected[i] {
			t.Errorf("Expected suggestion %s, got %s", expected[i], suggestion)
		}
	}
}
