package liquid

import (
	"testing"
	"time"
)

// strftime follows Ruby's Time#strftime: directives are expanded in one pass and any
// other text is copied as written.
func TestStrftimeRubyDirectives(t *testing.T) {
	afternoon := time.Date(2024, 3, 5, 14, 30, 7, 123456789, time.UTC)
	morning := time.Date(2024, 3, 5, 9, 5, 0, 0, time.FixedZone("CET", 3600))
	tests := []struct {
		when   time.Time
		format string
		want   string
	}{
		{afternoon, "%-d %B", "5 March"},
		{afternoon, "%d %B %Y, 2pm", "05 March 2024, 2pm"},
		{afternoon, "Copyright 2006-%Y", "Copyright 2006-2024"},
		{afternoon, "Jan 2 15:04:05 MST", "Jan 2 15:04:05 MST"},
		{afternoon, "%F %T", "2024-03-05 14:30:07"},
		{afternoon, "%s", "1709649007"},
		{afternoon, "%u %w", "2 2"},
		{afternoon, "%k|%l|%I|%p|%P", "14| 2|02|PM|pm"},
		{afternoon, "%-m/%-d/%y", "3/5/24"},
		{afternoon, "%^B %^a %#p", "MARCH TUE pm"},
		{afternoon, "%e", " 5"},
		{afternoon, "%10A|%-10A|", "   Tuesday|Tuesday|"},
		{afternoon, "%05d %_3m", "00005   3"},
		{afternoon, "%j", "065"},
		{afternoon, "%U %W %V %G %g", "09 10 10 2024 24"},
		{afternoon, "%L %N %3N", "123 123456789 123"},
		{afternoon, "%z %:z %Z", "+0000 +00:00 UTC"},
		{afternoon, "%D %R %r", "03/05/24 14:30 02:30:07 PM"},
		{afternoon, "%c", "Tue Mar  5 14:30:07 2024"},
		{afternoon, "%x %X", "03/05/24 14:30:07"},
		{afternoon, "%C %y %h", "20 24 Mar"},
		{afternoon, "%v", " 5-MAR-2024"},
		{afternoon, "100%% %q %", "100% %q %"},
		{afternoon, "%n%t", "\n\t"},
		{morning, "%-H:%M %-I%P", "9:05 9am"},
		{morning, "%H %M %S", "09 05 00"},
		{morning, "%z %:z %::z %Z", "+0100 +01:00 +01:00:00 CET"},
		{morning, "%A %a %B %b", "Tuesday Tue March Mar"},
	}
	for _, tt := range tests {
		got := strftime(&tt.when, tt.format)
		if got != tt.want {
			t.Errorf("strftime(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}

func TestToDateFormats(t *testing.T) {
	tests := []struct {
		in   interface{}
		want string // formatted with %Y-%m-%d %H:%M:%S %z
	}{
		{1709649007.0, ""},
		{int64(1709649007), ""},
		{"1709649007", ""},
		{"2024-03-05T14:30:07Z", "2024-03-05 14:30:07 +0000"},
		{"2024-03-05T14:30:07+01:00", "2024-03-05 14:30:07 +0100"},
		{"2024-03-05T14:30:07", "2024-03-05 14:30:07 +0000"},
		{"2024-03-05 14:30:07", "2024-03-05 14:30:07 +0000"},
		{"2024-03-05 14:30:07 +0100", "2024-03-05 14:30:07 +0100"},
		{"2024-03-05 14:30:07 UTC", "2024-03-05 14:30:07 +0000"},
		{"2024-03-05", "2024-03-05 00:00:00 +0000"},
		{"Fri Jul 16 01:00:00 2004", "2004-07-16 01:00:00 +0000"},
		{"March 5, 2024", "2024-03-05 00:00:00 +0000"},
		{"5 March 2024", "2024-03-05 00:00:00 +0000"},
	}
	for _, tt := range tests {
		got := ToDate(tt.in)
		if got == nil {
			t.Errorf("ToDate(%#v) = nil", tt.in)
			continue
		}
		if tt.want == "" {
			if got.Unix() != 1709649007 {
				t.Errorf("ToDate(%#v) = %v, want the timestamp 1709649007", tt.in, got)
			}
			continue
		}
		if s := got.Format("2006-01-02 15:04:05 -0700"); s != tt.want {
			t.Errorf("ToDate(%#v) = %s, want %s", tt.in, s, tt.want)
		}
	}
	for _, bad := range []interface{}{"", "not a date", true, []interface{}{}} {
		if got := ToDate(bad); got != nil {
			t.Errorf("ToDate(%#v) = %v, want nil", bad, got)
		}
	}
}
