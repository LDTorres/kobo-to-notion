package notion

import (
	"context"
	"errors"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/utils"
	"net/http"

	"github.com/jomei/notionapi"
)

// Constants for property names
const (
	PropBookTitle      = "Book Title"
	PropHighlightedText = "Highlighted Text"
	PropAnnotation     = "Annotation"
	PropType           = "Type"
	PropDateCreated    = "Date Created"
	PropBookmarkID     = "Bookmark ID"
	PropBookName       = "Book Name"
	
	ErrNotionClientNotInitialized = "notion client not initialized"
)

// Interfaces for the Notion API clients
type NotionDatabaseClient interface {
	Query(ctx context.Context, id notionapi.DatabaseID, req *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error)
}

type NotionPageClient interface {
	Create(ctx context.Context, req *notionapi.PageCreateRequest) (*notionapi.Page, error)
	Update(ctx context.Context, pageId notionapi.PageID, req *notionapi.PageUpdateRequest) (*notionapi.Page, error)
	Get(ctx context.Context, pageId notionapi.PageID) (*notionapi.Page, error)
}

// Block interface to abstract the notion API
type NotionBlockClient interface {
	AppendChildren(ctx context.Context, blockID notionapi.BlockID, request *notionapi.AppendBlockChildrenRequest) (*notionapi.AppendBlockChildrenResponse, error)
	GetChildren(ctx context.Context, blockID notionapi.BlockID, pagination *notionapi.Pagination) (*notionapi.GetChildrenResponse, error)
	Delete(ctx context.Context, blockID notionapi.BlockID) (notionapi.Block, error)
}

// NotionService encapsulates all Notion operations
type NotionService struct {
	client           *notionapi.Client
	dbClient         NotionDatabaseClient
	pageClient       NotionPageClient
	blockClient      notionapi.BlockService // Using the actual BlockService from the API
	contextFunc      func() context.Context
	createIndividual bool // Flag to control bookmark creation mode
}

// NewNotionService creates a new NotionService
func NewNotionService(notionToken string, createIndividual bool) *NotionService {
	client := notionapi.NewClient(notionapi.Token(notionToken))
	return &NotionService{
		client:           client,
		dbClient:         client.Database, // Use client's DB interface
		pageClient:       client.Page,     // Use client's Page interface
		blockClient:      client.Block,    // Use client's Block interface
		contextFunc:      context.Background,
		createIndividual: createIndividual,
	}
}

// WithHTTPClient allows configuring the HTTP client
func (s *NotionService) WithHTTPClient(httpClient *http.Client) *NotionService {
	s.client = notionapi.NewClient(notionapi.Token(s.client.Token), notionapi.WithHTTPClient(httpClient))
	s.dbClient = s.client.Database
	s.pageClient = s.client.Page
	s.blockClient = s.client.Block
	return s
}

// WithDatabaseClient allows setting a custom database client (mainly for testing)
func (s *NotionService) WithDatabaseClient(dbClient NotionDatabaseClient) *NotionService {
	s.dbClient = dbClient
	return s
}

// WithPageClient allows setting a custom page client (mainly for testing)
func (s *NotionService) WithPageClient(pageClient NotionPageClient) *NotionService {
	s.pageClient = pageClient
	return s
}

// WithBlockClient allows setting a custom block client (mainly for testing)
func (s *NotionService) WithBlockClient(blockClient notionapi.BlockService) *NotionService {
	s.blockClient = blockClient
	return s
}

// WithContextFunc allows setting a custom context function (mainly for testing)
func (s *NotionService) WithContextFunc(ctxFunc func() context.Context) *NotionService {
	s.contextFunc = ctxFunc
	return s
}

// InitializeWithCert initializes the service with a certificate file
func (s *NotionService) InitializeWithCert(certPath string) error {
	// Configure a secure HTTP client with embedded certificates
	httpClient, err := utils.ConfigureSecureHTTPClientWithFile(certPath)
	if err != nil {
		return err
	}

	if httpClient != nil {
		s.WithHTTPClient(httpClient)
	}

	return nil
}

