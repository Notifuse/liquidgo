package liquid

import (
	"strconv"
	"strings"
	"time"
)

// strftime formats a time like Ruby's Time#strftime. The format is scanned once:
// each directive %[flags][width]conversion is replaced by its value and any other text
// is copied as written. Flags: "-" no padding, "_" pad with spaces, "0" pad with zeros,
// "^" upper case, "#" change case, ":" (for %z) separate hours and minutes. An unknown
// directive is copied as written.
func strftime(t *time.Time, format string) string {
	if t == nil {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(format); {
		if format[i] != '%' {
			b.WriteByte(format[i])
			i++
			continue
		}
		start := i
		i++
		var d strftimeDirective
		for ; i < len(format) && strings.IndexByte("-_0^#:", format[i]) >= 0; i++ {
			if format[i] == ':' {
				d.colons++
			} else {
				d.flags += string(format[i])
			}
		}
		d.width = -1
		for widthStart := i; i < len(format) && format[i] >= '0' && format[i] <= '9'; i++ {
			d.width, _ = strconv.Atoi(format[widthStart : i+1])
		}
		if i >= len(format) {
			b.WriteString(format[start:])
			break
		}
		d.conversion = format[i]
		i++
		if s, ok := d.format(t); ok {
			b.WriteString(s)
		} else {
			b.WriteString(format[start:i])
		}
	}
	return b.String()
}

type strftimeDirective struct {
	flags      string
	colons     int
	width      int
	conversion byte
}

func (d strftimeDirective) has(flag byte) bool {
	return strings.IndexByte(d.flags, flag) >= 0
}

// number pads n to the directive's width (or defaultWidth) with zeros, or spaces when
// defaultPad is ' ', following the flags.
func (d strftimeDirective) number(n int64, defaultWidth int, defaultPad byte) string {
	s := strconv.FormatInt(n, 10)
	negative := n < 0
	if negative {
		s = s[1:]
	}
	pad := defaultPad
	switch {
	case d.has('-'):
		pad = 0
	case d.has('_'):
		pad = ' '
	case d.has('0'):
		pad = '0'
	}
	width := defaultWidth
	if d.width >= 0 {
		width = d.width
	}
	if negative {
		width--
	}
	if pad != 0 {
		for len(s) < width {
			if pad == '0' {
				s = "0" + s
			} else {
				s = " " + s
			}
		}
	}
	if negative {
		s = "-" + s
	}
	return s
}

// text applies case flags and pads s with spaces (or zeros with the "0" flag) to the
// directive's width. swapToLower is the "#" behaviour of %p, which lowers instead.
func (d strftimeDirective) text(s string, swapToLower bool) string {
	if d.has('^') {
		s = strings.ToUpper(s)
	}
	if d.has('#') {
		if swapToLower {
			s = strings.ToLower(s)
		} else {
			s = strings.ToUpper(s)
		}
	}
	if d.width > 0 && !d.has('-') {
		pad := " "
		if d.has('0') {
			pad = "0"
		}
		if n := len([]rune(s)); n < d.width {
			s = strings.Repeat(pad, d.width-n) + s
		}
	}
	return s
}

func (d strftimeDirective) composite(t *time.Time, format string) string {
	return d.text(strftime(t, format), false)
}

func (d strftimeDirective) format(t *time.Time) (string, bool) {
	hour12 := t.Hour() % 12
	if hour12 == 0 {
		hour12 = 12
	}
	switch d.conversion {
	case 'Y':
		return d.number(int64(t.Year()), 4, '0'), true
	case 'C':
		return d.number(int64(floorDiv(t.Year(), 100)), 2, '0'), true
	case 'y':
		return d.number(int64(floorMod(t.Year(), 100)), 2, '0'), true
	case 'm':
		return d.number(int64(t.Month()), 2, '0'), true
	case 'B':
		return d.text(t.Month().String(), false), true
	case 'b', 'h':
		return d.text(t.Month().String()[:3], false), true
	case 'd':
		return d.number(int64(t.Day()), 2, '0'), true
	case 'e':
		return d.number(int64(t.Day()), 2, ' '), true
	case 'j':
		return d.number(int64(t.YearDay()), 3, '0'), true
	case 'H':
		return d.number(int64(t.Hour()), 2, '0'), true
	case 'k':
		return d.number(int64(t.Hour()), 2, ' '), true
	case 'I':
		return d.number(int64(hour12), 2, '0'), true
	case 'l':
		return d.number(int64(hour12), 2, ' '), true
	case 'P':
		return d.text(strings.ToLower(meridian(t)), false), true
	case 'p':
		return d.text(meridian(t), true), true
	case 'M':
		return d.number(int64(t.Minute()), 2, '0'), true
	case 'S':
		return d.number(int64(t.Second()), 2, '0'), true
	case 'L', 'N':
		digits := 3
		if d.conversion == 'N' {
			digits = 9
		}
		if d.width > 0 {
			digits = d.width
		}
		s := strconv.Itoa(t.Nanosecond())
		s = strings.Repeat("0", 9-len(s)) + s
		if digits <= 9 {
			return s[:digits], true
		}
		return s + strings.Repeat("0", digits-9), true
	case 'z':
		return d.zone(t), true
	case 'Z':
		return d.text(t.Format("MST"), false), true
	case 'A':
		return d.text(t.Weekday().String(), false), true
	case 'a':
		return d.text(t.Weekday().String()[:3], false), true
	case 'u':
		wday := int64(t.Weekday())
		if wday == 0 {
			wday = 7
		}
		return d.number(wday, 1, '0'), true
	case 'w':
		return d.number(int64(t.Weekday()), 1, '0'), true
	case 'G':
		year, _ := t.ISOWeek()
		return d.number(int64(year), 4, '0'), true
	case 'g':
		year, _ := t.ISOWeek()
		return d.number(int64(floorMod(year, 100)), 2, '0'), true
	case 'V':
		_, week := t.ISOWeek()
		return d.number(int64(week), 2, '0'), true
	case 'U':
		return d.number(int64((t.YearDay()-1+7-int(t.Weekday()))/7), 2, '0'), true
	case 'W':
		return d.number(int64((t.YearDay()-1+7-(int(t.Weekday())+6)%7)/7), 2, '0'), true
	case 's':
		return d.number(t.Unix(), 1, '0'), true
	case 'n':
		return "\n", true
	case 't':
		return "\t", true
	case '%':
		return "%", true
	case 'c':
		return d.composite(t, "%a %b %e %H:%M:%S %Y"), true
	case 'D', 'x':
		return d.composite(t, "%m/%d/%y"), true
	case 'F':
		return d.composite(t, "%Y-%m-%d"), true
	case 'T', 'X':
		return d.composite(t, "%H:%M:%S"), true
	case 'R':
		return d.composite(t, "%H:%M"), true
	case 'r':
		return d.composite(t, "%I:%M:%S %p"), true
	case 'v':
		return d.composite(t, "%e-%^b-%4Y"), true
	case '+':
		return d.composite(t, "%a %b %e %H:%M:%S %Z %Y"), true
	}
	return "", false
}

func meridian(t *time.Time) string {
	if t.Hour() < 12 {
		return "AM"
	}
	return "PM"
}

// zone formats the UTC offset: +hhmm, +hh:mm with %:z, +hh:mm:ss with %::z.
func (d strftimeDirective) zone(t *time.Time) string {
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	two := func(n int) string {
		if n < 10 {
			return "0" + strconv.Itoa(n)
		}
		return strconv.Itoa(n)
	}
	h, m, s := two(offset/3600), two(offset%3600/60), two(offset%60)
	switch d.colons {
	case 0:
		return sign + h + m
	case 1:
		return sign + h + ":" + m
	default:
		return sign + h + ":" + m + ":" + s
	}
}

func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func floorMod(a, b int) int {
	return a - b*floorDiv(a, b)
}
