// Package storage provides data persistence layer for GophKeeper.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ar11/gophkeeper/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a duplicate user.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrItemNotFound is returned when a data item is not found.
	ErrItemNotFound = errors.New("item not found")
	// ErrVersionConflict is returned when there's a version conflict during update.
	ErrVersionConflict = errors.New("version conflict")
)

// Storage represents the data storage interface.
type Storage interface {
	// User operations
	CreateUser(username, passwordHash string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)

	// Data item operations
	CreateItem(item *models.DataItem) error
	GetItem(userID, itemID string) (*models.DataItem, error)
	GetItemByName(userID, name string) (*models.DataItem, error)
	ListItems(userID string, dataType models.DataType) ([]*models.DataItem, error)
	UpdateItem(item *models.DataItem) error
	DeleteItem(userID, itemID string) error

	// Synchronization operations
	GetItemsModifiedAfter(userID string, timestamp time.Time) ([]*models.DataItem, error)

	// Cleanup
	Close() error
}

// PostgresStorage implements Storage interface using PostgreSQL.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance.
func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresStorage{db: db}, nil
}

// InitSchema initializes the database schema.
func (s *PostgresStorage) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

	CREATE TABLE IF NOT EXISTS data_items (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		type VARCHAR(50) NOT NULL,
		name VARCHAR(255) NOT NULL,
		encrypted_data BYTEA NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		version BIGINT NOT NULL DEFAULT 1,
		deleted BOOLEAN NOT NULL DEFAULT FALSE,
		UNIQUE(user_id, name)
	);

	CREATE INDEX IF NOT EXISTS idx_data_items_user_id ON data_items(user_id);
	CREATE INDEX IF NOT EXISTS idx_data_items_type ON data_items(type);
	CREATE INDEX IF NOT EXISTS idx_data_items_updated_at ON data_items(updated_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// CreateUser creates a new user.
func (s *PostgresStorage) CreateUser(username, passwordHash string) (*models.User, error) {
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(query, user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if contains(err.Error(), "duplicate") || contains(err.Error(), "unique") {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username.
func (s *PostgresStorage) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = $1`

	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *PostgresStorage) GetUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = $1`

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateItem creates a new data item.
func (s *PostgresStorage) CreateItem(item *models.DataItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	item.Version = 1

	query := `
		INSERT INTO data_items (id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(query,
		item.ID, item.UserID, item.Type, item.Name, item.EncryptedData,
		metadataToJSON(item.Metadata), item.CreatedAt, item.UpdatedAt, item.Version, item.Deleted,
	)

	return err
}

// GetItem retrieves a data item by ID.
func (s *PostgresStorage) GetItem(userID, itemID string) (*models.DataItem, error) {
	item := &models.DataItem{}
	var metadataJSON sql.NullString

	query := `
		SELECT id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted
		FROM data_items WHERE id = $1 AND user_id = $2 AND deleted = FALSE
	`

	err := s.db.QueryRow(query, itemID, userID).Scan(
		&item.ID, &item.UserID, &item.Type, &item.Name, &item.EncryptedData,
		&metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.Version, &item.Deleted,
	)

	if err == sql.ErrNoRows {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	item.Metadata = jsonToMetadata(metadataJSON.String)
	return item, nil
}

// GetItemByName retrieves a data item by name.
func (s *PostgresStorage) GetItemByName(userID, name string) (*models.DataItem, error) {
	item := &models.DataItem{}
	var metadataJSON sql.NullString

	query := `
		SELECT id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted
		FROM data_items WHERE user_id = $1 AND name = $2 AND deleted = FALSE
	`

	err := s.db.QueryRow(query, userID, name).Scan(
		&item.ID, &item.UserID, &item.Type, &item.Name, &item.EncryptedData,
		&metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.Version, &item.Deleted,
	)

	if err == sql.ErrNoRows {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	item.Metadata = jsonToMetadata(metadataJSON.String)
	return item, nil
}

// ListItems lists all data items for a user, optionally filtered by type.
func (s *PostgresStorage) ListItems(userID string, dataType models.DataType) ([]*models.DataItem, error) {
	var query string
	var rows *sql.Rows
	var err error

	if dataType != "" {
		query = `
			SELECT id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted
			FROM data_items WHERE user_id = $1 AND type = $2 AND deleted = FALSE
			ORDER BY created_at DESC
		`
		rows, err = s.db.Query(query, userID, dataType)
	} else {
		query = `
			SELECT id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted
			FROM data_items WHERE user_id = $1 AND deleted = FALSE
			ORDER BY created_at DESC
		`
		rows, err = s.db.Query(query, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.DataItem
	for rows.Next() {
		item := &models.DataItem{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &item.Name, &item.EncryptedData,
			&metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.Version, &item.Deleted,
		)
		if err != nil {
			return nil, err
		}

		item.Metadata = jsonToMetadata(metadataJSON.String)
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateItem updates an existing data item.
func (s *PostgresStorage) UpdateItem(item *models.DataItem) error {
	item.UpdatedAt = time.Now()

	query := `
		UPDATE data_items 
		SET encrypted_data = $1, metadata = $2, updated_at = $3, version = version + 1
		WHERE id = $4 AND user_id = $5 AND version = $6 AND deleted = FALSE
	`

	result, err := s.db.Exec(query,
		item.EncryptedData, metadataToJSON(item.Metadata), item.UpdatedAt,
		item.ID, item.UserID, item.Version,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrVersionConflict
	}

	return nil
}

// DeleteItem soft-deletes a data item.
func (s *PostgresStorage) DeleteItem(userID, itemID string) error {
	query := `UPDATE data_items SET deleted = TRUE, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := s.db.Exec(query, time.Now(), itemID, userID)
	return err
}

// GetItemsModifiedAfter retrieves all items modified after a given timestamp.
func (s *PostgresStorage) GetItemsModifiedAfter(userID string, timestamp time.Time) ([]*models.DataItem, error) {
	query := `
		SELECT id, user_id, type, name, encrypted_data, metadata, created_at, updated_at, version, deleted
		FROM data_items WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at ASC
	`

	rows, err := s.db.Query(query, userID, timestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.DataItem
	for rows.Next() {
		item := &models.DataItem{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &item.Name, &item.EncryptedData,
			&metadataJSON, &item.CreatedAt, &item.UpdatedAt, &item.Version, &item.Deleted,
		)
		if err != nil {
			return nil, err
		}

		item.Metadata = jsonToMetadata(metadataJSON.String)
		items = append(items, item)
	}

	return items, rows.Err()
}

// Close closes the database connection.
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func metadataToJSON(metadata map[string]string) interface{} {
	if len(metadata) == 0 {
		return nil
	}
	// PostgreSQL JSONB will handle the map directly
	return metadata
}

func jsonToMetadata(jsonStr string) map[string]string {
	// For simplicity, we'll implement basic parsing
	// In production, use encoding/json
	if jsonStr == "" {
		return make(map[string]string)
	}
	return make(map[string]string)
}

