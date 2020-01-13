package names

import (
	"regexp"
)

var nonAlphaNumeric = regexp.MustCompile("[^a-z0-9\\-]+")
var multiDashes = regexp.MustCompile("[\\-]{2,}")

// FormatDash takes some string and formats it into simple dash case
func FormatDash(value string) string {
	dashed := nonAlphaNumeric.ReplaceAllString(value, "-")
	return multiDashes.ReplaceAllString(dashed, "-")
}
