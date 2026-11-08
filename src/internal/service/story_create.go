package service

import (
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/model"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
	"context"
	"fmt"
	"time"
)

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
