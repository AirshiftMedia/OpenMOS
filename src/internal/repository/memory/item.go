package memory

import (
	"context"
	"errors"
	"fmt"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type items struct{ r *Repository }

func (s items) Create(_ context.Context, item *model.Item) (*model.Item, error) {
	if item.ID == "" {
		return nil, errors.New("item ID is required")
	}
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.items[item.ID]; ok {
		return nil, fmt.Errorf("item %s: %w", item.ID, repository.ErrAlreadyExists)
	}
	if err := s.checkParentLocked(item); err != nil {
		return nil, err
	}
	t := now()
	item.CreatedAt, item.UpdatedAt = t, t
	s.r.items[item.ID] = cloneItem(item)
	return cloneItem(item), nil
}

func (s items) Get(_ context.Context, id string) (*model.Item, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	item, ok := s.r.items[id]
	if !ok {
		return nil, fmt.Errorf("item %s: %w", id, repository.ErrNotFound)
	}
	return cloneItem(item), nil
}

func (s items) Update(_ context.Context, item *model.Item) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	current, ok := s.r.items[item.ID]
	if !ok {
		return fmt.Errorf("item %s: %w", item.ID, repository.ErrNotFound)
	}
	if err := s.checkParentLocked(item); err != nil {
		return err
	}
	item.CreatedAt = current.CreatedAt
	item.UpdatedAt = now()
	s.r.items[item.ID] = cloneItem(item)
	return nil
}

func (s items) Delete(_ context.Context, id string) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.items[id]; !ok {
		return fmt.Errorf("item %s: %w", id, repository.ErrNotFound)
	}
	delete(s.r.items, id)
	return nil
}

func (s items) ListByStory(_ context.Context, storyID string) ([]*model.Item, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return sorted(s.r.items,
		func(item *model.Item) bool { return item.StoryID == storyID },
		cloneItem,
		func(a, b *model.Item) int { return byOrder(a.Order, b.Order, a.ID, b.ID) },
	), nil
}

// checkParentLocked returns ErrNotFound if the item's story does not exist.
// The caller holds s.r.mu.
func (s items) checkParentLocked(item *model.Item) error {
	if _, ok := s.r.stories[item.StoryID]; !ok {
		return fmt.Errorf("story %s of item %s: %w", item.StoryID, item.ID, repository.ErrNotFound)
	}
	return nil
}
