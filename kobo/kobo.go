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

// DatabaseAccessor defines an interface for database operations
type DatabaseAccessor interface {
	GetBookmarks() ([]Bookmark, error)
}

// SQLiteAccessor implements DatabaseAccessor for SQLite database
type SQLiteAccessor struct {
	DBPath string
}

// NewSQLiteAccessor creates a new instance of SQLiteAccessor
func NewSQLiteAccessor(dbPath string) *SQLiteAccessor {
	return &SQLiteAccessor{
		DBPath: dbPath,
	}
}

// GetBookmarks fetches bookmarks from a Kobo SQLite database
func (sa *SQLiteAccessor) GetBookmarks() ([]Bookmark, error) {
	db, err := sql.Open("sqlite3", sa.DBPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return queryBookmarks(db)
}

func queryBookmarks(db *sql.DB) ([]Bookmark, error) {
	query := `
    SELECT
      BookmarkID,
      VolumeID,
      IFNULL(Text, '') AS Text,
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
		var bm Bookmark
		if err := rows.Scan(&bm.BookmarkID, &bm.VolumeID, &bm.Text, &bm.Annotation, &bm.Type, &bm.DateCreated); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, bm)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bookmarks, nil
}

func GetBookmarks(dbPath string) ([]Bookmark, error) {
	accessor := NewSQLiteAccessor(dbPath)
	return accessor.GetBookmarks()
}