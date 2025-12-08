package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestValidateToken(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with correct secret
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: expected %s, got %s", userID, claims.UserID)
	}

	if claims.Username != username {
		t.Errorf("Username mismatch: expected %s, got %s", username, claims.Username)
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with wrong secret
	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("Validation should fail with wrong secret")
	}

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got: %v", err)
	}
}

func TestValidateExpiredToken(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	// Create an expired token
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Validate expired token
	_, err = ValidateToken(tokenString, secret)
	if err == nil {
		t.Error("Validation should fail for expired token")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	secret := "test-secret"

	// Test with completely invalid token
	_, err := ValidateToken("invalid.token.string", secret)
	if err == nil {
		t.Error("Validation should fail for invalid token")
	}

	// Test with empty token
	_, err = ValidateToken("", secret)
	if err == nil {
		t.Error("Validation should fail for empty token")
	}
}

func TestExtractUserID(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Extract user ID without validation
	extractedUserID, err := ExtractUserID(token)
	if err != nil {
		t.Fatalf("ExtractUserID failed: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Extracted UserID mismatch: expected %s, got %s", userID, extractedUserID)
	}
}

func TestExtractUserIDFromInvalidToken(t *testing.T) {
	_, err := ExtractUserID("invalid.token")
	if err == nil {
		t.Error("ExtractUserID should fail for invalid token")
	}
}

func TestTokenExpiration(t *testing.T) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, username, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Check that expiration is set correctly
	if claims.ExpiresAt == nil {
		t.Error("Token should have expiration time")
	}

	expectedExpiry := time.Now().Add(TokenDuration)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 10 seconds difference
	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -10*time.Second || diff > 10*time.Second {
		t.Errorf("Token expiry time is not as expected. Diff: %v", diff)
	}
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateToken(userID, username, secret)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	userID := "user123"
	username := "test@example.com"
	secret := "test-secret"

	token, _ := GenerateToken(userID, username, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token, secret)
	}
}

