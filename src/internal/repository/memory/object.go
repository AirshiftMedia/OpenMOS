package memory

import (
	"context"
	"fmt"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type objects struct{ r *Repository }

func (s objects) Create(_ context.Context, obj *model.MOSObject) (*model.MOSObject, error) {
	if err := repository.CheckID("MOS", obj.MosID); err != nil {
		return nil, err
	}
	if err := repository.CheckID("object", obj.ID); err != nil {
		return nil, err
	}
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	k := objectKey{obj.MosID, obj.ID}
	if _, ok := s.r.objects[k]; ok {
		return nil, fmt.Errorf("object %s/%s: %w", k.mos, k.obj, repository.ErrAlreadyExists)
	}
	t := now()
	obj.CreatedAt, obj.UpdatedAt = t, t
	s.r.objects[k] = cloneObject(obj)
	return cloneObject(obj), nil
}

func (s objects) Get(_ context.Context, mosID, objID string) (*model.MOSObject, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	obj, ok := s.r.objects[objectKey{mosID, objID}]
	if !ok {
		return nil, fmt.Errorf("object %s/%s: %w", mosID, objID, repository.ErrNotFound)
	}
	return cloneObject(obj), nil
}

func (s objects) Update(_ context.Context, obj *model.MOSObject) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	k := objectKey{obj.MosID, obj.ID}
	current, ok := s.r.objects[k]
	if !ok {
		return fmt.Errorf("object %s/%s: %w", k.mos, k.obj, repository.ErrNotFound)
	}
	obj.CreatedAt = current.CreatedAt
	obj.UpdatedAt = now()
	s.r.objects[k] = cloneObject(obj)
	return nil
}

func (s objects) Delete(_ context.Context, mosID, objID string) error {
	s.r.mu.Lock()
	defer s.r.mu.Unlock()
	k := objectKey{mosID, objID}
	if _, ok := s.r.objects[k]; !ok {
		return fmt.Errorf("object %s/%s: %w", mosID, objID, repository.ErrNotFound)
	}
	delete(s.r.objects, k)
	return nil
}

func (s objects) List(_ context.Context) ([]*model.MOSObject, error) {
	s.r.mu.RLock()
	defer s.r.mu.RUnlock()
	return sorted(s.r.objects,
		func(*model.MOSObject) bool { return true },
		cloneObject,
		func(a, b *model.MOSObject) int {
			return newestFirst(a.CreatedAt, b.CreatedAt, a.MosID+"/"+a.ID, b.MosID+"/"+b.ID)
		},
	), nil
}
