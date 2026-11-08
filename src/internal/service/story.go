package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"airshift/openmos/internal/events"
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
