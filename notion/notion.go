package notion

import (
	"context"
	"kobo-to-notion/kobo"
	"kobo-to-notion/logger"
	"kobo-to-notion/utils"

	"github.com/jomei/notionapi"
)

// Fetch all BookmarkIDs from Notion
func GetNotionBookmarkIDs(client *notionapi.Client, databaseID string) (map[string]bool, error) {
	existingBookmarks := make(map[string]bool)
	var startCursor notionapi.Cursor

	for {
		query := &notionapi.DatabaseQueryRequest{}

		// Add cursor only if it's not empty
		if startCursor != "" {
			query.StartCursor = startCursor
		}

		res, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseID), query)
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

// Add a new bookmark to Notion
func AddBookmarkToNotion(client *notionapi.Client, databaseID string, bookmark kobo.Bookmark) error {
    bookName := utils.GetBookNameFromVolumeID(bookmark.VolumeID)
    emoji := notionapi.Emoji("📚")
    
    parsedDate, err := utils.ParseKoboBookmarkDate(bookmark.DateCreated)
    
    if err != nil {
        return err
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
                                Bold: true,
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
                                Bold: true,
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
                        Type: "emoji",
                        Emoji: &emoji,
                    },
                    Color: string(notionapi.ColorGray),
                },
            },
        },
    }

	_, err = client.Page.Create(context.Background(), payload)

    if err != nil {
		logger.Logger.Fatalf("Create page error: %v", err)
	}

	logger.Logger.Println("Page created successfully!")

    return err
}