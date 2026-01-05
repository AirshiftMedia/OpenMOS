package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"airshift/openmos/internal/events"
	"airshift/openmos/internal/model"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
)

// ProcessStoryAction processes a story action request from an NCS
func (s *MOSService) ProcessStoryAction(ctx context.Context, action xml.NCSReqStoryAction) error {
	logger.Infof("Processing story action: %s", action.Operation)

	switch strings.ToUpper(action.Operation) {
	case "NEW":
		return s.createNewStory(ctx, action.ROStorySend)
	case "UPDATE":
		return s.updateStory(ctx, action.ROStorySend)
	case "REPLACE":
		return s.replaceStory(ctx, action.ROStorySend)
	default:
		return fmt.Errorf("unsupported story operation: %s", action.Operation)
	}
}

// createNewStory creates a new story from the provided ROStorySend
func (s *MOSService) createNewStory(ctx context.Context, storySend xml.ROStorySend) error {
	// Check if Running Order exists, create if not
	var ro *model.RunningOrder
	var err error

	if storySend.ROID != "" {
		ro, err = s.runningOrderRepo.Get(ctx, storySend.ROID)
		if err != nil {
			// Create a new running order if it doesn't exist
			ro = &model.RunningOrder{
				ID:        storySend.ROID,
				Slug:      "Auto-created RO",
				Status:    model.StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			}

			ro, err = s.runningOrderRepo.Create(ctx, ro)
			if err != nil {
				return fmt.Errorf("failed to create running order: %w", err)
			}
		}
	} else {
		return fmt.Errorf("no running order ID specified")
	}

	// Create a story ID if not provided
	storyID := storySend.StoryID
	if storyID == "" {
		storyID = fmt.Sprintf("S%d", time.Now().UnixNano())
	}

	// Create the new story
	story := &model.Story{
		ID:             storyID,
		RunningOrderID: ro.ID,
		Slug:           storySend.StorySlug,
		Number:         storySend.StoryNum,
		Status:         model.StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Try to determine the story's order in the running order
	stories, err := s.storyRepo.ListByRunningOrder(ctx, ro.ID)
	if err != nil {
		return fmt.Errorf("failed to list stories: %w", err)
	}

	// New story goes at the end
	story.Order = len(stories) + 1

	// Process story body
	err = s.processStoryBody(ctx, story, &storySend.StoryBody)
	if err != nil {
		return fmt.Errorf("failed to process story body: %w", err)
	}

	// Create the story
	_, err = s.storyRepo.Create(ctx, story)
	if err != nil {
		return fmt.Errorf("failed to create story: %w", err)
	}

	// Publish event after successful creation
	if s.eventBus != nil {
		s.eventBus.Publish(events.Event{
			Type:    events.StoryModified,
			Payload: story.ID,
			Source:  "mos_service",
		})
	}

	logger.Infof("Created new story %s in running order %s", story.ID, ro.ID)
	return nil
}

// updateStory updates an existing story from the provided ROStorySend
func (s *MOSService) updateStory(ctx context.Context, storySend xml.ROStorySend) error {
	// Check if the story exists
	story, err := s.storyRepo.Get(ctx, storySend.StoryID)
	if err != nil {
		return fmt.Errorf("story not found: %w", err)
	}

	// Update story fields
	story.Slug = storySend.StorySlug
	story.Number = storySend.StoryNum
	story.UpdatedAt = time.Now()

	// Process story body
	err = s.processStoryBody(ctx, story, &storySend.StoryBody)
	if err != nil {
		return fmt.Errorf("failed to process story body: %w", err)
	}

	// Update the story
	err = s.storyRepo.Update(ctx, story)
	if err != nil {
		return fmt.Errorf("failed to update story: %w", err)
	}

	// Publish event after successful update
	if s.eventBus != nil {
		s.eventBus.Publish(events.Event{
			Type:    events.StoryModified,
			Payload: story.ID,
			Source:  "mos_service",
		})
	}

	logger.Infof("Updated story %s in running order %s", story.ID, story.RunningOrderID)
	return nil
}

// replaceStory replaces an existing story from the provided ROStorySend
func (s *MOSService) replaceStory(ctx context.Context, storySend xml.ROStorySend) error {
	// For now, implement as delete + create
	// Delete existing story
	err := s.storyRepo.Delete(ctx, storySend.StoryID)
	if err != nil {
		return fmt.Errorf("failed to delete existing story: %w", err)
	}

	// Create new story
	return s.createNewStory(ctx, storySend)
}

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