// GetBookmarkIDs fetches all BookmarkIDs from Notion
func (s *NotionService) GetBookmarkIDs(databaseID string) (map[string]bool, error) {
	existingBookmarks := make(map[string]bool)
	var startCursor notionapi.Cursor

	for {
		query := &notionapi.DatabaseQueryRequest{}
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := s.dbClient.Query(s.contextFunc(), notionapi.DatabaseID(databaseID), query)
		if err != nil {
			return nil, err
		}

		s.extractBookmarkIDs(res, existingBookmarks)

		if !res.HasMore || res.NextCursor == "" {
			break
		}
		startCursor = res.NextCursor
	}

	return existingBookmarks, nil
}

// extractBookmarkIDs extracts bookmark IDs from query results and adds them to the map
func (s *NotionService) extractBookmarkIDs(res *notionapi.DatabaseQueryResponse, bookmarks map[string]bool) {
	for _, page := range res.Results {
		if prop, ok := page.Properties[PropBookmarkID].(*notionapi.RichTextProperty); ok {
			for _, text := range prop.RichText {
				bookmarks[text.PlainText] = true
			}
		}
	}
}

// GetPagesByBookName fetches pages grouped by book name
func (s *NotionService) GetPagesByBookName(databaseID string) (map[string]notionapi.PageID, error) {
	bookPages := make(map[string]notionapi.PageID)
	var startCursor notionapi.Cursor

	for {
		query := &notionapi.DatabaseQueryRequest{}
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := s.dbClient.Query(s.contextFunc(), notionapi.DatabaseID(databaseID), query)
		if err != nil {
			return nil, err
		}

		for _, page := range res.Results {
			// Extract book name from Book Title property
			if titleProp, ok := page.Properties[PropBookTitle].(*notionapi.TitleProperty); ok && len(titleProp.Title) > 0 {
				bookName := titleProp.Title[0].PlainText
				bookPages[bookName] = notionapi.PageID(page.ID)
			}
		}

		if !res.HasMore || res.NextCursor == "" {
			break
		}
		startCursor = res.NextCursor
	}

	return bookPages, nil
}

// AddBookmark adds a new bookmark to Notion
func (s *NotionService) AddBookmark(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	if s.createIndividual {
		return s.addIndividualBookmark(databaseID, bookmark)
	}
	
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
	if s.createIndividual {
		// Handle individual bookmarks (legacy mode)
		for _, bookmark := range bookmarks {
			_, err := s.addIndividualBookmark(databaseID, bookmark)
			if err != nil {
				return err
			}
		}
		return nil
	}
	
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
			// Page exists, clear existing blocks and add all bookmarks at once
			logger.Logger.Printf("Updating existing page for book: %s with %d bookmarks\n", bookName, len(bookBookmarks))
			
			// Clear existing blocks
			err := s.clearPageBlocks(pageID)
			if err != nil {
				logger.Logger.Printf("Error clearing blocks for book %s: %v", bookName, err)
				continue
			}
			
			// Append all bookmarks in a single call
			_, err = s.appendBookmarksToPage(pageID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error updating bookmarks for book %s: %v", bookName, err)
				continue
			}
		} else {
			// Create a new page with all bookmarks
			logger.Logger.Printf("Creating new page for book: %s with %d bookmarks\n", bookName, len(bookBookmarks))
			_, err = s.createBookPageWithBookmarks(databaseID, bookBookmarks)
			if err != nil {
				logger.Logger.Printf("Error creating page for book %s: %v", bookName, err)
				continue
			}
		}
		
		logger.Logger.Printf("Successfully processed book: %s\n", bookName)
	}
	
	return nil
}

// clearPageBlocks removes all existing blocks from a page
func (s *NotionService) clearPageBlocks(pageID notionapi.PageID) error {
	// Get all blocks from the page
	blocks, err := s.getAllBlocksFromPage(pageID)
	if err != nil {
		return err
	}
	
	// Delete each block
	for _, block := range blocks {
		// We need to use type assertion to access the ID
		if basicBlock, ok := block.(*notionapi.BasicBlock); ok {
			_, err := s.blockClient.Delete(s.contextFunc(), basicBlock.ID)
			if err != nil {
				return err
			}
		}
	}
	
	logger.Logger.Printf("Cleared %d blocks from page %s\n", len(blocks), pageID)
	return nil
}

