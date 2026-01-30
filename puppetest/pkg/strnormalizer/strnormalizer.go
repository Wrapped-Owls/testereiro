package strnormalizer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func ToSnakeCase(str string) string {
	var builder strings.Builder
	// Grow the builder to hold the string + some changes
	{
		strLen := len(str)
		builder.Grow(strLen + strLen/4)
	}

	var pendingUnderscore bool
	runeSlice := []rune(str)
	for index, currentRune := range runeSlice {
		if !unicode.IsLetter(currentRune) && !unicode.IsNumber(currentRune) {
			pendingUnderscore = true
			continue
		}

		if pendingUnderscore || shouldSplit(runeSlice, index) {
			if builder.Len() > 0 {
				builder.WriteByte('_')
			}
			pendingUnderscore = false
		}

		builder.WriteRune(unicode.ToLower(currentRune))
	}
	return builder.String()
}

func ToCamelCase(str string) string {
	var (
		builder   strings.Builder
		nextUpper bool
	)
	builder.Grow(len(str))

	runeSlice := []rune(str)
	for index, currentRune := range runeSlice {
		if !unicode.IsLetter(currentRune) && !unicode.IsNumber(currentRune) {
			nextUpper = true
			continue
		}

		if !nextUpper && shouldSplit(runeSlice, index) {
			nextUpper = true
		}

		shouldLower := true
		if nextUpper {
			if builder.Len() > 0 {
				currentRune = unicode.ToUpper(currentRune)
				shouldLower = false
			}
			nextUpper = false
		}
		if shouldLower {
			currentRune = unicode.ToLower(currentRune)
		}

		builder.WriteRune(currentRune)
	}
	return builder.String()
}

// shouldSplit checks if the current position represents a word boundary that requires splitting/transformation.
func shouldSplit(str []rune, index int) bool {
	currentRune := str[index]
	if !unicode.IsUpper(currentRune) {
		return false
	}

	prevRune := currentRune
	if index > 0 {
		prevRune = str[index-1]
	}

	// Check for split conditions based on casing
	var (
		prevIsLower = prevRune != 0 && unicode.IsLower(prevRune)
		nextIsLower bool
	)

	// Robust lookahead for next rune
	nextIdx := index + utf8.RuneLen(currentRune)
	if nextIdx < len(str) {
		b := str[nextIdx]
		nextIsLower = unicode.IsLower(b)
	}

	return prevIsLower || nextIsLower
}
