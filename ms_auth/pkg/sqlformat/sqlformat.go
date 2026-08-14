package sqlformat

import "strings"

func MinifySQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
