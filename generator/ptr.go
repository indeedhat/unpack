package generator

import "strings"

func isPtr(typ string) bool {
	return strings.HasPrefix(typ, "*")
}

func stripPtr(typ string) string {
	return strings.TrimPrefix(typ, "*")
}
