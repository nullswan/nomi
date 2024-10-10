package migrations

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed *.sql
var migrationsFS embed.FS

func GetMigrations() (fs.FS, error) {
	filesys, err := fs.Sub(migrationsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("error getting migrations: %w", err)
	}

	return filesys, nil
}
