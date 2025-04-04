package notion_test

import (
	"context"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/notion"
	"os"
	"path/filepath"
	"testing"

	"github.com/jomei/notionapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Import constants from the notion package
var (
	PropBookTitle       = notion.PropBookTitle
	PropHighlightedText = notion.PropHighlightedText
	PropAnnotation      = notion.PropAnnotation
	PropType            = notion.PropType
	PropDateCreated     = notion.PropDateCreated
	PropBookmarkID      = notion.PropBookmarkID
	PropBookName        = notion.PropBookName
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
	args := m.Called(ctx, pageId, req)
	return args.Get(0).(*notionapi.Page), args.Error(1)
}

func (m *MockPageClient) Get(ctx context.Context, pageId notionapi.PageID) (*notionapi.Page, error) {
	args := m.Called(ctx, pageId)
	return args.Get(0).(*notionapi.Page), args.Error(1)
}

// MockBlockClient mocks the NotionBlockClient interface
type MockBlockClient struct {
	mock.Mock
}

func (m *MockBlockClient) AppendChildren(ctx context.Context, blockID notionapi.BlockID, request *notionapi.AppendBlockChildrenRequest) (*notionapi.AppendBlockChildrenResponse, error) {
	args := m.Called(ctx, blockID, request)
	return args.Get(0).(*notionapi.AppendBlockChildrenResponse), args.Error(1)
}

func (m *MockBlockClient) GetChildren(ctx context.Context, blockID notionapi.BlockID, pagination *notionapi.Pagination) (*notionapi.GetChildrenResponse, error) {
	args := m.Called(ctx, blockID, pagination)
	return args.Get(0).(*notionapi.GetChildrenResponse), args.Error(1)
}

func (m *MockBlockClient) Delete(ctx context.Context, blockID notionapi.BlockID) (notionapi.Block, error) {
	args := m.Called(ctx, blockID)
	return args.Get(0).(notionapi.Block), args.Error(1)
}

func (m *MockBlockClient) Get(ctx context.Context, blockID notionapi.BlockID) (notionapi.Block, error) {
	args := m.Called(ctx, blockID)
	return args.Get(0).(notionapi.Block), args.Error(1)
}

func (m *MockBlockClient) Update(ctx context.Context, blockID notionapi.BlockID, request *notionapi.BlockUpdateRequest) (notionapi.Block, error) {
	args := m.Called(ctx, blockID, request)
	return args.Get(0).(notionapi.Block), args.Error(1)
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
					PropBookmarkID: &notionapi.RichTextProperty{
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
					PropBookmarkID: &notionapi.RichTextProperty{
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
					PropBookmarkID: &notionapi.RichTextProperty{
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
					PropBookmarkID: &notionapi.RichTextProperty{
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

func TestGetPagesByBookName(t *testing.T) {
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
				ID: "page1",
				Properties: notionapi.Properties{
					PropBookTitle: &notionapi.TitleProperty{
						Title: []notionapi.RichText{
							{
								PlainText: "Book 1",
							},
						},
					},
				},
			},
			{
				ID: "page2",
				Properties: notionapi.Properties{
					PropBookTitle: &notionapi.TitleProperty{
						Title: []notionapi.RichText{
							{
								PlainText: "Book 2",
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
	bookPages, err := service.GetPagesByBookName("test-db-id")

	assert.NoError(t, err, "GetPagesByBookName should not return an error")
	assert.Equal(t, 2, len(bookPages), "Should return 2 book pages")
	assert.Equal(t, notionapi.PageID("page1"), bookPages["Book 1"], "Book 1 should map to page1")
	assert.Equal(t, notionapi.PageID("page2"), bookPages["Book 2"], "Book 2 should map to page2")

	mockDBClient.AssertExpectations(t)
}

func TestAddBookmark(t *testing.T) {
	setupLogger()
	defer logger.Close()

	// Create mock clients
	mockPageClient := new(MockPageClient)
	mockDBClient := new(MockDatabaseClient)

	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithPageClient(mockPageClient)
	service.WithDatabaseClient(mockDBClient)
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

	// Mock response for GetPagesByBookName (no existing pages)
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.Anything).Return(&notionapi.DatabaseQueryResponse{
		Results:    []notionapi.Page{},
		HasMore:    false,
		NextCursor: "",
	}, nil)

	// Configure mock for page creation
	mockPageClient.On("Create", mock.Anything, mock.Anything).Return(&notionapi.Page{
		ID: "test",
	}, nil)

	// Test the function
	_, err := service.AddBookmark("test-db-id", bookmark)

	assert.NoError(t, err, "AddBookmark should not return an error")
	mockPageClient.AssertExpectations(t)
	mockDBClient.AssertExpectations(t)

	// Verify the page creation request
	createCall := mockPageClient.Calls[0]
	req := createCall.Arguments.Get(1).(*notionapi.PageCreateRequest)

	// Check some aspects of the request
	titleProp, ok := req.Properties[PropBookTitle].(notionapi.TitleProperty)
	assert.True(t, ok, "Book Title should be a TitleProperty")
	assert.Contains(t, titleProp.Title[0].Text.Content, "test-volume-id")

	// Ensure children blocks were created for the bookmark
	assert.Greater(t, len(req.Children), 0, "Should have blocks for the bookmark content")
}

func TestAddBookmarksGroup(t *testing.T) {
	setupLogger()
	defer logger.Close()

	// Create mock clients
	mockDBClient := new(MockDatabaseClient)
	mockPageClient := new(MockPageClient)
	mockBlockClient := &MockBlockClient{}

	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithDatabaseClient(mockDBClient)
	service.WithPageClient(mockPageClient)
	service.WithBlockClient(mockBlockClient)
	service.WithContextFunc(func() context.Context {
		return context.Background()
	})

	// Create test bookmarks for the same book
	bookmarks := []kobo.Bookmark{
		{
			BookmarkID:  "test-bookmark-id-1",
			VolumeID:    "test-volume-id",
			Text:        "This is a test highlight 1",
			Annotation:  "This is a test annotation 1",
			Type:        "highlight",
			DateCreated: "2023-01-01T12:00:00Z",
			Color:       "2",
		},
		{
			BookmarkID:  "test-bookmark-id-2",
			VolumeID:    "test-volume-id",
			Text:        "This is a test highlight 2",
			Annotation:  "This is a test annotation 2",
			Type:        "highlight",
			DateCreated: "2023-01-01T12:00:00Z",
			Color:       "3",
		},
	}

	// Mock response for GetPagesByBookName (no existing pages)
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.Anything).Return(&notionapi.DatabaseQueryResponse{
		Results:    []notionapi.Page{},
		HasMore:    false,
		NextCursor: "",
	}, nil)

	// Mock page creation
	mockPageClient.On("Create", mock.Anything, mock.Anything).Return(&notionapi.Page{
		ID: "new-page",
	}, nil)

	// Test the function
	err := service.AddBookmarks("test-db-id", bookmarks)

	assert.NoError(t, err, "AddBookmarks should not return an error")
	mockDBClient.AssertExpectations(t)
	mockPageClient.AssertExpectations(t)

	// Verify the page creation request
	createCall := mockPageClient.Calls[0]
	req := createCall.Arguments.Get(1).(*notionapi.PageCreateRequest)

	// Check request properties
	titleProp, ok := req.Properties[PropBookTitle].(notionapi.TitleProperty)
	assert.True(t, ok, "Book Title should be a TitleProperty")
	assert.Contains(t, titleProp.Title[0].Text.Content, "test-volume-id")

	// Should have multiple children blocks
	assert.Greater(t, len(req.Children), 2, "Should have multiple blocks")
}

func TestAddBookmarksGroupExistingPage(t *testing.T) {
	setupLogger()
	defer logger.Close()

	// Create mock clients
	mockDBClient := new(MockDatabaseClient)
	mockPageClient := new(MockPageClient)
	mockBlockClient := &MockBlockClient{}

	// Create a test service with our mock
	service := notion.NewNotionService("test-token")
	service.WithDatabaseClient(mockDBClient)
	service.WithPageClient(mockPageClient)
	service.WithBlockClient(mockBlockClient)
	service.WithContextFunc(func() context.Context {
		return context.Background()
	})

	// Create test bookmarks for the same book
	bookmarks := []kobo.Bookmark{
		{
			BookmarkID:  "test-bookmark-id-1",
			VolumeID:    "test-volume-id",
			Text:        "This is a test highlight 1",
			Annotation:  "This is a test annotation 1",
			Type:        "highlight",
			DateCreated: "2023-01-01T12:00:00Z",
		},
	}

	// Mock response for GetPagesByBookName (existing page)
	mockDBClient.On("Query", mock.Anything, notionapi.DatabaseID("test-db-id"), mock.Anything).Return(&notionapi.DatabaseQueryResponse{
		Results: []notionapi.Page{
			{
				ID: "existing-page",
				Properties: notionapi.Properties{
					PropBookTitle: &notionapi.TitleProperty{
						Title: []notionapi.RichText{
							{
								PlainText: "test-volume-id",
							},
						},
					},
				},
			},
		},
		HasMore:    false,
		NextCursor: "",
	}, nil)

	// Mock getting existing blocks
	mockBlockClient.On("GetChildren", mock.Anything, notionapi.BlockID("existing-page"), mock.Anything).Return(&notionapi.GetChildrenResponse{
		Results:    []notionapi.Block{},
		HasMore:    false,
		NextCursor: "",
	}, nil)

	// Mock appending blocks
	mockBlockClient.On("AppendChildren", mock.Anything, notionapi.BlockID("existing-page"), mock.Anything).Return(&notionapi.AppendBlockChildrenResponse{}, nil)

	// Mock getting the page
	mockPageClient.On("Get", mock.Anything, notionapi.PageID("existing-page")).Return(&notionapi.Page{
		ID: "existing-page",
	}, nil)

	// Test the function
	err := service.AddBookmarks("test-db-id", bookmarks)

	assert.NoError(t, err, "AddBookmarks should not return an error")
	mockDBClient.AssertExpectations(t)
	mockBlockClient.AssertExpectations(t)
	mockPageClient.AssertExpectations(t)
}
