package memory

import (
	"context"
	"fmt"
	"slices"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type stories struct{ r *Repository }

func (s stories) Create(_ context.Context, story *model.Story) (*model.Story, error) {
	if err := checkStoryKey(story.RunningOrderID, story.ID); err != nil {
		return nil, err
	}
	defer s.r.lockWrite()()
	k := storyKey{story.RunningOrderID, story.ID}
	if _, ok := s.r.stories[k]; ok {
		return nil, fmt.Errorf("story %s/%s: %w", k.ro, k.story, repository.ErrAlreadyExists)
	}
	if err := s.checkParentLocked(k.ro); err != nil {
		return nil, err
	}
	s.insertLocked(story, s.nextOrderLocked(k.ro))
	return cloneStory(story), nil
}

func (s stories) Get(_ context.Context, roID, storyID string) (*model.Story, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	story, ok := s.r.stories[storyKey{roID, storyID}]
	if !ok {
		return nil, fmt.Errorf("story %s/%s: %w", roID, storyID, repository.ErrNotFound)
	}
	return cloneStory(story), nil
}

func (s stories) Update(_ context.Context, story *model.Story) error {
	defer s.r.lockWrite()()
	k := storyKey{story.RunningOrderID, story.ID}
	current, ok := s.r.stories[k]
	if !ok {
		return fmt.Errorf("story %s/%s: %w", k.ro, k.story, repository.ErrNotFound)
	}
	story.Order = current.Order
	story.CreatedAt = current.CreatedAt
	story.UpdatedAt = now()
	s.r.stories[k] = cloneStory(story)
	return nil
}

func (s stories) Delete(_ context.Context, roID, storyID string) error {
	defer s.r.lockWrite()()
	k := storyKey{roID, storyID}
	if _, ok := s.r.stories[k]; !ok {
		return fmt.Errorf("story %s/%s: %w", roID, storyID, repository.ErrNotFound)
	}
	s.r.deleteStoryLocked(k)
	return nil
}

func (s stories) ListByRunningOrder(_ context.Context, roID string) ([]*model.Story, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return s.listLocked(roID), nil
}

func (s stories) Move(_ context.Context, roID string, storyIDs []string, beforeID string) error {
	defer s.r.lockWrite()()
	order, err := repository.MoveIDs(s.idsLocked(roID), storyIDs, beforeID)
	if err != nil {
		return fmt.Errorf("move stories in %s: %w", roID, err)
	}
	s.setOrderLocked(roID, order)
	return nil
}

func (s stories) Swap(_ context.Context, roID, storyID1, storyID2 string) error {
	defer s.r.lockWrite()()
	order, err := repository.SwapIDs(s.idsLocked(roID), storyID1, storyID2)
	if err != nil {
		return fmt.Errorf("swap stories in %s: %w", roID, err)
	}
	s.setOrderLocked(roID, order)
	return nil
}

func (s stories) ReplaceAll(_ context.Context, roID string, list []*model.Story) error {
	ids := make([]string, len(list))
	for i, story := range list {
		if err := checkStoryKey(roID, story.ID); err != nil {
			return err
		}
		ids[i] = story.ID
	}
	if err := repository.CheckDistinct(ids); err != nil {
		return fmt.Errorf("replace stories in %s: %w", roID, err)
	}
	defer s.r.lockWrite()()
	if err := s.checkParentLocked(roID); err != nil {
		return err
	}
	for k := range s.r.stories {
		if k.ro == roID {
			s.r.deleteStoryLocked(k)
		}
	}
	for i, story := range list {
		story.RunningOrderID = roID
		s.insertLocked(story, i)
	}
	return nil
}

// insertLocked stores a new story at the given Order. The caller holds s.r.mu.
func (s stories) insertLocked(story *model.Story, order int) {
	t := now()
	story.Order, story.CreatedAt, story.UpdatedAt = order, t, t
	s.r.stories[storyKey{story.RunningOrderID, story.ID}] = cloneStory(story)
}

func (s stories) listLocked(roID string) []*model.Story {
	return sorted(s.r.stories,
		func(story *model.Story) bool { return story.RunningOrderID == roID },
		cloneStory,
		func(a, b *model.Story) int { return byOrder(a.Order, b.Order, a.ID, b.ID) },
	)
}

func (s stories) idsLocked(roID string) []string {
	list := s.listLocked(roID)
	ids := make([]string, len(list))
	for i, story := range list {
		ids[i] = story.ID
	}
	return ids
}

func (s stories) nextOrderLocked(roID string) int {
	var orders []int
	for k, story := range s.r.stories {
		if k.ro == roID {
			orders = append(orders, story.Order)
		}
	}
	return nextOrder(orders)
}

// setOrderLocked renumbers the stories of roID to follow ids.
func (s stories) setOrderLocked(roID string, ids []string) {
	for k, story := range s.r.stories {
		if k.ro == roID {
			story.Order = slices.Index(ids, k.story)
		}
	}
}

func (s stories) checkParentLocked(roID string) error {
	if _, ok := s.r.runningOrders[roID]; !ok {
		return fmt.Errorf("running order %s: %w", roID, repository.ErrNotFound)
	}
	return nil
}

func checkStoryKey(roID, storyID string) error {
	if err := repository.CheckID("running order", roID); err != nil {
		return err
	}
	return repository.CheckID("story", storyID)
}
