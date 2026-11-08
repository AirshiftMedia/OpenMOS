package memory

import (
	"context"
	"errors"
	"fmt"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type objects struct{ r *Repository }

func (s objects) Create(_ context.Context, obj *model.MOSObject) (*model.MOSObject, error) {
	if obj.ID == "" {
		return nil, errors.New("object ID is required")
	}
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.objects[obj.ID]; ok {
		return nil, fmt.Errorf("object %s: %w", obj.ID, repository.ErrAlreadyExists)
	}
	t := now()
	obj.CreatedAt, obj.UpdatedAt = t, t
	s.r.objects[obj.ID] = cloneObject(obj)
	return cloneObject(obj), nil
}

func (s objects) Get(_ context.Context, id string) (*model.MOSObject, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	obj, ok := s.r.objects[id]
	if !ok {
		return nil, fmt.Errorf("object %s: %w", id, repository.ErrNotFound)
	}
	return cloneObject(obj), nil
}

func (s objects) Update(_ context.Context, obj *model.MOSObject) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	current, ok := s.r.objects[obj.ID]
	if !ok {
		return fmt.Errorf("object %s: %w", obj.ID, repository.ErrNotFound)
	}
	obj.CreatedAt = current.CreatedAt
	obj.UpdatedAt = now()
	s.r.objects[obj.ID] = cloneObject(obj)
	return nil
}

func (s objects) Delete(_ context.Context, id string) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	if _, ok := s.r.objects[id]; !ok {
		return fmt.Errorf("object %s: %w", id, repository.ErrNotFound)
	}
	delete(s.r.objects, id)
	return nil
}

func (s objects) List(_ context.Context) ([]*model.MOSObject, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return sorted(s.r.objects,
		func(*model.MOSObject) bool { return true },
		cloneObject,
		func(a, b *model.MOSObject) int { return newestFirst(a.CreatedAt, b.CreatedAt, a.ID, b.ID) },
	), nil
}
