// Package server implements the GophKeeper gRPC server.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ar11/gophkeeper/internal/crypto"
	"github.com/ar11/gophkeeper/internal/models"
	"github.com/ar11/gophkeeper/internal/storage"
	pb "github.com/ar11/gophkeeper/pkg/api/proto"
	"github.com/ar11/gophkeeper/pkg/auth"
)

// Server implements the GophKeeper gRPC server.
type Server struct {
	pb.UnimplementedGophKeeperServer
	storage   storage.Storage
	jwtSecret string
}

// NewServer creates a new GophKeeper server instance.
func NewServer(storage storage.Storage, jwtSecret string) *Server {
	return &Server{
		storage:   storage,
		jwtSecret: jwtSecret,
	}
}

// Register handles user registration.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	log.Printf("Register request for user: %s", req.Username)

	// Validate input
	if req.Username == "" || req.Password == "" {
		return &pb.RegisterResponse{
			Message: "username and password are required",
		}, nil
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.storage.CreateUser(req.Username, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return &pb.RegisterResponse{
				Message: "user already exists",
			}, nil
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &pb.RegisterResponse{
		UserId:  user.ID,
		Token:   token,
		Message: "registration successful",
	}, nil
}

// Login handles user authentication.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login request for user: %s", req.Username)

	// Validate input
	if req.Username == "" || req.Password == "" {
		return &pb.LoginResponse{
			Message: "username and password are required",
		}, nil
	}

	// Get user
	user, err := s.storage.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return &pb.LoginResponse{
				Message: "invalid credentials",
			}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if err := crypto.CheckPasswordHash(req.Password, user.PasswordHash); err != nil {
		return &pb.LoginResponse{
			Message: "invalid credentials",
		}, nil
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &pb.LoginResponse{
		UserId:  user.ID,
		Token:   token,
		Message: "login successful",
	}, nil
}

// AddItem handles adding a new data item.
func (s *Server) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("AddItem request for user: %s, type: %s, name: %s", claims.UserID, req.Type, req.Name)

	// Create data item
	item := &models.DataItem{
		UserID:        claims.UserID,
		Type:          protoTypeToModelType(req.Type),
		Name:          req.Name,
		EncryptedData: req.EncryptedData,
		Metadata:      req.Metadata,
	}

	if err := s.storage.CreateItem(item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return &pb.AddItemResponse{
		ItemId:  item.ID,
		Message: "item added successfully",
	}, nil
}

// GetItem handles retrieving a data item.
func (s *Server) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("GetItem request for user: %s, item_id: %s, name: %s", claims.UserID, req.ItemId, req.Name)

	var item *models.DataItem

	// Get by ID or name
	if req.ItemId != "" {
		item, err = s.storage.GetItem(claims.UserID, req.ItemId)
	} else if req.Name != "" {
		item, err = s.storage.GetItemByName(claims.UserID, req.Name)
	} else {
		return nil, errors.New("either item_id or name must be provided")
	}

	if err != nil {
		if errors.Is(err, storage.ErrItemNotFound) {
			return nil, errors.New("item not found")
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &pb.GetItemResponse{
		Item: modelItemToProtoItem(item),
	}, nil
}

// ListItems handles listing all data items.
func (s *Server) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("ListItems request for user: %s, type: %s", claims.UserID, req.Type)

	// List items
	items, err := s.storage.ListItems(claims.UserID, protoTypeToModelType(req.Type))
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Convert to proto
	protoItems := make([]*pb.DataItem, len(items))
	for i, item := range items {
		protoItems[i] = modelItemToProtoItem(item)
	}

	return &pb.ListItemsResponse{
		Items: protoItems,
	}, nil
}

// UpdateItem handles updating a data item.
func (s *Server) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("UpdateItem request for user: %s, item_id: %s", claims.UserID, req.ItemId)

	// Get existing item
	item, err := s.storage.GetItem(claims.UserID, req.ItemId)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// Update fields
	item.EncryptedData = req.EncryptedData
	if req.Metadata != nil {
		item.Metadata = req.Metadata
	}

	// Update in storage
	if err := s.storage.UpdateItem(item); err != nil {
		if errors.Is(err, storage.ErrVersionConflict) {
			return &pb.UpdateItemResponse{
				Message: "version conflict - item was modified by another client",
			}, nil
		}
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return &pb.UpdateItemResponse{
		Message:    "item updated successfully",
		NewVersion: item.Version + 1,
	}, nil
}

// DeleteItem handles deleting a data item.
func (s *Server) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("DeleteItem request for user: %s, item_id: %s", claims.UserID, req.ItemId)

	// Delete item
	if err := s.storage.DeleteItem(claims.UserID, req.ItemId); err != nil {
		return nil, fmt.Errorf("failed to delete item: %w", err)
	}

	return &pb.DeleteItemResponse{
		Message: "item deleted successfully",
	}, nil
}

// Sync handles data synchronization.
func (s *Server) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	// Validate token
	claims, err := auth.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	log.Printf("Sync request for user: %s, last_sync: %d", claims.UserID, req.LastSyncTimestamp)

	// Get items modified after last sync
	lastSync := time.Unix(req.LastSyncTimestamp, 0)
	items, err := s.storage.GetItemsModifiedAfter(claims.UserID, lastSync)
	if err != nil {
		return nil, fmt.Errorf("failed to get modified items: %w", err)
	}

	// Convert to proto
	protoItems := make([]*pb.DataItem, len(items))
	for i, item := range items {
		protoItems[i] = modelItemToProtoItem(item)
	}

	return &pb.SyncResponse{
		Items:         protoItems,
		SyncTimestamp: time.Now().Unix(),
		Conflicts:     nil, // TODO: Implement conflict detection
	}, nil
}

// Helper functions

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

func modelItemToProtoItem(item *models.DataItem) *pb.DataItem {
	return &pb.DataItem{
		Id:            item.ID,
		UserId:        item.UserID,
		Type:          modelTypeToProtoType(item.Type),
		Name:          item.Name,
		EncryptedData: item.EncryptedData,
		Metadata:      item.Metadata,
		CreatedAt:     item.CreatedAt.Unix(),
		UpdatedAt:     item.UpdatedAt.Unix(),
		Version:       item.Version,
		Deleted:       item.Deleted,
	}
}

