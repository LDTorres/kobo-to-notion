package kobo

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDatabase(t *testing.T) (string, func()) {
	// Create a temporary directory for our test database
	tempDir, err := os.MkdirTemp("", "kobo_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	dbPath := filepath.Join(tempDir, "KoboReader.sqlite")

	// Create and setup the test database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the Bookmark table
	_, err = db.Exec(`
		CREATE TABLE Bookmark (
			BookmarkID TEXT PRIMARY KEY,
			VolumeID TEXT,
			Text TEXT,
			Annotation TEXT,
			Type TEXT,
			DateCreated TEXT,
			Color TEXT
		);
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create Bookmark table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO Bookmark (BookmarkID, VolumeID, Text, Annotation, Type, DateCreated, Color) VALUES
		('bm1', 'vol1', 'Sample text 1', 'Sample annotation 1', 'highlight', '2023-01-04T12:00:00Z', '0'),
		('bm2', 'vol1', 'Sample text 2', NULL, 'highlight', '2023-01-03T12:00:00Z', '1'),
		('bm3', 'vol2', NULL, 'Sample annotation 3', 'note', '2023-01-02T12:00:00Z', '2'),
		('bm4', 'vol2', 'Sample text 4', 'Sample annotation 4', 'highlight', '2023-01-01T12:00:00Z', '3'),
		('bm5', 'vol3', NULL, NULL, 'bookmark', '2023-01-05T12:00:00Z', '4');
	`)
	if err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Return the path to the test database and a cleanup function
	return dbPath, func() {
		os.RemoveAll(tempDir)
	}
}

func TestGetBookmarks(t *testing.T) {
	// Create a test database
	dbPath, cleanup := createTestDatabase(t)
	defer cleanup()

	// Test GetBookmarks function
	bookmarks, err := GetBookmarks(dbPath)
	if err != nil {
		t.Fatalf("GetBookmarks failed: %v", err)
	}

	// Check that we got the expected number of bookmarks
	// We expect 4 bookmarks since the query filters out rows where both Text and Annotation are NULL
	expectedCount := 4
	if len(bookmarks) != expectedCount {
		t.Errorf("Expected %d bookmarks, got %d", expectedCount, len(bookmarks))
	}

	// Verify that the results are ordered by DateCreated DESC
	if len(bookmarks) >= 2 {
		if bookmarks[0].DateCreated < bookmarks[1].DateCreated {
			t.Errorf("Bookmarks not ordered by DateCreated DESC")
		}
	}

	// Check that NULL Text values are replaced with an empty string
	for _, bm := range bookmarks {
		if bm.BookmarkID == "bm3" {
			if bm.Text != "" {
				t.Errorf("Expected NULL Text to be replaced with empty string, got '%s'", bm.Text)
			}
		}
	}

	// Check that NULL Annotation values are replaced with ''
	for _, bm := range bookmarks {
		if bm.BookmarkID == "bm2" {
			if bm.Annotation != "" {
				t.Errorf("Expected NULL Annotation to be replaced with '', got '%s'", bm.Annotation)
			}
		}
	}
}

func TestGetBookmarksWithEmptyDB(t *testing.T) {
	// Create a temporary directory for our empty test database
	tempDir, err := os.MkdirTemp("", "kobo_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "EmptyKoboReader.sqlite")

	// Create an empty database with the table but no data
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open empty test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE Bookmark (
			BookmarkID TEXT PRIMARY KEY,
			VolumeID TEXT,
			Text TEXT,
			Annotation TEXT,
			Type TEXT,
			DateCreated TEXT,
			Color TEXT
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create empty Bookmark table: %v", err)
	}
	db.Close()

	// Test GetBookmarks with empty database
	bookmarks, err := GetBookmarks(dbPath)
	if err != nil {
		t.Fatalf("GetBookmarks failed with empty database: %v", err)
	}

	// Check that we got 0 bookmarks
	if len(bookmarks) != 0 {
		t.Errorf("Expected 0 bookmarks from empty database, got %d", len(bookmarks))
	}
}

func TestGetBookmarksWithInvalidPath(t *testing.T) {
	// Test with an invalid database path
	_, err := GetBookmarks("/path/to/nonexistent/database.sqlite")
	if err == nil {
		t.Error("Expected error with invalid database path, got nil")
	}
}
