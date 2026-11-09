package service

import (
	"context"
	"fmt"
	"strings"

	"airshift/openmos/internal/events"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
)

// ProcessStoryAction processes a story action request from an NCS. The
// action runs in one transaction.
func (s *MOSService) ProcessStoryAction(ctx context.Context, action xml.NCSReqStoryAction) error {
	logger.Infof("Processing story action: %s", action.Operation)

	var apply func(context.Context, repository.Repository, xml.ROStorySend) (string, error)
	switch strings.ToUpper(action.Operation) {
	case "NEW":
		apply = s.createNewStory
	case "UPDATE":
		apply = s.updateStory
	case "REPLACE":
		apply = s.replaceStory
	default:
		return fmt.Errorf("unsupported story operation: %s", action.Operation)
	}

	var storyID string
	err := s.repo.InTx(ctx, func(tx repository.Repository) error {
		var err error
		storyID, err = apply(ctx, tx, action.ROStorySend)
		return err
	})
	if err != nil {
		return err
	}
	s.publish(events.StoryModified, storyID)
	return nil
}

// updateStory updates an existing story from the provided ROStorySend and
// returns its ID.
func (s *MOSService) updateStory(ctx context.Context, tx repository.Repository, storySend xml.ROStorySend) (string, error) {
	story, err := tx.Stories().Get(ctx, storySend.ROID, storySend.StoryID)
	if err != nil {
		return "", fmt.Errorf("story not found: %w", err)
	}

	story.Slug = storySend.StorySlug
	story.Number = storySend.StoryNum

	if err := s.processStoryBody(ctx, story, &storySend.StoryBody); err != nil {
		return "", fmt.Errorf("failed to process story body: %w", err)
	}
	if err := tx.Stories().Update(ctx, story); err != nil {
		return "", fmt.Errorf("failed to update story: %w", err)
	}

	logger.Infof("Updated story %s in running order %s", story.ID, story.RunningOrderID)
	return story.ID, nil
}

// replaceStory replaces an existing story from the provided ROStorySend and
// returns its ID.
func (s *MOSService) replaceStory(ctx context.Context, tx repository.Repository, storySend xml.ROStorySend) (string, error) {
	// For now, implement as delete + create
	if err := tx.Stories().Delete(ctx, storySend.ROID, storySend.StoryID); err != nil {
		return "", fmt.Errorf("failed to delete existing story: %w", err)
	}
	return s.createNewStory(ctx, tx, storySend)
}
