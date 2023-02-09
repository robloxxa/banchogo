package banchogo

import (
	"unicode/utf8"
)

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length])
}

func RandomString() string {
	// TODO:
	return ""
}
