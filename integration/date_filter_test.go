package integration

import "testing"

// date formats like Ruby's strftime, and reads timestamps from JSON data (float64). The
// timestamp is noon UTC, so that its date is the same in most local time zones.
func TestDateFilter(t *testing.T) {
	data := `{"ts": 1709640000, "iso": "2024-03-05T14:30:07Z", "local": "2024-03-05 14:30:07 +0100"}`
	tests := []struct {
		template string
		expected string
	}{
		{`{{ ts | date: "%d/%m/%Y" }}`, "05/03/2024"},
		{`{{ iso | date: "%-d %B %Y" }}`, "5 March 2024"},
		{`{{ iso | date: "%d %B %Y, 2pm" }}`, "05 March 2024, 2pm"},
		{`{{ iso | date: "Copyright 2006-%Y" }}`, "Copyright 2006-2024"},
		{`{{ iso | date: "%F %T" }}`, "2024-03-05 14:30:07"},
		{`{{ local | date: "%H:%M %:z" }}`, "14:30 +01:00"},
		{`{{ "Fri Jul 16 01:00:00 2004" | date: "%m/%d/%Y" }}`, "07/16/2004"},
		{`{{ "not a date" | date: "%Y" }}`, "not a date"},
	}
	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			assertTemplateResult(t, tt.expected, tt.template, jsonAssigns(t, data))
		})
	}
}
