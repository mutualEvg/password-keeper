// Package client implements the GophKeeper client logic.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ar11/gophkeeper/internal/crypto"
	"github.com/ar11/gophkeeper/internal/models"
	pb "github.com/ar11/gophkeeper/pkg/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents a GophKeeper client.
type Client struct {
	conn          *grpc.ClientConn
	grpcClient    pb.GophKeeperClient
	configDir     string
	token         string
	masterPass    string
	serverAddress string
}

// Config represents the client configuration.
type Config struct {
	ServerAddress string `json:"server_address"`
	Token         string `json:"token"`
	LastSync      int64  `json:"last_sync"`
}

var (
	// ErrNotAuthenticated is returned when the user is not authenticated.
	ErrNotAuthenticated = errors.New("not authenticated - please login first")
	// ErrMasterPasswordRequired is returned when master password is required.
	ErrMasterPasswordRequired = errors.New("master password required")
)

// NewClient creates a new GophKeeper client instance.
func NewClient(serverAddress string) (*Client, error) {
	// Get config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".gophkeeper")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	client := &Client{
		configDir:     configDir,
		serverAddress: serverAddress,
	}

	// Load config
	if err := client.loadConfig(); err != nil {
		// Config doesn't exist yet, that's ok
	}

	return client, nil
}

