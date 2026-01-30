package dbastidor

import "unicode"

func dbNameNormalizer(name string) string {
	nameBuilder := ([]rune(name))[:0]
	for _, character := range name {
		character = unicode.ToLower(character)
		if !unicode.IsLetter(character) && !unicode.IsNumber(character) {
			character = '_'
		}

		nameBuilder = append(nameBuilder, character)
	}

	dbName := "test_" + string(nameBuilder)
	return dbName
}