// getAllBlocksFromPage retrieves all blocks from a page
func (s *NotionService) getAllBlocksFromPage(pageID notionapi.PageID) ([]notionapi.Block, error) {
	var blocks []notionapi.Block
	var startCursor notionapi.Cursor
	
	for {
		pagination := &notionapi.Pagination{}
		if startCursor != "" {
			pagination.StartCursor = startCursor
		}
		
		resp, err := s.blockClient.GetChildren(s.contextFunc(), notionapi.BlockID(pageID), pagination)
		if err != nil {
			return nil, err
		}
		
		blocks = append(blocks, resp.Results...)
		
		if !resp.HasMore {
			break
		}
		
		startCursor = notionapi.Cursor(resp.NextCursor)
	}
	
	return blocks, nil
}

// appendBookmarksToPage appends multiple bookmarks to a page in a single call
func (s *NotionService) appendBookmarksToPage(pageID notionapi.PageID, bookmarks []kobo.Bookmark) (*notionapi.Page, error) {
	// Create blocks for all bookmarks
	var allBlocks []notionapi.Block
	
	for _, bookmark := range bookmarks {
		blocks := s.createBookmarkBlocks(bookmark)
		allBlocks = append(allBlocks, blocks...)
	}
	
	// Use the Notion API to append all blocks at once
	_, err := s.blockClient.AppendChildren(s.contextFunc(), notionapi.BlockID(pageID), &notionapi.AppendBlockChildrenRequest{
		Children: allBlocks,
	})
	
	if err != nil {
		return nil, err
	}
	
	logger.Logger.Printf("Added %d blocks to page %s in a single call\n", len(allBlocks), pageID)
	
	// Return the updated page
	page, err := s.pageClient.Get(s.contextFunc(), pageID)
	if err != nil {
		// Even if we can't get the page, the blocks were successfully appended
		logger.Logger.Printf("Warning: could not get updated page: %v\n", err)
		return nil, nil
	}
	
	return page, nil
}

// createBookPageWithBookmarks creates a new page with multiple bookmarks
func (s *NotionService) createBookPageWithBookmarks(databaseID string, bookmarks []kobo.Bookmark) (*notionapi.Page, error) {
	if len(bookmarks) == 0 {
		return nil, errors.New("no bookmarks provided")
	}
	
	// Use the first bookmark for page properties
	firstBookmark := bookmarks[0]
	bookName := utils.GetBookNameFromVolumeID(firstBookmark.VolumeID)
	
	parsedDate, err := utils.ParseKoboBookmarkDate(firstBookmark.DateCreated)
	if err != nil {
		return nil, err
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
	
	page, err := s.pageClient.Create(s.contextFunc(), payload)
	if err != nil {
		return nil, err
	}
	
	logger.Logger.Printf("Book page created successfully with %d bookmarks!\n", len(bookmarks))
	return page, nil
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

// createBookmarkBlocks creates a set of blocks for a bookmark
func (s *NotionService) createBookmarkBlocks(bookmark kobo.Bookmark) []notionapi.Block {
	emoji := notionapi.Emoji("📚")

	return []notionapi.Block{
		&notionapi.Heading3Block{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeHeading3,
			},
			Heading3: notionapi.Heading{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: PropHighlightedText,
						},
						Annotations: &notionapi.Annotations{
							Bold:  true,
							Color: notionapi.ColorBlue,
						},
					},
				},
			},
		},
		&notionapi.ParagraphBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeParagraph,
			},
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: bookmark.Text,
						},
					},
				},
			},
		},
		&notionapi.DividerBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeDivider,
			},
		},
		&notionapi.Heading3Block{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeHeading3,
			},
			Heading3: notionapi.Heading{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: PropAnnotation,
						},
						Annotations: &notionapi.Annotations{
							Bold:  true,
							Color: notionapi.ColorPurple,
						},
					},
				},
			},
		},
		&notionapi.ParagraphBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeParagraph,
			},
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: bookmark.Annotation,
						},
						Annotations: &notionapi.Annotations{
							Italic: bookmark.Type == "highlight",
						},
					},
				},
			},
		},
		&notionapi.DividerBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeDivider,
			},
		},
		&notionapi.CalloutBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeCallout,
			},
			Callout: notionapi.Callout{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: "Type: " + bookmark.Type,
						},
					},
				},
				Icon: &notionapi.Icon{
					Type:  "emoji",
					Emoji: &emoji,
				},
				Color: string(notionapi.ColorGray),
			},
		},
	}
}

