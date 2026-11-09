package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
)

// createNewStory creates a new story from the provided ROStorySend, creating
// its running order if needed, and returns the story's ID.
func (s *MOSService) createNewStory(ctx context.Context, tx repository.Repository, storySend xml.ROStorySend) (string, error) {
	if storySend.ROID == "" {
		return "", fmt.Errorf("no running order ID specified")
	}
	_, err := tx.RunningOrders().Get(ctx, storySend.ROID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		ro := &model.RunningOrder{
			ID:     storySend.ROID,
			Slug:   "Auto-created RO",
			Status: model.StatusPending,
		}
		if _, err := tx.RunningOrders().Create(ctx, ro); err != nil {
			return "", fmt.Errorf("failed to create running order: %w", err)
		}
	case err != nil:
		return "", fmt.Errorf("failed to get running order: %w", err)
	}

	// Create a story ID if not provided
	storyID := storySend.StoryID
	if storyID == "" {
		storyID = fmt.Sprintf("S%d", time.Now().UnixNano())
	}

	story := &model.Story{
		ID:             storyID,
		RunningOrderID: storySend.ROID,
		Slug:           storySend.StorySlug,
		Number:         storySend.StoryNum,
		Status:         model.StatusPending,
	}

	if err := s.processStoryBody(ctx, story, &storySend.StoryBody); err != nil {
		return "", fmt.Errorf("failed to process story body: %w", err)
	}

	// The repository appends the new story to the running order
	if _, err := tx.Stories().Create(ctx, story); err != nil {
		return "", fmt.Errorf("failed to create story: %w", err)
	}

	logger.Infof("Created new story %s in running order %s", story.ID, storySend.ROID)
	return story.ID, nil
}
