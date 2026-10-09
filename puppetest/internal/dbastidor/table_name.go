package dbastidor

import (
	"reflect"
	"strings"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/strnormalizer"
)

type tableNamer interface {
	TableName() string
}

func tableNameOf(item any, typ reflect.Type) string {
	if named, isNamed := item.(tableNamer); isNamed {
		return named.TableName()
	}

	return pluralize(strnormalizer.ToSnakeCase(typ.Name()))
}

func pluralize(noun string) string {
	switch {
	case strings.HasSuffix(noun, "x"):
		return noun
	case len(noun) > 1 && strings.HasSuffix(noun, "y") && !strings.ContainsRune("aeiou", rune(noun[len(noun)-2])):
		return strings.TrimSuffix(noun, "y") + "ies"
	default:
		return noun + "s"
	}
}
