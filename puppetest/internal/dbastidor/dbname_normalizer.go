package dbastidor

import "unicode"

func NormalizeDBName(name string) string {
	nameBuilder := ([]rune(name))[:0]
	for _, character := range name {
		character = unicode.ToLower(character)
		if !unicode.IsLetter(character) && !unicode.IsNumber(character) {
			character = '_'
		}

		nameBuilder = append(nameBuilder, character)
	}

	dbName := string(nameBuilder)
	return dbName
}
