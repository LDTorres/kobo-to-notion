package notion_test

import (
	"context"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/notion"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabaseClient mocks the NotionDatabaseClient interface
type MockDatabaseClient struct {
	mock.Mock
}

func (m *MockDatabaseClient) Query(ctx context.Context, id notionapi.DatabaseID, req *notionapi.DatabaseQueryRequest) (*notionapi.DatabaseQueryResponse, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*notionapi.DatabaseQueryResponse), args.Error(1)
}

// MockPageClient mocks the NotionPageClient interface
type MockPageClient struct {
	mock.Mock
}

func (m *MockPageClient) Create(ctx context.Context, req *notionapi.PageCreateRequest) (*notionapi.Page, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*notionapi.Page), args.Error(1)
}

func (m *MockPageClient) Update(ctx context.Context, pageId notionapi.PageID, req *notionapi.PageUpdateRequest) (*notionapi.Page, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*notionapi.Page), args.Error(1)
}

// Setup function to initialize logger for tests
func setupLogger() {
	// Create a temporary log file
	tempDir := filepath.Join(os.TempDir(), "notion_test")
	os.MkdirAll(tempDir, 0755)
	logPath := filepath.Join(tempDir, "test.log")
	
	logger.Init(logPath)
}

func TestNotionServiceCreation(t *testing.T) {
	// Test creating a new service
	service := notion.NewNotionService("test-token")
	assert.NotNil(t, service, "Service should not be nil")
}

func TestGetBookmarkIDs(t *testing.T) {
	setupLogger()
	defer logger.Close()
	
	// Create mock clients
	mockDBClient := new(MockDatabaseClient)
	
	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithDatabaseClient(mockDBClient)
	service.WithContextFunc(func() context.Context {
		return context.Background()
	})
	
	// Create mock response
	mockResponse := &notionapi.DatabaseQueryResponse{
		Results: []notionapi.Page{
			{
				Properties: notionapi.Properties{
					"Bookmark ID": &notionapi.RichTextProperty{
						RichText: []notionapi.RichText{
							{
								PlainText: "bookmark1",
							},
						},
					},
				},
			},
			{
				Properties: notionapi.Properties{
					"Bookmark ID": &notionapi.RichTextProperty{
						RichText: []notionapi.RichText{
							{
								PlainText: "bookmark2",
							},
						},
					},
				},
			},
		},
		HasMore:    false,
		NextCursor: "",
	}
	
	// Configure mock to return our response
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.Anything).Return(mockResponse, nil)
	
	// Test the function
	bookmarkIDs, err := service.GetBookmarkIDs("test-db-id")
	
	assert.NoError(t, err, "GetBookmarkIDs should not return an error")
	assert.Equal(t, 2, len(bookmarkIDs), "Should return 2 bookmark IDs")
	assert.True(t, bookmarkIDs["bookmark1"], "bookmark1 should be in the map")
	assert.True(t, bookmarkIDs["bookmark2"], "bookmark2 should be in the map")
	
	mockDBClient.AssertExpectations(t)
}

func TestGetBookmarkIDsWithPagination(t *testing.T) {
	setupLogger()
	defer logger.Close()
	
	// Create mock clients
	mockDBClient := new(MockDatabaseClient)
	
	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithDatabaseClient(mockDBClient)
	service.WithContextFunc(func() context.Context {
		return context.Background()
	})
	
	// Create first page response
	firstPageResponse := &notionapi.DatabaseQueryResponse{
		Results: []notionapi.Page{
			{
				Properties: notionapi.Properties{
					"Bookmark ID": &notionapi.RichTextProperty{
						RichText: []notionapi.RichText{
							{
								PlainText: "bookmark1",
							},
						},
					},
				},
			},
		},
		HasMore:    true,
		NextCursor: "cursor1",
	}
	
	// Create second page response
	secondPageResponse := &notionapi.DatabaseQueryResponse{
		Results: []notionapi.Page{
			{
				Properties: notionapi.Properties{
					"Bookmark ID": &notionapi.RichTextProperty{
						RichText: []notionapi.RichText{
							{
								PlainText: "bookmark2",
							},
						},
					},
				},
			},
		},
		HasMore:    false,
		NextCursor: "",
	}
	
	// Configure mocks for pagination
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.MatchedBy(func(req *notionapi.DatabaseQueryRequest) bool {
		return req.StartCursor == ""
	})).Return(firstPageResponse, nil)
	
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.MatchedBy(func(req *notionapi.DatabaseQueryRequest) bool {
		return req.StartCursor == "cursor1"
	})).Return(secondPageResponse, nil)
	
	// Test the function
	bookmarkIDs, err := service.GetBookmarkIDs("test-db-id")
	
	assert.NoError(t, err, "GetBookmarkIDs should not return an error")
	assert.Equal(t, 2, len(bookmarkIDs), "Should return 2 bookmark IDs")
	assert.True(t, bookmarkIDs["bookmark1"], "bookmark1 should be in the map")
	assert.True(t, bookmarkIDs["bookmark2"], "bookmark2 should be in the map")
	
	mockDBClient.AssertExpectations(t)
}

