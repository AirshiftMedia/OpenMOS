package service

import (
	"airshift/openmos/internal/model"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
	"context"
	"fmt"
	"time"
)

// processStoryBody processes the story body and extracts items
func (s *MOSService) processStoryBody(ctx context.Context, story *model.Story, storyBody *xml.StoryBody) error {
	// Find all items in the story body
	var items []*model.Item

	// Iterate through paragraphs
	for i, paragraph := range storyBody.Paragraphs {
		// Process each story item
		for j, storyItem := range paragraph.Items {
			// Create a new item
			item := &model.Item{
				ID:        fmt.Sprintf("%s_I%d_%d", story.ID, i, j),
				StoryID:   story.ID,
				Slug:      storyItem.ItemSlug,
				ObjectID:  storyItem.ObjID,
				Duration:  storyItem.ItemEdDur,
				Status:    model.StatusPending,
				Order:     j + 1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			items = append(items, item)
		}
	}

	// Handle item creation/update (will be implemented in a later step)
	// For now, just log what we found
	logger.Infof("Found %d items in story %s", len(items), story.ID)

	return nil
}
