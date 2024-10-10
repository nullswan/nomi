package chat

// TODO(nullswan): Use DDD here
// TODO(nullswan): Add sqlc here

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/nullswan/golem/internal/migrations"
)

type Repository interface {
	SaveConversation(conversation Conversation) error
	LoadConversation(id string) (Conversation, error)
	DeleteConversation(id string) error

	GetConversations() ([]Conversation, error)

	Close() error
}

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
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

func (r *sqliteRepository) SaveConversation(
	conversation Conversation,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert or ignore the conversation
	insertConversation := `INSERT OR IGNORE INTO conversations (id, created_at) VALUES (?, ?)`
	_, err = tx.Exec(
		insertConversation,
		conversation.GetId(),
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("error inserting conversation: %w", err)
	}

	// Insert messages
	insertMessage := `INSERT OR IGNORE INTO messages (id, conversation_id, role, content, created_at, is_file) VALUES (?, ?, ?, ?, ?)`
	for _, msg := range conversation.GetMessages() {
		_, err = tx.Exec(
			insertMessage,
			msg.ID,
			conversation.GetId(),
			msg.Role,
			msg.Content,
			msg.CreatedAt,
			msg.IsFile,
		)
		if err != nil {
			return fmt.Errorf("error inserting message: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (r *sqliteRepository) LoadConversation(
	id string,
) (Conversation, error) {
	queryConversation := `SELECT id, created_at FROM conversations WHERE id = ?`
	row := r.db.QueryRow(queryConversation, id)

	var convoID string
	var convoCreatedAt time.Time
	err := row.Scan(&convoID, &convoCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error scanning conversation: %w", err)
	}

	queryMessages := `SELECT id, role, content, created_at, is_file FROM messages WHERE conversation_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(queryMessages, id)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
			&msg.IsFile,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		msg.CreatedAt = msg.CreatedAt.UTC()
		messages = append(messages, msg)
	}

	return &stackedConversation{
		repo:      r,
		id:        convoID,
		messages:  messages,
		createdAt: convoCreatedAt.UTC(),
	}, nil
}

func (r *sqliteRepository) DeleteConversation(id string) error {
	// This will cascade delete messages
	deleteConversation := `DELETE FROM conversations WHERE id = ?`
	_, err := r.db.Exec(deleteConversation, id)
	if err != nil {
		return fmt.Errorf("error deleting conversation: %w", err)
	}

	return nil
}

func (r *sqliteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	return nil
}

func (r *sqliteRepository) Reset() error {
	_, err := r.db.Exec("DELETE FROM conversations")
	if err != nil {
		return fmt.Errorf("error deleting conversations: %w", err)
	}

	_, err = r.db.Exec("DELETE FROM messages")
	if err != nil {
		return fmt.Errorf("error deleting messages: %w", err)
	}

	return nil
}

func (r *sqliteRepository) GetConversations() ([]Conversation, error) {
	queryConversations := `SELECT id FROM conversations`
	rows, err := r.db.Query(queryConversations)
	if err != nil {
		return nil, fmt.Errorf("error getting conversations: %w", err)
	}
	defer rows.Close()

	var convos []Conversation
	for rows.Next() {
		var convoID string
		err := rows.Scan(&convoID)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		convo, err := r.LoadConversation(convoID)
		if err != nil {
			return nil, fmt.Errorf("error loading conversation: %w", err)
		}

		convos = append(convos, convo)
	}

	return convos, nil
}
