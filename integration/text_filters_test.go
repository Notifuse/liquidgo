package integration

import "testing"

// Text is handled by characters, as Ruby strings are, so accented and non-Latin text is
// never cut inside a character.
func TestTextFiltersCountCharacters(t *testing.T) {
	tests := []struct {
		template string
		expected string
	}{
		{`{{ name | capitalize }}`, "Élodie"},
		{`{{ "Événement spécial" | truncate: 6 }}`, "Évé..."},
		{`{{ "测试测试测试测试" | truncate: 5 }}`, "测试..."},
		{`{{ "été" | size }}`, "3"},
		{`{{ name.size }}`, "6"},
		{`{{ "été" | slice: 0, 2 }}`, "ét"},
		{`{{ "Bonjour" | slice: -3, 2 }}`, "ou"},
		{`{{ "un deux trois" | truncatewords: 0 }}`, "un..."},
		{`{{ "a,b,," | split: "," | size }}`, "2"},
		{`{{ 12345 | slice: 0, 2 }}`, "12"},
		{"{{ \"un\u00a0deux trois\" | truncatewords: 1 }}", "un\u00a0deux..."},
		{`{{ "un deux " | truncatewords: 2 }}`, "un deux..."},
		{`{{ "abcdef" | truncate: "7.0" }}`, "abcdef"},
		{`{{ 2024 | date: "%s" }}`, "2024"}, // a timestamp: 1970 in UTC
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, map[string]interface{}{"name": "élodie"})
		})
	}
}