// addIndividualBookmark is the original method for adding bookmarks individually
func (s *NotionService) addIndividualBookmark(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	bookName := utils.GetBookNameFromVolumeID(bookmark.VolumeID)
	emoji := notionapi.Emoji("📚")

	parsedDate, err := utils.ParseKoboBookmarkDate(bookmark.DateCreated)
	if err != nil {
		return nil, err
	}

	createdAt := notionapi.Date(parsedDate)

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
			PropHighlightedText: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Text,
						},
					},
				},
			},
			PropAnnotation: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Annotation,
						},
					},
				},
			},
			PropType: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Type,
						},
					},
				},
			},
			PropDateCreated: notionapi.DateProperty{
				Date: &notionapi.DateObject{
					Start: &createdAt,
				},
			},
			PropBookmarkID: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.BookmarkID,
						},
					},
				},
			},
			PropBookName: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{
							Content: utils.GetBookNameFromVolumeID(bookmark.VolumeID),
						},
					},
				},
			},
		},
		Children: []notionapi.Block{
			&notionapi.Heading3Block{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeHeading3,
				},
				Heading3: notionapi.Heading{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: PropHighlightedText,
							},
							Annotations: &notionapi.Annotations{
								Bold:  true,
								Color: notionapi.ColorBlue,
							},
						},
					},
				},
			},
			&notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: bookmark.Text,
							},
						},
					},
				},
			},
			&notionapi.DividerBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeDivider,
				},
			},
			&notionapi.Heading3Block{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeHeading3,
				},
				Heading3: notionapi.Heading{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: PropAnnotation,
							},
							Annotations: &notionapi.Annotations{
								Bold:  true,
								Color: notionapi.ColorPurple,
							},
						},
					},
				},
			},
			&notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: bookmark.Annotation,
							},
							Annotations: &notionapi.Annotations{
								Italic: bookmark.Type == "highlight",
							},
						},
					},
				},
			},
			&notionapi.DividerBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeDivider,
				},
			},
			&notionapi.CalloutBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeCallout,
				},
				Callout: notionapi.Callout{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: "Type: " + bookmark.Type,
							},
						},
					},
					Icon: &notionapi.Icon{
						Type:  "emoji",
						Emoji: &emoji,
					},
					Color: string(notionapi.ColorGray),
				},
			},
		},
	}

	page, err := s.pageClient.Create(s.contextFunc(), payload)
	if err != nil {
		return nil, err
	}

	logger.Logger.Println("Page created successfully!")
	return page, nil
}

func (s *NotionService) ArchiveBookmark(databaseID string, bookmarkId notionapi.ObjectID) (*notionapi.Page, error) {
	pageID := notionapi.PageID(bookmarkId)
	
	page, err := s.pageClient.Update(s.contextFunc(), pageID, &notionapi.PageUpdateRequest{
		Archived: true, 
		Properties: nil, 
		Icon: nil, 
		Cover: nil,
	})

	if err != nil {
		return nil, err
	}

	return page, nil
}

var defaultService *NotionService

// InitializeNotionClient initializes the global NotionClient
func InitializeNotionClient(certPath string, notionToken string, databaseID string, createIndividual bool) error {
	defaultService = NewNotionService(notionToken, createIndividual)
	return defaultService.InitializeWithCert(certPath)
}

// GetNotionBookmarkIDs fetches all BookmarkIDs from Notion using the global client
func GetNotionBookmarkIDs(databaseID string) (map[string]bool, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.GetBookmarkIDs(databaseID)
}

// GetBookPagesByName fetches all book pages from Notion using the global client
func GetBookPagesByName(databaseID string) (map[string]notionapi.PageID, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.GetPagesByBookName(databaseID)
}

// AddBookmarkToNotion adds a new bookmark to Notion using the global client
func AddBookmarkToNotion(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	if defaultService == nil {
		return nil, errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.AddBookmark(databaseID, bookmark)
}

// AddBookmarksToNotion adds multiple bookmarks to Notion in a batch using the global client
func AddBookmarksToNotion(databaseID string, bookmarks []kobo.Bookmark) error {
	if defaultService == nil {
		return errors.New(ErrNotionClientNotInitialized)
	}
	return defaultService.AddBookmarks(databaseID, bookmarks)
}