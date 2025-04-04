package notion

import (
	"errors"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/utils"

	"github.com/jomei/notionapi"
)

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
		logger.Logger.Printf("Processing book: %s\n", bookName)

		if pageID, exists := bookPages[bookName]; exists {
			err = s.updateBookPage(pageID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error updating page for book %s: %v", bookName, err)
				continue
			}
		} else {
			err = s.createBookPageWithBookmarks(databaseID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error creating page for book %s: %v", bookName, err)
				continue
			}
		}

		logger.Logger.Printf("Successfully processed book: %s\n", bookName)
	}

	// Remove deleted books from notion
	for bookName, pageID := range bookPages {
		if bookmarks, exists := bookmarksByBook[bookName]; !exists {
			if len(bookmarks) != 0 {
				continue
			}

			logger.Logger.Printf("Removing book page for book: %s\n", bookName)
			err := s.ArchivePage(databaseID, pageID)
			if err != nil {
				logger.Logger.Printf("Error removing page for book %s: %v", bookName, err)
				continue
			}
		}
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
		// Skip if the bookmark is empty
		if bookmark.Text != "" && !utils.ContainsBlockRichText(bookmark.Text, currentBlocks) {
			blocks := s.createBookmarkTextBlocks(bookmark)
			allBlocks = append(allBlocks, blocks...)
		}

		// If bookmark text is already in the current blocks, skip
		if bookmark.Annotation != "" && !utils.ContainsBlockRichText(bookmark.Annotation, currentBlocks) {
			blocks := s.createBookmarkAnnotationBlocks(bookmark)
			allBlocks = append(allBlocks, blocks...)
		}
	}

	// Delete existing blocks that are not in the new blocks
	var deletedBlocks []notionapi.BlockID
	for _, block := range currentBlocks {
		if utils.ContainsBookmark(block.GetRichTextString(), bookmarks) {
			continue
		}

		blockID := block.GetID()
		logger.Logger.Printf("Block does not exist in new blocks: %s\n", block.GetRichTextString())

		_, err := s.blockClient.Delete(s.contextFunc(), blockID)

		if err != nil {
			logger.Logger.Printf("Warning: could not delete block %s: %v\n", blockID, err)
		}

		logger.Logger.Printf("Deleted block %s\n", blockID)
		deletedBlocks = append(deletedBlocks, blockID)
	}

	if len(deletedBlocks) > 0 {
		logger.Logger.Printf("Page deleted blocks: %d\n", len(deletedBlocks))
	}

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

	return nil
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
		allBlocks = append(allBlocks, s.createBookmarkTextBlocks(bookmark)...)
		allBlocks = append(allBlocks, s.createBookmarkAnnotationBlocks(bookmark)...)
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
