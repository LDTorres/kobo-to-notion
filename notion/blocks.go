package notion

import (
	"kobo-to-notion/kobo"
	"kobo-to-notion/utils"

	"github.com/jomei/notionapi"
)

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

		// Append blocks ids to the blocks slice
		blocks = append(blocks, resp.Results...)

		if !resp.HasMore {
			break
		}

		startCursor = notionapi.Cursor(resp.NextCursor)
	}

	return blocks, nil
}

func getColorsMap() map[string]notionapi.Color {
	return map[string]notionapi.Color{
		"0": notionapi.ColorOrange,
		"1": notionapi.ColorPurple,
		"2": notionapi.ColorBlue,
		"3": notionapi.ColorGreen,
		"4": notionapi.ColorRed,
	}
}

// createBookmarkBlocks creates a set of blocks for a bookmark
func (s *NotionService) createBookmarkTextBlocks(bookmark kobo.Bookmark) []notionapi.Block {
	colorsMap := getColorsMap();

	if bookmark.Text == "" {
		return []notionapi.Block{}
	}

	// Split text into chunks of 2000 characters
	const bookMarkTextSplit = 2000
	textChunks := utils.SplitText(bookmark.Text, bookMarkTextSplit)

	// Create notionapi.RichText blocks for each chunk
	paragraphBlocks := []notionapi.RichText{}

	if len(textChunks) > 0 {
		paragraphBlocks = append(paragraphBlocks, notionapi.RichText{
			Type: notionapi.ObjectTypeText,
			Text: &notionapi.Text{
				Content: PropHighlightedText,
			},
			Annotations: &notionapi.Annotations{
				Bold:  true,
				Color: colorsMap[bookmark.Color],
			},
		}, notionapi.RichText{
			Type: notionapi.ObjectTypeText,
			Text: &notionapi.Text{
				Content: `
`,
			},
		})

		for _, chunk := range textChunks {
			paragraphBlocks = append(paragraphBlocks, notionapi.RichText{
				Type: notionapi.ObjectTypeText,
				Text: &notionapi.Text{
					Content: chunk,
				},
			})
		}
	}

	blocks := []notionapi.Block{}

	if len(paragraphBlocks) > 0 {
		blocks = append(blocks, &notionapi.QuoteBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeQuote,
			},
			Quote: notionapi.Quote{
				RichText: paragraphBlocks,
			},
		})
	}

	return blocks
}

func (s *NotionService) createBookmarkAnnotationBlocks(bookmark kobo.Bookmark) []notionapi.Block {
	colorsMap := getColorsMap();
	
	if bookmark.Annotation == "" {
		return []notionapi.Block{}
	}

	// Split text into chunks of 2000 characters
	const bookMarkTextSplit = 2000
	annotationChunks := utils.SplitText(bookmark.Annotation, bookMarkTextSplit)

	blocks := []notionapi.Block{}

	// Create notionapi.RichText blocks for each chunk
	annotationBlocks := []notionapi.RichText{}

	if len(annotationChunks) > 0 {
		annotationBlocks = append(annotationBlocks, notionapi.RichText{
			Type: notionapi.ObjectTypeText,
			Text: &notionapi.Text{
				Content: PropAnnotation,
			},
			Annotations: &notionapi.Annotations{
				Bold:  true,
				Color: colorsMap[bookmark.Color],
			},
		}, notionapi.RichText{
			Type: notionapi.ObjectTypeText,
			Text: &notionapi.Text{
				Content: `
`,
			},
		})

		for _, chunk := range annotationChunks {
			annotationBlocks = append(annotationBlocks, notionapi.RichText{
				Type: notionapi.ObjectTypeText,
				Text: &notionapi.Text{
					Content: chunk,
				},
			})
		}
	}

	if len(annotationBlocks) > 0 {
		blocks = append(blocks,
			&notionapi.QuoteBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeQuote,
				},
				Quote: notionapi.Quote{
					RichText: annotationBlocks,
				},
			},
		)
	}

	return blocks
}
