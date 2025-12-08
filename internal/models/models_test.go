package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserSerialization(t *testing.T) {
	user := User{
		ID:           "user123",
		Username:     "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Serialize to JSON
	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	// Deserialize from JSON
	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal user: %v", err)
	}

	// Verify fields
	if decoded.ID != user.ID {
		t.Errorf("ID mismatch: expected %s, got %s", user.ID, decoded.ID)
	}

	if decoded.Username != user.Username {
		t.Errorf("Username mismatch: expected %s, got %s", user.Username, decoded.Username)
	}

	// PasswordHash should not be in JSON due to json:"-" tag
	if string(data) != "" {
		jsonStr := string(data)
		if contains(jsonStr, "password_hash") || contains(jsonStr, "hashed_password") {
			t.Error("PasswordHash should not be serialized to JSON")
		}
	}
}

func TestDataItemSerialization(t *testing.T) {
	metadata := map[string]string{
		"website": "example.com",
		"notes":   "test notes",
	}

	item := DataItem{
		ID:            "item123",
		UserID:        "user123",
		Type:          DataTypeCredential,
		Name:          "Test Item",
		EncryptedData: []byte("encrypted_data"),
		Metadata:      metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       1,
		Deleted:       false,
	}

	// Serialize to JSON
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal data item: %v", err)
	}

	// Deserialize from JSON
	var decoded DataItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal data item: %v", err)
	}

	// Verify fields
	if decoded.ID != item.ID {
		t.Errorf("ID mismatch: expected %s, got %s", item.ID, decoded.ID)
	}

	if decoded.Type != item.Type {
		t.Errorf("Type mismatch: expected %s, got %s", item.Type, decoded.Type)
	}

	if decoded.Metadata["website"] != metadata["website"] {
		t.Error("Metadata not properly serialized")
	}
}

func TestCredentialSerialization(t *testing.T) {
	cred := Credential{
		Login:    "user@example.com",
		Password: "secretpass",
	}

	// Serialize to JSON
	data, err := json.Marshal(cred)
	if err != nil {
		t.Fatalf("Failed to marshal credential: %v", err)
	}

	// Deserialize from JSON
	var decoded Credential
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal credential: %v", err)
	}

	if decoded.Login != cred.Login {
		t.Errorf("Login mismatch: expected %s, got %s", cred.Login, decoded.Login)
	}

	if decoded.Password != cred.Password {
		t.Errorf("Password mismatch: expected %s, got %s", cred.Password, decoded.Password)
	}
}

func TestTextDataSerialization(t *testing.T) {
	text := TextData{
		Content: "Secret notes",
	}

	// Serialize to JSON
	data, err := json.Marshal(text)
	if err != nil {
		t.Fatalf("Failed to marshal text data: %v", err)
	}

	// Deserialize from JSON
	var decoded TextData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal text data: %v", err)
	}

	if decoded.Content != text.Content {
		t.Errorf("Content mismatch: expected %s, got %s", text.Content, decoded.Content)
	}
}

func TestBinaryDataSerialization(t *testing.T) {
	binary := BinaryData{
		Filename: "test.bin",
		Data:     []byte{0x01, 0x02, 0x03, 0x04},
	}

	// Serialize to JSON
	data, err := json.Marshal(binary)
	if err != nil {
		t.Fatalf("Failed to marshal binary data: %v", err)
	}

	// Deserialize from JSON
	var decoded BinaryData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal binary data: %v", err)
	}

	if decoded.Filename != binary.Filename {
		t.Errorf("Filename mismatch: expected %s, got %s", binary.Filename, decoded.Filename)
	}

	if len(decoded.Data) != len(binary.Data) {
		t.Errorf("Data length mismatch: expected %d, got %d", len(binary.Data), len(decoded.Data))
	}
}

func TestCardDataSerialization(t *testing.T) {
	card := CardData{
		Number:      "4111111111111111",
		Holder:      "John Doe",
		CVV:         "123",
		ExpiryMonth: "12",
		ExpiryYear:  "25",
	}

	// Serialize to JSON
	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card data: %v", err)
	}

	// Deserialize from JSON
	var decoded CardData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal card data: %v", err)
	}

	if decoded.Number != card.Number {
		t.Errorf("Number mismatch: expected %s, got %s", card.Number, decoded.Number)
	}

	if decoded.Holder != card.Holder {
		t.Errorf("Holder mismatch: expected %s, got %s", card.Holder, decoded.Holder)
	}

	if decoded.CVV != card.CVV {
		t.Errorf("CVV mismatch: expected %s, got %s", card.CVV, decoded.CVV)
	}
}

func TestDataTypeConstants(t *testing.T) {
	types := []DataType{
		DataTypeCredential,
		DataTypeText,
		DataTypeBinary,
		DataTypeCard,
	}

	for _, dt := range types {
		if dt == "" {
			t.Error("DataType constant should not be empty")
		}
	}

	// Test uniqueness
	typeMap := make(map[DataType]bool)
	for _, dt := range types {
		if typeMap[dt] {
			t.Errorf("Duplicate DataType constant: %s", dt)
		}
		typeMap[dt] = true
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsMiddle(s, substr))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

