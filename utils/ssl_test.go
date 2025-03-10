package utils

import (
	"os"
	"testing"
)

// Test ConfigureSecureHTTPClientWithFile with a valid certificate
func TestConfigureSecureHTTPClientWithFile_ValidCert(t *testing.T) {
	client, err := ConfigureSecureHTTPClientWithFile("../assets/cacert.pem")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err) // Fail immediately on error
	}

	if client == nil {
		t.Fatal("Expected a client but got nil")
	}
}

// Test ConfigureSecureHTTPClientWithFile with a non-existent file
func TestConfigureSecureHTTPClientWithFile_NonExistentFile(t *testing.T) {
	_, err := ConfigureSecureHTTPClientWithFile("nonexistent.pem")
	if err == nil {
		t.Error("Expected an error for a non-existent file, but no error occurred")
	}
}

// Test ConfigureSecureHTTPClientWithFile with an empty file
func TestConfigureSecureHTTPClientWithFile_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-cert-empty-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close() // Ensure the file is empty

	_, err = ConfigureSecureHTTPClientWithFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected an error when passing an empty certificate file, but no error occurred")
	}
}

// Test ConfigureSecureHTTPClientWithFile with an empty path
func TestConfigureSecureHTTPClientWithFile_EmptyPath(t *testing.T) {
	client, err := ConfigureSecureHTTPClientWithFile("")
	if err != nil {
		t.Errorf("Expected a nil client without error, but got an error: %v", err)
	}
	if client != nil {
		t.Error("Expected a nil client when passing an empty path, but got a valid client")
	}
}