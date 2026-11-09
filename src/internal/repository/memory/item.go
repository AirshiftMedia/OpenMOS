package memory

import (
	"context"
	"fmt"
	"slices"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type items struct{ r *Repository }

func (s items) Create(_ context.Context, item *model.Item) (*model.Item, error) {
	if err := checkItemKey(item.RunningOrderID, item.StoryID, item.ID); err != nil {
		return nil, err
	}
	defer s.r.lockWrite()()
	k := itemKey{item.RunningOrderID, item.StoryID, item.ID}
	if _, ok := s.r.items[k]; ok {
		return nil, fmt.Errorf("item %s/%s/%s: %w", k.ro, k.story, k.item, repository.ErrAlreadyExists)
	}
	if err := s.checkParentLocked(k.ro, k.story); err != nil {
		return nil, err
	}
	s.insertLocked(item, s.nextOrderLocked(k.ro, k.story))
	return cloneItem(item), nil
}

func (s items) Get(_ context.Context, roID, storyID, itemID string) (*model.Item, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	item, ok := s.r.items[itemKey{roID, storyID, itemID}]
	if !ok {
		return nil, fmt.Errorf("item %s/%s/%s: %w", roID, storyID, itemID, repository.ErrNotFound)
	}
	return cloneItem(item), nil
}

func (s items) Update(_ context.Context, item *model.Item) error {
	defer s.r.lockWrite()()
	k := itemKey{item.RunningOrderID, item.StoryID, item.ID}
	current, ok := s.r.items[k]
	if !ok {
		return fmt.Errorf("item %s/%s/%s: %w", k.ro, k.story, k.item, repository.ErrNotFound)
	}
	item.Order = current.Order
	item.CreatedAt = current.CreatedAt
	item.UpdatedAt = now()
	s.r.items[k] = cloneItem(item)
	return nil
}

func (s items) Delete(_ context.Context, roID, storyID, itemID string) error {
	defer s.r.lockWrite()()
	k := itemKey{roID, storyID, itemID}
	if _, ok := s.r.items[k]; !ok {
		return fmt.Errorf("item %s/%s/%s: %w", roID, storyID, itemID, repository.ErrNotFound)
	}
	delete(s.r.items, k)
	return nil
}

func (s items) ListByStory(_ context.Context, roID, storyID string) ([]*model.Item, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return s.listLocked(roID, storyID), nil
}

func (s items) Move(_ context.Context, roID, storyID string, itemIDs []string, beforeID string) error {
	defer s.r.lockWrite()()
	order, err := repository.MoveIDs(s.idsLocked(roID, storyID), itemIDs, beforeID)
	if err != nil {
		return fmt.Errorf("move items in %s/%s: %w", roID, storyID, err)
	}
	s.setOrderLocked(roID, storyID, order)
	return nil
}

func (s items) Swap(_ context.Context, roID, storyID, itemID1, itemID2 string) error {
	defer s.r.lockWrite()()
	order, err := repository.SwapIDs(s.idsLocked(roID, storyID), itemID1, itemID2)
	if err != nil {
		return fmt.Errorf("swap items in %s/%s: %w", roID, storyID, err)
	}
	s.setOrderLocked(roID, storyID, order)
	return nil
}

func (s items) ReplaceAll(_ context.Context, roID, storyID string, list []*model.Item) error {
	ids := make([]string, len(list))
	for i, item := range list {
		if err := checkItemKey(roID, storyID, item.ID); err != nil {
			return err
		}
		ids[i] = item.ID
	}
	if err := repository.CheckDistinct(ids); err != nil {
		return fmt.Errorf("replace items in %s/%s: %w", roID, storyID, err)
	}
	defer s.r.lockWrite()()
	if err := s.checkParentLocked(roID, storyID); err != nil {
		return err
	}
	for k := range s.r.items {
		if k.ro == roID && k.story == storyID {
			delete(s.r.items, k)
		}
	}
	for i, item := range list {
		item.RunningOrderID, item.StoryID = roID, storyID
		s.insertLocked(item, i)
	}
	return nil
}

// insertLocked stores a new item at the given Order. The caller holds s.r.mu.
func (s items) insertLocked(item *model.Item, order int) {
	t := now()
	item.Order, item.CreatedAt, item.UpdatedAt = order, t, t
	s.r.items[itemKey{item.RunningOrderID, item.StoryID, item.ID}] = cloneItem(item)
}

func (s items) listLocked(roID, storyID string) []*model.Item {
	return sorted(s.r.items,
		func(item *model.Item) bool { return item.RunningOrderID == roID && item.StoryID == storyID },
		cloneItem,
		func(a, b *model.Item) int { return byOrder(a.Order, b.Order, a.ID, b.ID) },
	)
}

func (s items) idsLocked(roID, storyID string) []string {
	list := s.listLocked(roID, storyID)
	ids := make([]string, len(list))
	for i, item := range list {
		ids[i] = item.ID
	}
	return ids
}

func (s items) nextOrderLocked(roID, storyID string) int {
	var orders []int
	for k, item := range s.r.items {
		if k.ro == roID && k.story == storyID {
			orders = append(orders, item.Order)
		}
	}
	return nextOrder(orders)
}

// setOrderLocked renumbers the items of a story to follow ids.
func (s items) setOrderLocked(roID, storyID string, ids []string) {
	for k, item := range s.r.items {
		if k.ro == roID && k.story == storyID {
			item.Order = slices.Index(ids, k.item)
		}
	}
}

func (s items) checkParentLocked(roID, storyID string) error {
	if _, ok := s.r.stories[storyKey{roID, storyID}]; !ok {
		return fmt.Errorf("story %s/%s: %w", roID, storyID, repository.ErrNotFound)
	}
	return nil
}

func checkItemKey(roID, storyID, itemID string) error {
	if err := checkStoryKey(roID, storyID); err != nil {
		return err
	}
	return repository.CheckID("item", itemID)
}
