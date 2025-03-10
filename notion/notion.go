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

// Interfaces for the Notion API clients
type NotionDatabaseClient interface {
	Query(ctx context.Context, id notionapi.DatabaseID, req *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error)
}

type NotionPageClient interface {
	Create(ctx context.Context, req *notionapi.PageCreateRequest) (*notionapi.Page, error)
	Update(ctx context.Context, pageId notionapi.PageID, req *notionapi.PageUpdateRequest) (*notionapi.Page, error)
}

// NotionService encapsulates all Notion operations
type NotionService struct {
	client      *notionapi.Client
	dbClient    NotionDatabaseClient
	pageClient  NotionPageClient
	contextFunc func() context.Context
}

// NewNotionService creates a new NotionService
func NewNotionService(notionToken string) *NotionService {
	client := notionapi.NewClient(notionapi.Token(notionToken))
	return &NotionService{
		client:      client,
		dbClient:    client.Database, // Use client's DB interface
		pageClient:  client.Page,     // Use client's Page interface
		contextFunc: context.Background,
	}
}

// WithHTTPClient allows configuring the HTTP client
func (s *NotionService) WithHTTPClient(httpClient *http.Client) *NotionService {
	s.client = notionapi.NewClient(notionapi.Token(s.client.Token), notionapi.WithHTTPClient(httpClient))
	s.dbClient = s.client.Database
	s.pageClient = s.client.Page
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

		// Add cursor only if it's not empty
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := s.dbClient.Query(s.contextFunc(), notionapi.DatabaseID(databaseID), query)
		if err != nil {
			return nil, err
		}

		// Extract BookmarkIDs
		for _, page := range res.Results {
			if prop, ok := page.Properties["Bookmark ID"].(*notionapi.RichTextProperty); ok {
				for _, text := range prop.RichText {
					existingBookmarks[text.PlainText] = true
				}
			}
		}

		// Check if there are more pages to fetch
		if res.HasMore && res.NextCursor != "" {
			startCursor = res.NextCursor
		} else {
			break
		}
	}

	return existingBookmarks, nil
}

// AddBookmark adds a new bookmark to Notion
func (s *NotionService) AddBookmark(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
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
			"Book Title": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookName,
						},
					},
				},
			},
			"Highlighted Text": notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Text,
						},
					},
				},
			},
			"Annotation": notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Annotation,
						},
					},
				},
			},
			"Type": notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.Type,
						},
					},
				},
			},
			"Date Created": notionapi.DateProperty{
				Date: &notionapi.DateObject{
					Start: &createdAt,
				},
			},
			"Bookmark ID": notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: &notionapi.Text{
							Content: bookmark.BookmarkID,
						},
					},
				},
			},
			"Book Name": notionapi.RichTextProperty{
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
								Content: "Highlighted Text",
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
								Content: "Annotation",
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
func InitializeNotionClient(certPath string, notionToken string, databaseID string) error {
	defaultService = NewNotionService(notionToken)
	return defaultService.InitializeWithCert(certPath)
}

// GetNotionBookmarkIDs fetches all BookmarkIDs from Notion using the global client
func GetNotionBookmarkIDs(databaseID string) (map[string]bool, error) {
	if defaultService == nil {
		return nil, errors.New("notion client not initialized")
	}
	return defaultService.GetBookmarkIDs(databaseID)
}

// AddBookmarkToNotion adds a new bookmark to Notion using the global client
func AddBookmarkToNotion(databaseID string, bookmark kobo.Bookmark) (*notionapi.Page, error) {
	if defaultService == nil {
		return nil, errors.New("notion client not initialized")
	}
	return defaultService.AddBookmark(databaseID, bookmark)
}