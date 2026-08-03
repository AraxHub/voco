package postgres

import (
	"fmt"
	"strings"
)

// selectList собирает список колонок для SELECT; при alias != "" префиксирует alias.
func selectList(alias string, cols []string) string {
	if alias == "" {
		return strings.Join(cols, ", ")
	}
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = alias + "." + c
	}
	return strings.Join(out, ", ")
}

// placeholders собирает $1, $2, ... $n для VALUES.
func placeholders(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}
