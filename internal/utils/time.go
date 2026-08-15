package utils

import (
	"strings"
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format(time.DateTime)
}

func ValidateKeyExists(m map[any]any, keys []string) bool {
	return true
}

// 将标准时间格式化格式转为 Go 时间格式化格式
// Supported tokens:
// yyyy -> 2006  (4-digit year)
// yy   -> 06    (2-digit year)
// MM   -> 01    (2-digit month)
// M    -> 1     (1-2 digit month)
// dd   -> 02    (2-digit day)
// d    -> 2     (1-2 digit day)
// HH   -> 15    (24-hour, 2-digit)
// H    -> 15    (24-hour, 1-2 digit - Go has no single H, use 15)
// hh   -> 03    (12-hour, 2-digit)
// h    -> 3     (12-hour, 1-2 digit)
// mm   -> 04    (2-digit minute)
// m    -> 4     (1-2 digit minute)
// ss   -> 05    (2-digit second)
// s    -> 5     (1-2 digit second)
// SSS  -> .000  (milliseconds)
// SS   -> .00   (hundredths of second)
// S    -> .0    (tenths of second)
// a    -> PM    (AM/PM marker)
// z    -> MST   (time zone name)
// Z    -> -0700 (time zone offset RFC822)
// X    -> -07   (time zone offset ISO8601 hour)
// XX   -> -0700 (time zone offset ISO8601 hour+minute)
// XXX  -> -07:00 (time zone offset ISO8601 with colon)
func ConvertToGoFormat(format string) string {
	// var err error
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("invalid format: %v", r)
	// 	}
	// }()

	var sb strings.Builder
	i := 0
	runes := []rune(format)

	for i < len(runes) {
		r := runes[i]

		// Count consecutive identical letters
		count := 1
		for i+count < len(runes) && runes[i+count] == r {
			count++
		}

		switch r {
		case 'y':
			if count >= 4 {
				sb.WriteString("2006")
			} else {
				sb.WriteString("06")
			}
		case 'M':
			if count >= 2 {
				sb.WriteString("01")
			} else {
				sb.WriteString("1")
			}
		case 'd':
			if count >= 2 {
				sb.WriteString("02")
			} else {
				sb.WriteString("2")
			}
		case 'H':
			// Go has no single-H 24h format; "15" handles both 1 and 2 digit for parsing,
			// but for formatting always produces 2 digits. This is the closest match.
			sb.WriteString("15")
		case 'h':
			if count >= 2 {
				sb.WriteString("03")
			} else {
				sb.WriteString("3")
			}
		case 'm':
			if count >= 2 {
				sb.WriteString("04")
			} else {
				sb.WriteString("4")
			}
		case 's':
			if count >= 2 {
				sb.WriteString("05")
			} else {
				sb.WriteString("5")
			}
		case 'S':
			// Fractional seconds - in Go these are separated by a period before the 0s
			// We output the preceding dot here; consecutive S's share one dot
			if i == 0 || (i > 0 && !strings.HasSuffix(sb.String(), ".")) {
				sb.WriteString(".")
			}
			for j := 0; j < count; j++ {
				sb.WriteString("0")
			}
		case 'a':
			sb.WriteString("PM")
		case 'z':
			sb.WriteString("MST")
		case 'Z':
			sb.WriteString("-0700")
		case 'X':
			switch count {
			case 1:
				sb.WriteString("-07")
			case 2:
				sb.WriteString("-0700")
			default:
				sb.WriteString("-07:00")
			}
		case '\'':
			// Quoted literal in Java format
			// Find closing quote
			j := i + 1
			for j < len(runes) && runes[j] != '\'' {
				sb.WriteRune(runes[j])
				j++
			}
			// If two consecutive single quotes inside literal, that's an escaped quote
			// For simplicity, we just skip the closing quote position
			if j < len(runes) {
				i = j // will advance by count (1) below
			} else {
				i = len(runes) - 1
			}
		default:
			// Non-pattern character, write as-is
			for j := 0; j < count; j++ {
				sb.WriteRune(r)
			}
		}

		i += count
	}

	return sb.String()
}
