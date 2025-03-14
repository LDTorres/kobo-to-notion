package utils

import (
	"errors"
	"kobo-to-notion/kobo"
	"path/filepath"
	"strings"
	"time"
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