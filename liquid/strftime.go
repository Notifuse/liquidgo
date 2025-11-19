package liquid

import (
	"strings"
	"time"
)

// strftime formats a time using strftime-style format codes.
// Converts Ruby/C strftime format codes to Go's time.Format.
func strftime(t *time.Time, format string) string {
	if t == nil {
		return ""
	}

	// Map of strftime codes to Go format codes
	// Go uses the reference time: Mon Jan 2 15:04:05 MST 2006
	replacements := map[string]string{
		"%a": "Mon",                     // Abbreviated weekday name
		"%A": "Monday",                  // Full weekday name
		"%b": "Jan",                     // Abbreviated month name
		"%B": "January",                 // Full month name
		"%c": "Mon Jan 2 15:04:05 2006", // Preferred date and time
		"%d": "02",                      // Day of the month (01-31)
		"%e": "_2",                      // Day of the month, space-padded ( 1-31)
		"%H": "15",                      // Hour (00-23)
		"%I": "03",                      // Hour (01-12)
		"%j": "002",                     // Day of the year (001-366)
		"%m": "01",                      // Month (01-12)
		"%M": "04",                      // Minute (00-59)
		"%p": "PM",                      // AM or PM
		"%P": "pm",                      // am or pm (lowercase)
		"%S": "05",                      // Second (00-60)
		"%U": "",                        // Week number of year (Sunday first)
		"%w": "",                        // Day of the week (0-6, Sunday is 0)
		"%W": "",                        // Week number of year (Monday first)
		"%x": "01/02/06",                // Preferred date representation
		"%X": "15:04:05",                // Preferred time representation
		"%y": "06",                      // Year without century (00-99)
		"%Y": "2006",                    // Year with century
		"%z": "-0700",                   // Numeric timezone offset (+0000)
		"%Z": "MST",                     // Time zone name
		"%%": "%",                       // Literal percent sign
	}

	result := format

	// Replace all strftime codes with Go format codes
	for code, goFormat := range replacements {
		if goFormat == "" {
			// Handle special cases that need calculation
			switch code {
			case "%w":
				// Day of week as number (0-6, Sunday=0)
				// We'll handle this specially after format
				continue
			case "%U", "%W":
				// Week number - not directly supported by Go
				// Would need custom calculation
				continue
			}
		}
		result = strings.ReplaceAll(result, code, goFormat)
	}

	// Special handling for %P (lowercase am/pm)
	if strings.Contains(format, "%P") {
		formatted := t.Format(result)
		formatted = strings.ReplaceAll(formatted, "AM", "am")
		formatted = strings.ReplaceAll(formatted, "PM", "pm")
		return formatted
	}

	return t.Format(result)
}
