package dbastidor

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidIdentifier = errors.New("invalid sql identifier")

// QuoteIdentifier doubles the quote rune so a crafted name cannot close the identifier and append
// its own statement; NormalizeDBName normalizes, which is not the same as escaping. Assumes a
// symmetric delimiter, so asymmetric brackets need their own escaping.
func QuoteIdentifier(name string, quote rune) (string, error) {
	if name == "" {
		return "", fmt.Errorf("%w: empty name", ErrInvalidIdentifier)
	}
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("%w: name holds a null byte", ErrInvalidIdentifier)
	}

	delimiter := string(quote)
	return delimiter + strings.ReplaceAll(name, delimiter, delimiter+delimiter) + delimiter, nil
}
