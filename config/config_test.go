package config

import (
	"os"
	"testing"
)

// MockEnvLoader implements EnvLoader for testing
type MockEnvLoader struct {
	values    map[string]string
	loadError error
}

func NewMockEnvLoader() *MockEnvLoader {
	return &MockEnvLoader{
		values: make(map[string]string),
	}
}

func (m *MockEnvLoader) Load() error {
	return m.loadError
}

func (m *MockEnvLoader) GetEnv(key string) string {
	return m.values[key]
}

func (m *MockEnvLoader) SetEnv(key, value string) {
	m.values[key] = value
}

func (m *MockEnvLoader) SetLoadError(err error) {
	m.loadError = err
}

func TestLoadEnv(t *testing.T) {
	// Create a temporary .env file for testing
	content := []byte(`
NOTION_TOKEN=test_token
NOTION_DATABASE_ID=test_database_id
KOBO_DB_PATH=/path/to/kobo.db
CERT_PATH=/path/to/cert
`)
	err := os.WriteFile(".env.test", content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Rename the original .env file if it exists
	_, err = os.Stat(".env")
	originalExists := !os.IsNotExist(err)
	if originalExists {
		err = os.Rename(".env", ".env.backup")
		if err != nil {
			t.Fatalf("Failed to backup original .env file: %v", err)
		}
	}

	// Move test file to .env
	err = os.Rename(".env.test", ".env")
	if err != nil {
		t.Fatalf("Failed to move test .env file: %v", err)
	}

	// Test LoadEnv function
	err = LoadEnv()
	if err != nil {
		t.Errorf("LoadEnv() error = %v, want nil", err)
	}

	// Clean up and restore original .env if it existed
	os.Remove(".env")
	if originalExists {
		os.Rename(".env.backup", ".env")
	}
}

func TestGetConfigWithLoader(t *testing.T) {
	// Test with all required values
	t.Run("All values present", func(t *testing.T) {
		mock := NewMockEnvLoader()
		mock.SetEnv("NOTION_TOKEN", "test_token")
		mock.SetEnv("NOTION_DATABASE_ID", "test_database_id")
		mock.SetEnv("KOBO_DB_PATH", "/path/to/kobo.db")
		mock.SetEnv("CERT_PATH", "/path/to/cert")

		config, err := GetConfigWithLoader(mock)

		if err != nil {
			t.Errorf("GetConfigWithLoader() error = %v, want nil", err)
		}

		if config.NotionToken != "test_token" {
			t.Errorf("config.NotionToken = %v, want %v", config.NotionToken, "test_token")
		}
		if config.DatabaseID != "test_database_id" {
			t.Errorf("config.DatabaseID = %v, want %v", config.DatabaseID, "test_database_id")
		}
		if config.DBPath != "/path/to/kobo.db" {
			t.Errorf("config.DBPath = %v, want %v", config.DBPath, "/path/to/kobo.db")
		}
		if config.CertPath != "/path/to/cert" {
			t.Errorf("config.CertPath = %v, want %v", config.CertPath, "/path/to/cert")
		}
	})

	// Test with missing values
	t.Run("Missing values", func(t *testing.T) {
		mock := NewMockEnvLoader()
		// Not setting required values

		_, err := GetConfigWithLoader(mock)

		if err == nil {
			t.Error("GetConfigWithLoader() error = nil, want error for missing env vars")
		}
	})

	// Test with some values missing
	t.Run("Some values missing", func(t *testing.T) {
		tests := []struct {
			name      string
			setupMock func(*MockEnvLoader)
			wantErr   bool
		}{
			{
				name: "Missing NOTION_TOKEN",
				setupMock: func(m *MockEnvLoader) {
					m.SetEnv("NOTION_DATABASE_ID", "test_database_id")
					m.SetEnv("KOBO_DB_PATH", "/path/to/kobo.db")
					m.SetEnv("CERT_PATH", "/path/to/cert")
				},
				wantErr: true,
			},
			{
				name: "Missing NOTION_DATABASE_ID",
				setupMock: func(m *MockEnvLoader) {
					m.SetEnv("NOTION_TOKEN", "test_token")
					m.SetEnv("KOBO_DB_PATH", "/path/to/kobo.db")
					m.SetEnv("CERT_PATH", "/path/to/cert")
				},
				wantErr: true,
			},
			{
				name: "Missing KOBO_DB_PATH",
				setupMock: func(m *MockEnvLoader) {
					m.SetEnv("NOTION_TOKEN", "test_token")
					m.SetEnv("NOTION_DATABASE_ID", "test_database_id")
					m.SetEnv("CERT_PATH", "/path/to/cert")
				},
				wantErr: true,
			},
			{
				name: "Only CERT_PATH missing (optional)",
				setupMock: func(m *MockEnvLoader) {
					m.SetEnv("NOTION_TOKEN", "test_token")
					m.SetEnv("NOTION_DATABASE_ID", "test_database_id")
					m.SetEnv("KOBO_DB_PATH", "/path/to/kobo.db")
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mock := NewMockEnvLoader()
				tt.setupMock(mock)

				_, err := GetConfigWithLoader(mock)

				if (err != nil) != tt.wantErr {
					t.Errorf("GetConfigWithLoader() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestGetConfig(t *testing.T) {
	// Setup test environment variables
	os.Setenv("NOTION_TOKEN", "test_token")
	os.Setenv("NOTION_DATABASE_ID", "test_database_id")
	os.Setenv("KOBO_DB_PATH", "/path/to/kobo.db")
	os.Setenv("CERT_PATH", "/path/to/cert")

	// Test actual GetConfig
	config, err := GetConfig()

	if err != nil {
		t.Errorf("GetConfig() error = %v, want nil", err)
	}

	// Cleanup
	os.Unsetenv("NOTION_TOKEN")
	os.Unsetenv("NOTION_DATABASE_ID")
	os.Unsetenv("KOBO_DB_PATH")
	os.Unsetenv("CERT_PATH")

	// Assert expected values
	if config.NotionToken != "test_token" {
		t.Errorf("config.NotionToken = %v, want %v", config.NotionToken, "test_token")
	}
	if config.DatabaseID != "test_database_id" {
		t.Errorf("config.DatabaseID = %v, want %v", config.DatabaseID, "test_database_id")
	}
	if config.DBPath != "/path/to/kobo.db" {
		t.Errorf("config.DBPath = %v, want %v", config.DBPath, "/path/to/kobo.db")
	}
	if config.CertPath != "/path/to/cert" {
		t.Errorf("config.CertPath = %v, want %v", config.CertPath, "/path/to/cert")
	}
}