func TestAddBookmark(t *testing.T) {
	setupLogger()
	defer logger.Close()
	
	// Create mock clients
	mockPageClient := new(MockPageClient)
	
	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithPageClient(mockPageClient)
	service.WithContextFunc(func() context.Context {
		return context.Background()
	})
	
	// Create a test bookmark
	bookmark := kobo.Bookmark{
		BookmarkID:  "test-bookmark-id",
		VolumeID:    "test-volume-id",
		Text:        "This is a test highlight",
		Annotation:  "This is a test annotation",
		Type:        "highlight",
		DateCreated: "2023-01-01T12:00:00Z",
	}
	
	// Configure mock
	mockPageClient.On("Create", mock.Anything, mock.Anything).Return(&notionapi.Page{
		ID: "test",
	}, nil)
	
	// Test the function
	_, err := service.AddBookmark("test-db-id", bookmark)

	assert.NoError(t, err, "AddBookmark should not return an error")
	mockPageClient.AssertExpectations(t)
	
	// Verify the page creation request (optional, can be more specific)
	createCall := mockPageClient.Calls[0]
	req := createCall.Arguments.Get(1).(*notionapi.PageCreateRequest)
	
	// Check some aspects of the request
	titleProp, ok := req.Properties["Book Title"].(notionapi.TitleProperty)
	assert.True(t, ok, "Book Title should be a TitleProperty")
	assert.Equal(t, "test-volume-id", titleProp.Title[0].Text.Content)
	
	bookmarkProp, ok := req.Properties["Bookmark ID"].(notionapi.RichTextProperty)
	assert.True(t, ok, "Bookmark ID should be a RichTextProperty")
	assert.Equal(t, "test-bookmark-id", bookmarkProp.RichText[0].Text.Content)
}

// TestIntegration would be used for integration testing with a real Notion API
func TestIntegrationWithNotion(t *testing.T) {
	err := godotenv.Load("../.env")

	if err != nil {
		t.Log(`Error loading .env file`, err)
		return
	}
	
	// Skip this test unless specifically running integration tests
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test")
	}
	
	// Get credentials from environment
	notionToken := os.Getenv("NOTION_TOKEN")
	databaseID := os.Getenv("NOTION_DATABASE_ID")
	
	if notionToken == "" || databaseID == "" {
		t.Fatal("NOTION_TOKEN and NOTION_DATABASE_ID environment variables must be set")
	}
	
	setupLogger()
	defer logger.Close()
	
	// Create and initialize service
	service := notion.NewNotionService(notionToken)
	
	// Test fetching bookmarks
	_, err = service.GetBookmarkIDs(databaseID)
	assert.NoError(t, err, "Failed to get bookmark IDs")
	
	// Create a test bookmark with a unique ID
	uniqueID := "test-" + time.Now().Format("20060102150405")
	bookmark := kobo.Bookmark{
		BookmarkID:  uniqueID,
		VolumeID:    "test-volume-id",
		Text:        "This is an integration test highlight",
		Annotation:  "This is an integration test annotation",
		Type:        "highlight",
		DateCreated: "2025-01-25T22:59:55.080",
	}
	
	// Add the bookmark
	page, err := service.AddBookmark(databaseID, bookmark)

	assert.IsType(t, notionapi.ObjectID(""), page.ID)

	assert.NoError(t, err, "Failed to add bookmark to Notion")
	
	// Verify it was added by fetching bookmarks again
	updatedBookmarkIDs, err := service.GetBookmarkIDs(databaseID)
	assert.NoError(t, err, "Failed to get updated bookmark IDs")
	
	assert.True(t, updatedBookmarkIDs[uniqueID], "Added bookmark not found in Notion")

	_, err = service.ArchiveBookmark(databaseID, page.ID)

	assert.NoError(t, err, "Failed to achieve bookmark")
}