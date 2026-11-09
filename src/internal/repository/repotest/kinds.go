package repotest

import (
	"context"
	"reflect"
	"testing"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

// common points at the fields every record type has.
type common struct {
	ID, Slug             *string
	Metadata             *map[string]string
	CreatedAt, UpdatedAt *time.Time
}

// kind describes one record type to the shared CRUD tests in crud.go. Get
// and delete take the key from the value passed, since keys are composite.
type kind[T any] struct {
	create  func(ctx context.Context, r repository.Repository, v *T) (*T, error)
	get     func(ctx context.Context, r repository.Repository, key *T) (*T, error)
	update  func(ctx context.Context, r repository.Repository, v *T) error
	del     func(ctx context.Context, r repository.Repository, key *T) error
	parents func(t *testing.T, r repository.Repository) // creates what fixture refers to
	fixture func(id string) *T
	minimal func(id string) *T // only the key set; every other field empty
	fields  func(v *T) common

	// normalize, if set, adjusts type-specific fields before comparison.
	normalize func(v *T)
}

// norm returns v without the fields a store sets itself, and with an empty
// map and nil map treated alike, so records can be compared with DeepEqual.
func (k kind[T]) norm(v T) T {
	f := k.fields(&v)
	*f.CreatedAt, *f.UpdatedAt = time.Time{}, time.Time{}
	if len(*f.Metadata) == 0 {
		*f.Metadata = nil
	}
	if k.normalize != nil {
		k.normalize(&v)
	}
	return v
}

// same fails the test unless got and want hold the same data.
func (k kind[T]) same(t *testing.T, op string, got, want *T) {
	t.Helper()
	if g, w := k.norm(*got), k.norm(*want); !reflect.DeepEqual(g, w) {
		t.Errorf("%s:\n got  %+v\n want %+v", op, g, w)
	}
}

// ids returns the IDs of list, in order.
func (k kind[T]) ids(list []*T) []string {
	out := make([]string, len(list))
	for i, v := range list {
		out[i] = *k.fields(v).ID
	}
	return out
}

var runningOrderKind = kind[model.RunningOrder]{
	create: func(ctx context.Context, r repository.Repository, v *model.RunningOrder) (*model.RunningOrder, error) {
		return r.RunningOrders().Create(ctx, v)
	},
	get: func(ctx context.Context, r repository.Repository, k *model.RunningOrder) (*model.RunningOrder, error) {
		return r.RunningOrders().Get(ctx, k.ID)
	},
	update: func(ctx context.Context, r repository.Repository, v *model.RunningOrder) error {
		return r.RunningOrders().Update(ctx, v)
	},
	del: func(ctx context.Context, r repository.Repository, k *model.RunningOrder) error {
		return r.RunningOrders().Delete(ctx, k.ID)
	},
	parents: func(*testing.T, repository.Repository) {},
	fixture: newRunningOrder,
	minimal: func(id string) *model.RunningOrder { return &model.RunningOrder{ID: id, Version: 1} },
	fields: func(v *model.RunningOrder) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
	// A store may return AirTime in another time zone; compare the instant.
	normalize: func(v *model.RunningOrder) {
		if v.AirTime != nil {
			air := v.AirTime.UTC()
			v.AirTime = &air
		}
	},
}

var storyKind = kind[model.Story]{
	create: func(ctx context.Context, r repository.Repository, v *model.Story) (*model.Story, error) {
		return r.Stories().Create(ctx, v)
	},
	get: func(ctx context.Context, r repository.Repository, k *model.Story) (*model.Story, error) {
		return r.Stories().Get(ctx, k.RunningOrderID, k.ID)
	},
	update: func(ctx context.Context, r repository.Repository, v *model.Story) error {
		return r.Stories().Update(ctx, v)
	},
	del: func(ctx context.Context, r repository.Repository, k *model.Story) error {
		return r.Stories().Delete(ctx, k.RunningOrderID, k.ID)
	},
	parents: func(t *testing.T, r repository.Repository) {
		createRunningOrder(t, r, parentRO)
	},
	fixture: func(id string) *model.Story { return newStory(parentRO, id) },
	minimal: func(id string) *model.Story { return &model.Story{RunningOrderID: parentRO, ID: id} },
	fields: func(v *model.Story) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
	// Order depends on creation order and is tested in ordering.go
	normalize: func(v *model.Story) { v.Order = 0 },
}

var itemKind = kind[model.Item]{
	create: func(ctx context.Context, r repository.Repository, v *model.Item) (*model.Item, error) {
		return r.Items().Create(ctx, v)
	},
	get: func(ctx context.Context, r repository.Repository, k *model.Item) (*model.Item, error) {
		return r.Items().Get(ctx, k.RunningOrderID, k.StoryID, k.ID)
	},
	update: func(ctx context.Context, r repository.Repository, v *model.Item) error {
		return r.Items().Update(ctx, v)
	},
	del: func(ctx context.Context, r repository.Repository, k *model.Item) error {
		return r.Items().Delete(ctx, k.RunningOrderID, k.StoryID, k.ID)
	},
	parents: func(t *testing.T, r repository.Repository) {
		createRunningOrder(t, r, parentRO)
		createStory(t, r, parentRO, parentStory)
	},
	fixture: func(id string) *model.Item { return newItem(parentRO, parentStory, id) },
	minimal: func(id string) *model.Item {
		return &model.Item{RunningOrderID: parentRO, StoryID: parentStory, ID: id}
	},
	fields: func(v *model.Item) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
	normalize: func(v *model.Item) { v.Order = 0 },
}

var objectKind = kind[model.MOSObject]{
	create: func(ctx context.Context, r repository.Repository, v *model.MOSObject) (*model.MOSObject, error) {
		return r.Objects().Create(ctx, v)
	},
	get: func(ctx context.Context, r repository.Repository, k *model.MOSObject) (*model.MOSObject, error) {
		return r.Objects().Get(ctx, k.MosID, k.ID)
	},
	update: func(ctx context.Context, r repository.Repository, v *model.MOSObject) error {
		return r.Objects().Update(ctx, v)
	},
	del: func(ctx context.Context, r repository.Repository, k *model.MOSObject) error {
		return r.Objects().Delete(ctx, k.MosID, k.ID)
	},
	parents: func(*testing.T, repository.Repository) {},
	fixture: func(id string) *model.MOSObject { return newObject(testMOS, id) },
	minimal: func(id string) *model.MOSObject { return &model.MOSObject{MosID: testMOS, ID: id} },
	fields: func(v *model.MOSObject) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
}
