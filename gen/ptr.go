package main

import "strings"

func isPtr(typ string) bool {
	return strings.HasPrefix(typ, "*")
}

func enforcePtr(typ string) string {
	if isPtr(typ) {
		return ""
	}
	return "&"
}

func stripPtr(typ string) string {
	return strings.TrimPrefix(typ, "*")
}
