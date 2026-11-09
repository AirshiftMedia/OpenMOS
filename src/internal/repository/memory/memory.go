// Package memory is an in-memory implementation of repository.Repository.
//
// It is the reference implementation of the contract in package repotest and
// a fake for service tests. Nothing is persisted.
package memory

import (
	"cmp"
	"maps"
	"slices"
	"sync"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

type storyKey struct{ ro, story string }

type itemKey struct{ ro, story, item string }

type objectKey struct{ mos, obj string }

// Repository holds all records behind one lock, so deletes can cascade
// across record types atomically.
type Repository struct {
	mu            sync.RWMutex
	runningOrders map[string]*model.RunningOrder
	stories       map[storyKey]*model.Story
	items         map[itemKey]*model.Item
	objects       map[objectKey]*model.MOSObject
}

// New returns an empty repository.
func New() *Repository {
	return &Repository{
		runningOrders: make(map[string]*model.RunningOrder),
		stories:       make(map[storyKey]*model.Story),
		items:         make(map[itemKey]*model.Item),
		objects:       make(map[objectKey]*model.MOSObject),
	}
}

var _ repository.Repository = (*Repository)(nil)

// RunningOrders returns the running order repository.
func (r *Repository) RunningOrders() repository.RunningOrderRepository { return runningOrders{r} }

// Stories returns the story repository.
func (r *Repository) Stories() repository.StoryRepository { return stories{r} }

// Items returns the item repository.
func (r *Repository) Items() repository.ItemRepository { return items{r} }

// Objects returns the MOS object repository.
func (r *Repository) Objects() repository.ObjectRepository { return objects{r} }

// now returns the current time without the monotonic reading, so stored
// timestamps compare like ones read back from a database.
func now() time.Time { return time.Now().Round(0) }

// deleteStoryLocked deletes a story and its items. The caller holds r.mu.
func (r *Repository) deleteStoryLocked(k storyKey) {
	delete(r.stories, k)
	for ik := range r.items {
		if ik.ro == k.ro && ik.story == k.story {
			delete(r.items, ik)
		}
	}
}

// Stored values are copied on the way in and out, so callers never share
// memory with the store.

func cloneRunningOrder(ro *model.RunningOrder) *model.RunningOrder {
	c := *ro
	c.Metadata = maps.Clone(ro.Metadata)
	if ro.AirTime != nil {
		t := *ro.AirTime
		c.AirTime = &t
	}
	return &c
}

func cloneStory(s *model.Story) *model.Story {
	c := *s
	c.Metadata = maps.Clone(s.Metadata)
	return &c
}

func cloneItem(i *model.Item) *model.Item {
	c := *i
	c.Metadata = maps.Clone(i.Metadata)
	return &c
}

func cloneObject(o *model.MOSObject) *model.MOSObject {
	c := *o
	c.Metadata = maps.Clone(o.Metadata)
	return &c
}

// newestFirst orders by CreatedAt descending, then by key for equal times.
func newestFirst(aCreated, bCreated time.Time, aKey, bKey string) int {
	if c := bCreated.Compare(aCreated); c != 0 {
		return c
	}
	return cmp.Compare(aKey, bKey)
}

// sorted returns copies of the values in m that keep accepts, sorted by
// cmpFn. The caller holds r.mu.
func sorted[K comparable, T any](m map[K]*T, keep func(*T) bool, clone func(*T) *T, cmpFn func(a, b *T) int) []*T {
	out := make([]*T, 0)
	for _, v := range m {
		if keep(v) {
			out = append(out, clone(v))
		}
	}
	slices.SortFunc(out, cmpFn)
	return out
}

// byOrder sorts children by Order, then by ID.
func byOrder(aOrder, bOrder int, aID, bID string) int {
	return cmp.Or(cmp.Compare(aOrder, bOrder), cmp.Compare(aID, bID))
}

// nextOrder returns the Order for a child appended after siblings.
func nextOrder(siblings []int) int {
	if len(siblings) == 0 {
		return 0
	}
	return slices.Max(siblings) + 1
}
