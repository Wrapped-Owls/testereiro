package puppetest

import (
	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

// PlaceholderStyle renders the bind marker for the index-th argument, counting from one.
type PlaceholderStyle = dbastidor.PlaceholderStyle

// QuestionPlaceholder is the default, used by MySQL and SQLite.
func QuestionPlaceholder(index int) string {
	return dbastidor.QuestionPlaceholder(index)
}

// OrdinalPlaceholder renders $1, $2, ... for Postgres.
func OrdinalPlaceholder(index int) string {
	return dbastidor.OrdinalPlaceholder(index)
}
