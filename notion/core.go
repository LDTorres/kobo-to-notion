package notion

import (
	"context"
	"kobo-to-notion/utils"
	"net/http"

	"github.com/jomei/notionapi"
)

// Constants for property names
const (
	PropBookTitle       = "Book Title"
	PropHighlightedText = "Highlighted Text"
	PropAnnotation      = "Annotation"
	PropType            = "Type"
	PropDateCreated     = "Date Created"
	PropBookmarkID      = "Bookmark ID"
	PropBookName        = "Book Name"

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
	client      *notionapi.Client
	dbClient    NotionDatabaseClient
	pageClient  NotionPageClient
	blockClient notionapi.BlockService // Using the actual BlockService from the API
	contextFunc func() context.Context
}

// NewNotionService creates a new NotionService
func NewNotionService(notionToken string) *NotionService {
	client := notionapi.NewClient(notionapi.Token(notionToken))
	return &NotionService{
		client:      client,
		dbClient:    client.Database, // Use client's DB interface
		pageClient:  client.Page,     // Use client's Page interface
		blockClient: client.Block,    // Use client's Block interface
		contextFunc: context.Background,
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

// Global service instance
var defaultService *NotionService

// InitializeNotionClient initializes the global NotionClient
func InitializeNotionClient(certPath string, notionToken string, databaseID string) error {
	defaultService = NewNotionService(notionToken)
	return defaultService.InitializeWithCert(certPath)
}
