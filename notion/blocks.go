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

// createBookmarkBlocks creates a set of blocks for a bookmark
func (s *NotionService) createBookmarkBlocks(bookmark kobo.Bookmark) []notionapi.Block {
	colorsMap := map[string]notionapi.Color{
		"0": notionapi.ColorRed,
		"1": notionapi.ColorOrange,
		"2": notionapi.ColorYellow,
		"3": notionapi.ColorGreen,
		"4": notionapi.ColorBlue,
		"5": notionapi.ColorPurple,
		"6": notionapi.ColorPink,
		"7": notionapi.ColorBrown,
	}

	if bookmark.Text == "" && bookmark.Annotation == "" {
		return []notionapi.Block{}
	}

	// Split text into chunks of 2000 characters
	const bookMarkTextSplit = 2000
	textChunks := utils.SplitText(bookmark.Text, bookMarkTextSplit)
	annotationChunks := utils.SplitText(bookmark.Annotation, bookMarkTextSplit)

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
