package sqlformat

import (
	"fmt"
	"strings"
)

func MinifySQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func NamedQuery(query string, params map[string]any) (string, []any) {
	args := []any{}
	i := 1

	for key, value := range params {
		placeholder := fmt.Sprintf("$%d", i)
		paramName := ":" + key
		query = strings.ReplaceAll(query, paramName, placeholder)
		args = append(args, value)
		i++
	}

	return query, args
}
