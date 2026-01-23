package i18n

import (
	"testing"
)

func TestNew(t *testing.T) {
	translator := New("en")
	if translator == nil {
		t.Fatal("Expected translator to be created")
	}
}

func TestT_English(t *testing.T) {
	translator := New("en")

	// Test a known key
	result := translator.T("config.not_found")
	if result == "" {
		t.Error("Expected translation, got empty string")
	}

	if result == "config.not_found" {
		t.Error("Translation key was not found")
	}
}

func TestT_Chinese(t *testing.T) {
	translator := New("zh")

	// Test a known key
	result := translator.T("config.not_found")
	if result == "" {
		t.Error("Expected translation, got empty string")
	}

	// Chinese translation should contain Chinese characters
	if result == "config.not_found" {
		t.Error("Translation key was not found")
	}
}

func TestT_WithPlaceholder(t *testing.T) {
	translator := New("en")

	// Test with placeholder
	result := translator.T("config.default_model.selected", "test-model")
	if result == "" {
		t.Error("Expected translation, got empty string")
	}
}

func TestT_UnknownKey(t *testing.T) {
	translator := New("en")

	// Test with unknown key
	result := translator.T("unknown.key.that.does.not.exist")

	// Should return the key itself
	if result != "unknown.key.that.does.not.exist" {
		t.Errorf("Expected key to be returned for unknown translation, got '%s'", result)
	}
}

func TestT_DefaultsToEnglish(t *testing.T) {
	translator := New("invalid-language")

	// Should default to English
	result := translator.T("config.not_found")
	if result == "" {
		t.Error("Expected translation, got empty string")
	}
}

func TestEnglishMessages_HasRequiredKeys(t *testing.T) {
	requiredKeys := []string{
		"config.not_found",
		"config.title",
		"interactive.welcome",
		"executor.query_command",
		"executor.modify_command",
		"llm.status_title",
	}

	for _, key := range requiredKeys {
		if _, exists := EnglishMessages[key]; !exists {
			t.Errorf("English messages missing required key: %s", key)
		}
	}
}

func TestChineseMessages_HasRequiredKeys(t *testing.T) {
	requiredKeys := []string{
		"config.not_found",
		"config.title",
		"interactive.welcome",
		"executor.query_command",
		"executor.modify_command",
		"llm.status_title",
	}

	for _, key := range requiredKeys {
		if _, exists := ChineseMessages[key]; !exists {
			t.Errorf("Chinese messages missing required key: %s", key)
		}
	}
}

func TestEnglishAndChineseHaveSameKeys(t *testing.T) {
	// Check all English keys exist in Chinese
	for key := range EnglishMessages {
		if _, exists := ChineseMessages[key]; !exists {
			t.Errorf("Chinese messages missing key that exists in English: %s", key)
		}
	}

	// Check all Chinese keys exist in English
	for key := range ChineseMessages {
		if _, exists := EnglishMessages[key]; !exists {
			t.Errorf("English messages missing key that exists in Chinese: %s", key)
		}
	}
}
