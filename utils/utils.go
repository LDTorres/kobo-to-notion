package utils

import (
	"errors"
	"kobo-to-notion/kobo"
	"path/filepath"
	"strings"
	"time"

	"github.com/jomei/notionapi"
)

// Extracts book name from VolumeID path
func GetBookNameFromVolumeID(volumeID string) string {
	path := strings.TrimPrefix(volumeID, "file://")
	filenameWithExt := filepath.Base(path)
	bookName := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	return bookName
}

// Filters bookmarks to only keep new ones
func FilterNewBookmarks(bookmarks []kobo.Bookmark, existingBookmarks map[string]bool) []kobo.Bookmark {
	var newBookmarks []kobo.Bookmark
	for _, bookmark := range bookmarks {
		if !existingBookmarks[bookmark.BookmarkID] {
			newBookmarks = append(newBookmarks, bookmark)
		}
	}
	return newBookmarks
}

func ParseKoboBookmarkDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, errors.New("empty date string")
	}

	if !strings.HasSuffix(dateStr, "Z") {
		dateStr += "Z"
	}

	parsedDate, err := time.Parse(time.RFC3339, dateStr)

	if err != nil {
		return time.Time{}, err
	}

	return parsedDate, nil
}

func SplitText(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		return nil
	}

	runes := []rune(text)
	var chunks []string

	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func ContainsBlockRichText(text string, blocks []notionapi.Block) bool {
	for _, block := range blocks {
		if strings.Contains(block.GetRichTextString(), text) {
			return true
		}
	}
	return false
}

func ContainsBookmark(text string, bookmarks []kobo.Bookmark) bool {
	for _, bookmark := range bookmarks {
		if bookmark.Text != "" && strings.Contains(text, bookmark.Text) {
			return true
		}
		if bookmark.Annotation != "" && strings.Contains(text, bookmark.Annotation) {
			return true
		}
	}
	return false
}
