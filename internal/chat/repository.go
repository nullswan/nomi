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
	_ "github.com/mattn/go-sqlite3"
	"github.com/nullswan/golem/internal/migrations"
)

type Repository interface {
	SaveConversation(conversation Conversation) error
	LoadConversation(id string) (Conversation, error)
	GetConversations() ([]Conversation, error)
	Close() error
}

type sqliteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, err
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
		return nil, err
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
		return err
	}

	// Insert messages
	insertMessage := `INSERT OR IGNORE INTO messages (id, conversation_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`
	for _, msg := range conversation.GetMessages() {
		_, err = tx.Exec(
			insertMessage,
			msg.Id,
			conversation.GetId(),
			msg.Role,
			msg.Content,
			msg.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteRepository) LoadConversation(
	id string,
) (Conversation, error) {
	queryConversation := `SELECT id, created_at FROM conversations WHERE id = ?`
	row := r.db.QueryRow(queryConversation, id)

	var convoId string
	var convoCreatedAt time.Time
	err := row.Scan(&convoId, &convoCreatedAt)
	if err != nil {
		return nil, err
	}

	queryMessages := `SELECT id, role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(queryMessages, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Id, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		msg.CreatedAt = msg.CreatedAt.UTC()
		messages = append(messages, msg)
	}

	return &stackedConversation{
		repo:      r,
		id:        convoId,
		messages:  messages,
		createdAt: convoCreatedAt.UTC(),
	}, nil
}

func (r *sqliteRepository) Close() error {
	return r.db.Close()
}

func (r *sqliteRepository) Reset() error {
	_, err := r.db.Exec("DELETE FROM conversations")
	if err != nil {
		return err
	}

	_, err = r.db.Exec("DELETE FROM messages")
	if err != nil {
		return err
	}

	return nil
}

func (r *sqliteRepository) GetConversations() ([]Conversation, error) {
	queryConversations := `SELECT id FROM conversations`
	rows, err := r.db.Query(queryConversations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convos []Conversation
	for rows.Next() {
		var convoId string
		err := rows.Scan(&convoId)
		if err != nil {
			return nil, err
		}

		convo, err := r.LoadConversation(convoId)
		if err != nil {
			return nil, err
		}

		convos = append(convos, convo)
	}

	return convos, nil
}
