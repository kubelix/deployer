package names

import (
	"regexp"
	"strings"
)

var nonAlphaNumeric = regexp.MustCompile("[^a-z0-9\\-]+")
var multiDashes = regexp.MustCompile("[\\-]{2,}")

// FormatDash takes some string and formats it into simple dash case
func FormatDash(value string) string {
	result := nonAlphaNumeric.ReplaceAllString(value, "-")
	result = multiDashes.ReplaceAllString(result, "-")
	return TrimDashes(result)
}

// TrimDashes cuts the string to 63 chars and trims all slashes
func TrimDashes(value string) string {
	if len(value) > 63 {
		value = value[:62]
	}

	return strings.Trim(value, "-")
}
