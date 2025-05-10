package ntql

import (
	"testing"
)

func TestCompletion1(t *testing.T) {
	engine := NewCompletionEngine([]string{"school", "work", "projects"})
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

func TestCompletion2(t *testing.T) {
	engine := NewCompletionEngine([]string{"school", "work", "projects"})
	suggestions, err := engine.Suggest("!tag.eq")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	t.Logf("Suggestions: %s", suggestions)
	expected := []string{"equals"}
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d", len(expected), len(suggestions))
	}
	for i, suggestion := range suggestions {
		if suggestion != expected[i] {
			t.Errorf("Expected suggestion %s, got %s", expected[i], suggestion)
		}
	}
}

func TestCompletion3(t *testing.T) {
	engine := NewCompletionEngine([]string{"school", "work", "projects"})
	suggestions, err := engine.Suggest("!tag.eq(")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	t.Logf("Suggestions: %s", suggestions)
	expected := []string{"school", "work", "projects"}
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d", len(expected), len(suggestions))
	}
	for i, suggestion := range suggestions {
		if suggestion != expected[i] {
			t.Errorf("Expected suggestion %s, got %s", expected[i], suggestion)
		}
	}
}

func TestCompletion4(t *testing.T) {
	engine := NewCompletionEngine([]string{"school", "work", "projects"})
	suggestions, err := engine.Suggest("")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	t.Logf("Suggestions: %s", suggestions)
	expected := []string{"title", "name", "due", "deadline", "status", "state", "priority", "project", "createdAt", "updatedAt", "completedAt", "createdBy", "tag"}
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d", len(expected), len(suggestions))
	}
	for i, suggestion := range suggestions {
		if suggestion != expected[i] {
			t.Errorf("Expected suggestion %s, got %s", expected[i], suggestion)
		}
	}
}

func TestCompletion5(t *testing.T) {
	engine := NewCompletionEngine([]string{"school", "work", "projects"})
	suggestions, err := engine.Suggest("!tag.eq(school) A")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	t.Logf("Suggestions: %s", suggestions)
	expected := []string{"AND"}
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d", len(expected), len(suggestions))
	}
	for i, suggestion := range suggestions {
		if suggestion != expected[i] {
			t.Errorf("Expected suggestion %s, got %s", expected[i], suggestion)
		}
	}
}
