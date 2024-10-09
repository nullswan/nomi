package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var migrationsFS embed.FS

func GetMigrations() (fs.FS, error) {
	return fs.Sub(migrationsFS, ".")
}
