package code

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"

	// sqlite driver
	_ "modernc.org/sqlite"

	"github.com/nullswan/nomi/internal/migrations"
)

// TODO(nullswan): Add sqlc here

type Repository interface {
	SaveCodeBlock(block CodeBlock) error
	LoadCodeBlock(id string) (CodeBlock, error)

	LoadCodeBlocks() ([]CodeBlock, error)

	Close() error
}

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("error creating sqlite driver: %w", err)
	}

	migrations, err := migrations.GetMigrations()
	if err != nil {
		return nil, fmt.Errorf("error getting migrations: %w", err)
	}

	sourceDriver, err := iofs.New(migrations, ".")
	if err != nil {
		return nil, fmt.Errorf("error creating source driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"sqlite",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	return &sqliteRepository{db: db}, nil
}

func (r *sqliteRepository) SaveCodeBlock(block CodeBlock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	insertCodeBlock := `
		INSERT OR REPLACE INTO code_snippets (id, created_at, description, code, language)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(
		insertCodeBlock,
		uuid.New().String(),
		time.Now().UTC(),
		block.Description,
		block.Code,
		block.Language,
	)
	if err != nil {
		return fmt.Errorf("error inserting code block: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (r *sqliteRepository) LoadCodeBlock(id string) (CodeBlock, error) {
	query := `
		SELECT id, description, code, language
		FROM code_snippets
		WHERE id = ?
	`
	row := r.db.QueryRow(query, id)

	var block CodeBlock
	err := row.Scan(
		&block.ID,
		&block.Description,
		&block.Code,
		&block.Language,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return block, fmt.Errorf("code block not found: %w", err)
		}
		return block, fmt.Errorf("error querying code block: %w", err)
	}

	return block, nil
}

func (r *sqliteRepository) LoadCodeBlocks() ([]CodeBlock, error) {
	query := `
		SELECT id, description, code, language
		FROM code_snippets
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying code blocks: %w", err)
	}
	defer rows.Close()

	var blocks []CodeBlock
	for rows.Next() {
		var block CodeBlock
		err := rows.Scan(
			&block.ID,
			&block.Description,
			&block.Code,
			&block.Language,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning code block: %w", err)
		}

		blocks = append(blocks, block)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return blocks, nil
}

func (r *sqliteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	return nil
}
