package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var sqlFileSystem embed.FS

func FS() fs.FS {
	return sqlFileSystem
}
