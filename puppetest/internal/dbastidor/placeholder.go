package dbastidor

import "strconv"

// PlaceholderStyle renders the bind marker for the index-th argument, counting from one.
type PlaceholderStyle func(index int) string

func QuestionPlaceholder(int) string {
	return "?"
}

func OrdinalPlaceholder(index int) string {
	return "$" + strconv.Itoa(index)
}
