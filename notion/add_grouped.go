package notion

import (
	"errors"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/utils"

	"github.com/jomei/notionapi"
)

// AddBookmark adds a new bookmark to Notion
func (s *NotionService) AddBookmark(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	// Get book name
	bookName := utils.GetBookNameFromVolumeID(bookmark.VolumeID)

	// Get existing pages by book name
	bookPages, err := s.GetPagesByBookName(databaseID)
	if err != nil {
		return nil, err
	}

	// Check if a page already exists for this book
	if pageID, exists := bookPages[bookName]; exists {
		// Page exists, add bookmark as a new block
		return s.appendBookmarkToPage(pageID, bookmark)
	}

	// No page exists, create a new one with this bookmark
	return s.createBookPage(databaseID, bookmark)
}

// AddBookmarks adds multiple bookmarks to Notion in a batch
func (s *NotionService) AddBookmarks(databaseID string, bookmarks []kobo.Bookmark) error {
	// Group bookmarks by book name
	bookmarksByBook := make(map[string][]kobo.Bookmark)
	for _, bookmark := range bookmarks {
		bookName := utils.GetBookNameFromVolumeID(bookmark.VolumeID)
		bookmarksByBook[bookName] = append(bookmarksByBook[bookName], bookmark)
	}

	// Get existing pages by book name
	bookPages, err := s.GetPagesByBookName(databaseID)
	if err != nil {
		return err
	}

	// Process each book
	for bookName, bookBookmarks := range bookmarksByBook {
		if pageID, exists := bookPages[bookName]; exists {
			// Page exists, update it completely rather than clearing and re-adding blocks
			logger.Logger.Printf("Updating existing page for book: %s with %d bookmarks\n", bookName, len(bookBookmarks))

			// Update the page with all bookmarks
			err = s.updateBookPage(pageID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error updating page for book %s: %v", bookName, err)
				continue
			}
		} else {
			// Create a new page with all bookmarks
			logger.Logger.Printf("Creating new page for book: %s with %d bookmarks\n", bookName, len(bookBookmarks))
			err = s.createBookPageWithBookmarks(databaseID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error creating page for book %s: %v", bookName, err)
				continue
			}
		}

		logger.Logger.Printf("Successfully processed book: %s\n", bookName)
	}

	return nil
}

// updateBookPage updates an existing page with new bookmarks, replacing all content
func (s *NotionService) updateBookPage(pageID notionapi.PageID, bookmarks []kobo.Bookmark) error {
	if len(bookmarks) == 0 {
		return errors.New("no bookmarks provided")
	}

	// First get the page to ensure it exists
	_, err := s.pageClient.Get(s.contextFunc(), pageID)
	if err != nil {
		return err
	}

	// Current blocks on the page
	currentBlocks, err := s.getAllBlocksFromPage(pageID)
	if err != nil {
		logger.Logger.Printf("Warning: failed to get existing blocks: %v\n", err)
	}

	// Create blocks for all bookmarks, excluding already existing blocks
	var allBlocks []notionapi.Block
	for _, bookmark := range bookmarks {
		// If bookmark text is already in the current blocks, skip
		if utils.ContainsBlockRichText(bookmark.Text, currentBlocks) || utils.ContainsBlockRichText(bookmark.Annotation, currentBlocks) {
			continue
		}

		blocks := s.createBookmarkBlocks(bookmark)
		allBlocks = append(allBlocks, blocks...)
	}

	// Delete existing blocks that are not in the new blocks
	var deletedBlocks []notionapi.BlockID
	/* for _, block := range currentBlocks {
		if block.GetType() != notionapi.BlockTypeQuote {
			continue
		}

		if !utils.ContainsBookmark(block.GetRichTextString(), bookmarks) {
			blockID := block.GetID()
			logger.Logger.Printf("Block does not exist in new blocks: %s\n", blockID)

			_, err := s.blockClient.Delete(s.contextFunc(), blockID)

			if err != nil {
				logger.Logger.Printf("Warning: could not delete block %s: %v\n", blockID, err)
			}

			logger.Logger.Printf("Deleted block %s\n", blockID)
			deletedBlocks = append(deletedBlocks, blockID)
		}
	} */

	// Create new blocks
	if len(allBlocks) > 0 {
		_, err = s.blockClient.AppendChildren(s.contextFunc(), notionapi.BlockID(pageID), &notionapi.AppendBlockChildrenRequest{
			Children: allBlocks,
		})

		if err != nil {
			return err
		}

		logger.Logger.Printf("Page updated with %d new blocks\n", len(allBlocks))
	}

	logger.Logger.Printf("Page deleted blocks: %d\n", len(deletedBlocks))

	return nil
}

