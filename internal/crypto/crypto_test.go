package crypto

import (
	"bytes"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plaintext password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	if err := CheckPasswordHash(password, hash); err != nil {
		t.Errorf("CheckPasswordHash failed for correct password: %v", err)
	}

	// Test incorrect password
	if err := CheckPasswordHash("wrongpassword", hash); err == nil {
		t.Error("CheckPasswordHash should fail for incorrect password")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if len(salt1) != SaltSize {
		t.Errorf("Salt size is %d, expected %d", len(salt1), SaltSize)
	}

	// Generate another salt to ensure randomness
	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("Two generated salts should not be equal")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "testpassword"
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	key := DeriveKey(password, salt)

	if len(key) != KeySize {
		t.Errorf("Key size is %d, expected %d", len(key), KeySize)
	}

	// Same password and salt should produce same key
	key2 := DeriveKey(password, salt)
	if !bytes.Equal(key, key2) {
		t.Error("Same password and salt should produce same key")
	}

	// Different password should produce different key
	key3 := DeriveKey("differentpassword", salt)
	if bytes.Equal(key, key3) {
		t.Error("Different password should produce different key")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("Hello, GophKeeper!")
	password := "testpassword123"

	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	key := DeriveKey(password, salt)

	// Encrypt
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext should not be empty")
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypted data does not match original plaintext")
	}
}

func TestEncryptDecryptWithPassword(t *testing.T) {
	plaintext := []byte("Secret data for testing")
	password := "mypassword"

	// Encrypt
	ciphertext, err := EncryptWithPassword(plaintext, password)
	if err != nil {
		t.Fatalf("EncryptWithPassword failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptWithPassword(ciphertext, password)
	if err != nil {
		t.Fatalf("DecryptWithPassword failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted data does not match original.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

func TestDecryptWithWrongPassword(t *testing.T) {
	plaintext := []byte("Secret data")
	password := "correctpassword"

	ciphertext, err := EncryptWithPassword(plaintext, password)
	if err != nil {
		t.Fatalf("EncryptWithPassword failed: %v", err)
	}

	// Try to decrypt with wrong password
	_, err = DecryptWithPassword(ciphertext, "wrongpassword")
	if err == nil {
		t.Error("Decryption with wrong password should fail")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	password := "testpassword"

	// Too short ciphertext
	_, err := DecryptWithPassword([]byte("short"), password)
	if err != ErrInvalidCiphertext {
		t.Errorf("Expected ErrInvalidCiphertext, got: %v", err)
	}

	// Random invalid data
	_, err = DecryptWithPassword(make([]byte, 100), password)
	if err == nil {
		t.Error("Decryption of random data should fail")
	}
}

func TestEncodeDecodeBase64(t *testing.T) {
	data := []byte("Hello, World!")

	encoded := EncodeBase64(data)
	if encoded == "" {
		t.Error("Encoded string should not be empty")
	}

	decoded, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatalf("DecodeBase64 failed: %v", err)
	}

	if !bytes.Equal(data, decoded) {
		t.Error("Decoded data does not match original")
	}
}

func TestDecodeBase64Invalid(t *testing.T) {
	_, err := DecodeBase64("invalid!@#$%^&*()")
	if err == nil {
		t.Error("Decoding invalid base64 should fail")
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "testpassword123"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkEncryptWithPassword(b *testing.B) {
	plaintext := []byte("Benchmark data for encryption testing")
	password := "benchmarkpassword"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EncryptWithPassword(plaintext, password)
	}
}

func BenchmarkDecryptWithPassword(b *testing.B) {
	plaintext := []byte("Benchmark data for decryption testing")
	password := "benchmarkpassword"

	ciphertext, _ := EncryptWithPassword(plaintext, password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecryptWithPassword(ciphertext, password)
	}
}