// Connect establishes a connection to the server.
func (c *Client) Connect() error {
	if c.serverAddress == "" {
		return errors.New("server address not configured")
	}

	// TODO: Add TLS credentials in production
	conn, err := grpc.Dial(c.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	c.grpcClient = pb.NewGophKeeperClient(conn)
	return nil
}

// Close closes the connection to the server.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Register registers a new user.
func (c *Client) Register(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.grpcClient.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	if resp.Token == "" {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Save token
	c.token = resp.Token
	c.masterPass = password

	if err := c.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Login authenticates with the server.
func (c *Client) Login(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.grpcClient.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if resp.Token == "" {
		return fmt.Errorf("login failed: %s", resp.Message)
	}

	// Save token
	c.token = resp.Token
	c.masterPass = password

	if err := c.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// AddCredential adds a new credential.
func (c *Client) AddCredential(name, login, password string, metadata map[string]string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	// Create credential data
	cred := models.Credential{
		Login:    login,
		Password: password,
	}

	// Serialize to JSON
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	// Encrypt data
	encrypted, err := crypto.EncryptWithPassword(data, c.masterPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Send to server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.grpcClient.AddItem(ctx, &pb.AddItemRequest{
		Token:         c.token,
		Type:          pb.DataType_DATA_TYPE_CREDENTIAL,
		Name:          name,
		EncryptedData: encrypted,
		Metadata:      metadata,
	})

	if err != nil {
		return fmt.Errorf("failed to add credential: %w", err)
	}

	return nil
}

// AddText adds new text data.
func (c *Client) AddText(name, content string, metadata map[string]string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	// Create text data
	textData := models.TextData{
		Content: content,
	}

	// Serialize to JSON
	data, err := json.Marshal(textData)
	if err != nil {
		return fmt.Errorf("failed to marshal text: %w", err)
	}

	// Encrypt data
	encrypted, err := crypto.EncryptWithPassword(data, c.masterPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Send to server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.grpcClient.AddItem(ctx, &pb.AddItemRequest{
		Token:         c.token,
		Type:          pb.DataType_DATA_TYPE_TEXT,
		Name:          name,
		EncryptedData: encrypted,
		Metadata:      metadata,
	})

	if err != nil {
		return fmt.Errorf("failed to add text: %w", err)
	}

	return nil
}

// AddBinary adds binary data.
func (c *Client) AddBinary(name, filename string, data []byte, metadata map[string]string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	// Create binary data
	binaryData := models.BinaryData{
		Filename: filename,
		Data:     data,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(binaryData)
	if err != nil {
		return fmt.Errorf("failed to marshal binary data: %w", err)
	}

	// Encrypt data
	encrypted, err := crypto.EncryptWithPassword(jsonData, c.masterPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Send to server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.grpcClient.AddItem(ctx, &pb.AddItemRequest{
		Token:         c.token,
		Type:          pb.DataType_DATA_TYPE_BINARY,
		Name:          name,
		EncryptedData: encrypted,
		Metadata:      metadata,
	})

	if err != nil {
		return fmt.Errorf("failed to add binary data: %w", err)
	}

	return nil
}

// AddCard adds bank card data.
func (c *Client) AddCard(name string, card models.CardData, metadata map[string]string) error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	// Serialize to JSON
	data, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %w", err)
	}

	// Encrypt data
	encrypted, err := crypto.EncryptWithPassword(data, c.masterPass)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Send to server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.grpcClient.AddItem(ctx, &pb.AddItemRequest{
		Token:         c.token,
		Type:          pb.DataType_DATA_TYPE_CARD,
		Name:          name,
		EncryptedData: encrypted,
		Metadata:      metadata,
	})

	if err != nil {
		return fmt.Errorf("failed to add card: %w", err)
	}

	return nil
}

// GetItem retrieves an item by name.
func (c *Client) GetItem(name string) (*models.DataItem, interface{}, error) {
	if c.token == "" {
		return nil, nil, ErrNotAuthenticated
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.grpcClient.GetItem(ctx, &pb.GetItemRequest{
		Token: c.token,
		Name:  name,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Convert proto to model
	item := protoToModelItem(resp.Item)

	// Decrypt data
	decrypted, err := crypto.DecryptWithPassword(item.EncryptedData, c.masterPass)
	if err != nil {
		return item, nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Unmarshal based on type
	var data interface{}
	switch item.Type {
	case models.DataTypeCredential:
		var cred models.Credential
		if err := json.Unmarshal(decrypted, &cred); err != nil {
			return item, nil, err
		}
		data = cred
	case models.DataTypeText:
		var text models.TextData
		if err := json.Unmarshal(decrypted, &text); err != nil {
			return item, nil, err
		}
		data = text
	case models.DataTypeBinary:
		var binary models.BinaryData
		if err := json.Unmarshal(decrypted, &binary); err != nil {
			return item, nil, err
		}
		data = binary
	case models.DataTypeCard:
		var card models.CardData
		if err := json.Unmarshal(decrypted, &card); err != nil {
			return item, nil, err
		}
		data = card
	}

	return item, data, nil
}

// ListItems lists all items.
func (c *Client) ListItems(dataType models.DataType) ([]*models.DataItem, error) {
	if c.token == "" {
		return nil, ErrNotAuthenticated
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.grpcClient.ListItems(ctx, &pb.ListItemsRequest{
		Token: c.token,
		Type:  modelTypeToProtoType(dataType),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Convert proto items to model items
	items := make([]*models.DataItem, len(resp.Items))
	for i, protoItem := range resp.Items {
		items[i] = protoToModelItem(protoItem)
	}

	return items, nil
}

// Sync synchronizes data with the server.
func (c *Client) Sync() error {
	if c.token == "" {
		return ErrNotAuthenticated
	}

	// Load last sync time
	config, _ := c.loadConfigData()
	lastSync := config.LastSync

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.grpcClient.Sync(ctx, &pb.SyncRequest{
		Token:             c.token,
		LastSyncTimestamp: lastSync,
	})

	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Update last sync time
	config.LastSync = resp.SyncTimestamp
	c.saveConfigData(config)

	return nil
}

// SetMasterPassword sets the master password for encryption.
func (c *Client) SetMasterPassword(password string) {
	c.masterPass = password
}

// Config management

func (c *Client) loadConfig() error {
	config, err := c.loadConfigData()
	if err != nil {
		return err
	}

	c.token = config.Token
	if c.serverAddress == "" {
		c.serverAddress = config.ServerAddress
	}

	return nil
}

func (c *Client) loadConfigData() (*Config, error) {
	configPath := filepath.Join(c.configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &Config{}, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return &Config{}, err
	}

	return &config, nil
}

func (c *Client) saveConfig() error {
	config := &Config{
		ServerAddress: c.serverAddress,
		Token:         c.token,
	}

	return c.saveConfigData(config)
}

func (c *Client) saveConfigData(config *Config) error {
	configPath := filepath.Join(c.configDir, "config.json")
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// Helper functions

func modelTypeToProtoType(mt models.DataType) pb.DataType {
	switch mt {
	case models.DataTypeCredential:
		return pb.DataType_DATA_TYPE_CREDENTIAL
	case models.DataTypeText:
		return pb.DataType_DATA_TYPE_TEXT
	case models.DataTypeBinary:
		return pb.DataType_DATA_TYPE_BINARY
	case models.DataTypeCard:
		return pb.DataType_DATA_TYPE_CARD
	default:
		return pb.DataType_DATA_TYPE_UNSPECIFIED
	}
}

func protoTypeToModelType(pt pb.DataType) models.DataType {
	switch pt {
	case pb.DataType_DATA_TYPE_CREDENTIAL:
		return models.DataTypeCredential
	case pb.DataType_DATA_TYPE_TEXT:
		return models.DataTypeText
	case pb.DataType_DATA_TYPE_BINARY:
		return models.DataTypeBinary
	case pb.DataType_DATA_TYPE_CARD:
		return models.DataTypeCard
	default:
		return ""
	}
}

func protoToModelItem(item *pb.DataItem) *models.DataItem {
	return &models.DataItem{
		ID:            item.Id,
		UserID:        item.UserId,
		Type:          protoTypeToModelType(item.Type),
		Name:          item.Name,
		EncryptedData: item.EncryptedData,
		Metadata:      item.Metadata,
		CreatedAt:     time.Unix(item.CreatedAt, 0),
		UpdatedAt:     time.Unix(item.UpdatedAt, 0),
		Version:       item.Version,
		Deleted:       item.Deleted,
	}
}