// createBookPage creates a new page for a book with the first bookmark
func (s *NotionService) createBookPage(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	bookName := utils.GetBookNameFromVolumeID(bookmark.VolumeID)

	parsedDate, err := utils.ParseKoboBookmarkDate(bookmark.DateCreated)
	if err != nil {
		return nil, err
	}

	createdAt := notionapi.Date(parsedDate)

	// Create blocks for the bookmark content
	bookmarkBlocks := s.createBookmarkBlocks(bookmark)

	payload := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: notionapi.Properties{
			PropBookTitle: notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookName,
						},
					},
				},
			},
			PropDateCreated: notionapi.DateProperty{
				Date: &notionapi.DateObject{
					Start: &createdAt,
				},
			},
			PropBookName: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: bookName,
						},
					},
				},
			},
		},
		Children: bookmarkBlocks,
	}

	page, err := s.pageClient.Create(s.contextFunc(), payload)
	if err != nil {
		return nil, err
	}

	logger.Logger.Println("Book page created successfully!")
	return page, nil
}

// createBookPageWithBookmarks creates a new page with multiple bookmarks
func (s *NotionService) createBookPageWithBookmarks(databaseID string, bookmarks []kobo.Bookmark) error {
	if len(bookmarks) == 0 {
		return errors.New("no bookmarks provided")
	}

	// Use the first bookmark for page properties
	firstBookmark := bookmarks[0]
	bookName := utils.GetBookNameFromVolumeID(firstBookmark.VolumeID)

	parsedDate, err := utils.ParseKoboBookmarkDate(firstBookmark.DateCreated)
	if err != nil {
		return err
	}

	createdAt := notionapi.Date(parsedDate)

	// Create blocks for all bookmarks
	var allBlocks []notionapi.Block

	for _, bookmark := range bookmarks {
		blocks := s.createBookmarkBlocks(bookmark)
		allBlocks = append(allBlocks, blocks...)
	}

	payload := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: notionapi.Properties{
			PropBookTitle: notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookName,
						},
					},
				},
			},
			PropDateCreated: notionapi.DateProperty{
				Date: &notionapi.DateObject{
					Start: &createdAt,
				},
			},
			PropBookName: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: bookName,
						},
					},
				},
			},
		},
		Children: allBlocks,
	}

	_, err = s.pageClient.Create(s.contextFunc(), payload)
	if err != nil {
		return err
	}

	logger.Logger.Printf("Book page created successfully with %d bookmarks!\n", len(bookmarks))
	return nil
}

// appendBookmarkToPage appends a bookmark to an existing page
func (s *NotionService) appendBookmarkToPage(pageID notionapi.PageID, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	// Create blocks for the bookmark
	bookmarkBlocks := s.createBookmarkBlocks(bookmark)

	// Use the Notion API to append blocks to the page
	_, err := s.blockClient.AppendChildren(s.contextFunc(), notionapi.BlockID(pageID), &notionapi.AppendBlockChildrenRequest{
		Children: bookmarkBlocks,
	})

	if err != nil {
		return nil, err
	}

	logger.Logger.Printf("Bookmark successfully appended to existing page: %s\n", pageID)

	// Return the updated page
	page, err := s.pageClient.Get(s.contextFunc(), pageID)
	if err != nil {
		// Even if we can't get the page, the blocks were successfully appended
		logger.Logger.Printf("Warning: could not get updated page: %v\n", err)
		return nil, nil
	}

	return page, nil
}
