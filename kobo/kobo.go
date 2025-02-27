package kobo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Bookmark struct {
	BookmarkID  string
	VolumeID    string
	Text        string
	Annotation  string
	Type        string
	DateCreated string
}

// Fetch bookmarks from Kobo SQLite database
func GetBookmarks(dbPath string) ([]Bookmark, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
    SELECT
      BookmarkID,
      VolumeID,
      Text,
      IFNULL(Annotation, 'None') AS Annotation,
      Type,
      DateCreated
    FROM Bookmark
    WHERE Annotation IS NOT NULL OR Text IS NOT NULL
    ORDER BY DateCreated DESC;
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var h Bookmark
		if err := rows.Scan(&h.BookmarkID, &h.VolumeID, &h.Text, &h.Annotation, &h.Type, &h.DateCreated); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, h)
	}

	return bookmarks, nil
}