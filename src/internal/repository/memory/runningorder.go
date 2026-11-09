package memory

import (
	"context"
	"fmt"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type runningOrders struct{ r *Repository }

func (s runningOrders) Create(_ context.Context, ro *model.RunningOrder) (*model.RunningOrder, error) {
	if err := repository.CheckID("running order", ro.ID); err != nil {
		return nil, err
	}
	defer s.r.lockWrite()()
	if _, ok := s.r.runningOrders[ro.ID]; ok {
		return nil, fmt.Errorf("running order %s: %w", ro.ID, repository.ErrAlreadyExists)
	}
	t := now()
	ro.CreatedAt, ro.UpdatedAt, ro.Version = t, t, 1
	s.r.runningOrders[ro.ID] = cloneRunningOrder(ro)
	return cloneRunningOrder(ro), nil
}

func (s runningOrders) Get(_ context.Context, roID string) (*model.RunningOrder, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	ro, ok := s.r.runningOrders[roID]
	if !ok {
		return nil, fmt.Errorf("running order %s: %w", roID, repository.ErrNotFound)
	}
	return cloneRunningOrder(ro), nil
}

func (s runningOrders) Update(_ context.Context, ro *model.RunningOrder) error {
	defer s.r.lockWrite()()
	current, ok := s.r.runningOrders[ro.ID]
	if !ok {
		return fmt.Errorf("running order %s: %w", ro.ID, repository.ErrNotFound)
	}
	if ro.Version != current.Version {
		return fmt.Errorf("running order %s at version %d, update based on %d: %w",
			ro.ID, current.Version, ro.Version, repository.ErrVersionConflict)
	}
	ro.Version++
	ro.CreatedAt = current.CreatedAt
	ro.UpdatedAt = now()
	s.r.runningOrders[ro.ID] = cloneRunningOrder(ro)
	return nil
}

func (s runningOrders) Delete(_ context.Context, roID string) error {
	defer s.r.lockWrite()()
	if _, ok := s.r.runningOrders[roID]; !ok {
		return fmt.Errorf("running order %s: %w", roID, repository.ErrNotFound)
	}
	delete(s.r.runningOrders, roID)
	for k := range s.r.stories {
		if k.ro == roID {
			s.r.deleteStoryLocked(k)
		}
	}
	return nil
}

func (s runningOrders) List(_ context.Context) ([]*model.RunningOrder, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return sorted(s.r.runningOrders,
		func(*model.RunningOrder) bool { return true },
		cloneRunningOrder,
		func(a, b *model.RunningOrder) int { return newestFirst(a.CreatedAt, b.CreatedAt, a.ID, b.ID) },
	), nil
}
