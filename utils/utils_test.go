package utils

import (
	"kobo-to-notion/kobo"
	"testing"
	"time"
)

func TestGetBookNameFromVolumeID(t *testing.T) {
	tests := []struct {
		volumeID string
		expected string
	}{
		{"file:///mnt/onboard/MyBook.epub", "MyBook"},
		{"file:///mnt/onboard/My Book.epub", "My Book"},
		{"file:///mnt/onboard/Folder/MyBook.pdf", "MyBook"},
		{"file:///mnt/onboard/Folder/MyBook", "MyBook"},
	}

	for _, test := range tests {
		result := GetBookNameFromVolumeID(test.volumeID)
		if result != test.expected {
			t.Errorf("GetBookNameFromVolumeID(%q) = %q; want %q", test.volumeID, result, test.expected)
		}
	}
}

func TestFilterNewBookmarks(t *testing.T) {
	bookmarks := []kobo.Bookmark{
		{BookmarkID: "1"},
		{BookmarkID: "2"},
		{BookmarkID: "3"},
	}

	existingBookmarks := map[string]bool{
		"2": true,
	}

	expected := []kobo.Bookmark{
		{BookmarkID: "1"},
		{BookmarkID: "3"},
	}

	result := FilterNewBookmarks(bookmarks, existingBookmarks)

	if len(result) != len(expected) {
		t.Errorf("FilterNewBookmarks length = %d; want %d", len(result), len(expected))
	}

	for i, bookmark := range result {
		if bookmark.BookmarkID != expected[i].BookmarkID {
			t.Errorf("FilterNewBookmarks[%d] = %q; want %q", i, bookmark.BookmarkID, expected[i].BookmarkID)
		}
	}
}

func TestParseKoboBookmarkDate(t *testing.T) {
	tests := []struct {
		dateStr  string
		expected string // Expected time in RFC3339 format
		wantErr  bool
	}{
		{"2024-03-10T12:34:56Z", "2024-03-10T12:34:56Z", false},
		{"2024-03-10T12:34:56", "2024-03-10T12:34:56Z", false},
		{"invalid-date", "", true},
	}

	for _, test := range tests {
		result, err := ParseKoboBookmarkDate(test.dateStr)
		if (err != nil) != test.wantErr {
			t.Errorf("ParseKoboBookmarkDate(%q) error = %v; wantErr %v", test.dateStr, err, test.wantErr)
		}
		if !test.wantErr && result.Format(time.RFC3339) != test.expected {
			t.Errorf("ParseKoboBookmarkDate(%q) = %q; want %q", test.dateStr, result.Format(time.RFC3339), test.expected)
		}
	}
}