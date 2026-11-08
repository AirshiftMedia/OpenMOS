package repotest

import (
	"context"
	"reflect"
	"testing"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

// crud is the part every record repository shares.
type crud[T any] interface {
	Create(ctx context.Context, v *T) (*T, error)
	Get(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, v *T) error
	Delete(ctx context.Context, id string) error
}

// common points at the fields every record type has.
type common struct {
	ID, Slug             *string
	Metadata             *map[string]string
	CreatedAt, UpdatedAt *time.Time
}

// kind describes one record type to the shared CRUD tests in crud.go.
type kind[T any] struct {
	repo    func(repository.Repository) crud[T]
	parents func(t *testing.T, repo repository.Repository) // creates what fixture refers to
	fixture func(id string) *T
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
	repo:    func(r repository.Repository) crud[model.RunningOrder] { return r.RunningOrders() },
	parents: func(*testing.T, repository.Repository) {},
	fixture: newRunningOrder,
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
	repo: func(r repository.Repository) crud[model.Story] { return r.Stories() },
	parents: func(t *testing.T, r repository.Repository) {
		createRunningOrder(t, r, parentRO)
	},
	fixture: func(id string) *model.Story { return newStory(id, parentRO, 1) },
	fields: func(v *model.Story) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
}

var itemKind = kind[model.Item]{
	repo: func(r repository.Repository) crud[model.Item] { return r.Items() },
	parents: func(t *testing.T, r repository.Repository) {
		createRunningOrder(t, r, parentRO)
		createStory(t, r, parentStory, parentRO, 1)
	},
	fixture: func(id string) *model.Item { return newItem(id, parentStory, 1) },
	fields: func(v *model.Item) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
}

var objectKind = kind[model.MOSObject]{
	repo:    func(r repository.Repository) crud[model.MOSObject] { return r.Objects() },
	parents: func(*testing.T, repository.Repository) {},
	fixture: newObject,
	fields: func(v *model.MOSObject) common {
		return common{&v.ID, &v.Slug, &v.Metadata, &v.CreatedAt, &v.UpdatedAt}
	},
}
