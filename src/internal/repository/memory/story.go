package memory

import (
	"context"
	"errors"
	"fmt"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type stories struct{ r *Repository }

func (s stories) Create(_ context.Context, story *model.Story) (*model.Story, error) {
	if story.ID == "" {
		return nil, errors.New("story ID is required")
	}
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.stories[story.ID]; ok {
		return nil, fmt.Errorf("story %s: %w", story.ID, repository.ErrAlreadyExists)
	}
	if err := s.checkParentLocked(story); err != nil {
		return nil, err
	}
	t := now()
	story.CreatedAt, story.UpdatedAt = t, t
	s.r.stories[story.ID] = cloneStory(story)
	return cloneStory(story), nil
}

func (s stories) Get(_ context.Context, id string) (*model.Story, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	story, ok := s.r.stories[id]
	if !ok {
		return nil, fmt.Errorf("story %s: %w", id, repository.ErrNotFound)
	}
	return cloneStory(story), nil
}

func (s stories) Update(_ context.Context, story *model.Story) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	current, ok := s.r.stories[story.ID]
	if !ok {
		return fmt.Errorf("story %s: %w", story.ID, repository.ErrNotFound)
	}
	if err := s.checkParentLocked(story); err != nil {
		return err
	}
	story.CreatedAt = current.CreatedAt
	story.UpdatedAt = now()
	s.r.stories[story.ID] = cloneStory(story)
	return nil
}

func (s stories) Delete(_ context.Context, id string) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.stories[id]; !ok {
		return fmt.Errorf("story %s: %w", id, repository.ErrNotFound)
	}
	s.r.deleteStoryLocked(id)
	return nil
}

func (s stories) ListByRunningOrder(_ context.Context, roID string) ([]*model.Story, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return sorted(s.r.stories,
		func(story *model.Story) bool { return story.RunningOrderID == roID },
		cloneStory,
		func(a, b *model.Story) int { return byOrder(a.Order, b.Order, a.ID, b.ID) },
	), nil
}

// checkParentLocked returns ErrNotFound if the story's running order does
// not exist. The caller holds s.r.mu.
func (s stories) checkParentLocked(story *model.Story) error {
	if _, ok := s.r.runningOrders[story.RunningOrderID]; !ok {
		return fmt.Errorf("running order %s of story %s: %w",
			story.RunningOrderID, story.ID, repository.ErrNotFound)
	}
	return nil
}
