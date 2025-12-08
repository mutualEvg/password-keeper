// Package models defines the data structures used in GophKeeper.
package models

import (
	"time"
)

// User represents a registered user in the system.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DataType represents the type of stored data.
type DataType string

const (
	// DataTypeCredential represents login/password pairs.
	DataTypeCredential DataType = "credential"
	// DataTypeText represents arbitrary text data.
	DataTypeText DataType = "text"
	// DataTypeBinary represents arbitrary binary data.
	DataTypeBinary DataType = "binary"
	// DataTypeCard represents bank card information.
	DataTypeCard DataType = "card"
)

// DataItem represents a stored data item.
type DataItem struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Type          DataType          `json:"type"`
	Name          string            `json:"name"`
	EncryptedData []byte            `json:"encrypted_data"`
	Metadata      map[string]string `json:"metadata"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Version       int64             `json:"version"`
	Deleted       bool              `json:"deleted"`
}

// Credential represents a login/password pair (before encryption).
type Credential struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// TextData represents arbitrary text data (before encryption).
type TextData struct {
	Content string `json:"content"`
}

// BinaryData represents arbitrary binary data (before encryption).
type BinaryData struct {
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

// CardData represents bank card information (before encryption).
type CardData struct {
	Number      string `json:"number"`
	Holder      string `json:"holder"`
	CVV         string `json:"cvv"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
}

// SyncConflict represents a synchronization conflict between client and server.
type SyncConflict struct {
	ItemID        string    `json:"item_id"`
	ServerVersion DataItem  `json:"server_version"`
	ClientVersion DataItem  `json:"client_version"`
}

// Config represents the application configuration.
type Config struct {
	ServerAddress string `json:"server_address"`
	DBHost        string `json:"db_host"`
	DBPort        string `json:"db_port"`
	DBUser        string `json:"db_user"`
	DBPassword    string `json:"db_password"`
	DBName        string `json:"db_name"`
	JWTSecret     string `json:"jwt_secret"`
}

