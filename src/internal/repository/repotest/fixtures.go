package repotest

import (
	"testing"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

// IDs of the parents created for the story and item CRUD tests.
const (
	parentRO    = "ro-parent"
	parentStory = "story-parent"
)

// The fixtures fill every field, so a store that drops one fails the
// round-trip comparison.

func newRunningOrder(id string) *model.RunningOrder {
	air := time.Date(2026, 11, 8, 18, 30, 0, 0, time.UTC)
	return &model.RunningOrder{
		ID: id, MosID: "test.mos", Slug: "slug " + id, Status: model.StatusReady,
		Duration: 1800, FirstStoryID: "s-first", LastStoryID: "s-last",
		AirTime: &air, Channel: "A", Metadata: map[string]string{"k": "v"},
		Version: 1, CreatedBy: "repotest",
	}
}

func newStory(id, roID string, order int) *model.Story {
	return &model.Story{
		ID: id, RunningOrderID: roID, Slug: "slug " + id, Number: "A1",
		Duration: 90, Status: model.StatusPending, Order: order,
		PreviousID: "s-prev", NextID: "s-next", Presenter: "presenter",
		Metadata: map[string]string{"k": "v"},
	}
}

func newItem(id, storyID string, order int) *model.Item {
	return &model.Item{
		ID: id, ObjectID: "obj-1", Slug: "slug " + id, Duration: 30,
		EditorialDuration: 25, TimeBase: 25, Status: model.StatusReady,
		Order: order, StoryID: storyID, Metadata: map[string]string{"k": "v"},
	}
}

func newObject(id string) *model.MOSObject {
	return &model.MOSObject{
		ID: id, ObjectType: "VIDEO", Slug: "slug " + id, Duration: 30,
		TimeBase: 25, Status: model.StatusReady, ObjectID: "ext-" + id,
		MediaID: "media-" + id, MosAbstract: "abstract",
		Metadata: map[string]string{"k": "v"},
	}
}

func createRunningOrder(t *testing.T, r repository.Repository, id string) {
	t.Helper()
	_, err := r.RunningOrders().Create(t.Context(), newRunningOrder(id))
	must(t, "create running order "+id, err)
}

func createStory(t *testing.T, r repository.Repository, id, roID string, order int) {
	t.Helper()
	_, err := r.Stories().Create(t.Context(), newStory(id, roID, order))
	must(t, "create story "+id, err)
}

func createItem(t *testing.T, r repository.Repository, id, storyID string, order int) {
	t.Helper()
	_, err := r.Items().Create(t.Context(), newItem(id, storyID, order))
	must(t, "create item "+id, err)
}

func createObject(t *testing.T, r repository.Repository, id string) {
	t.Helper()
	_, err := r.Objects().Create(t.Context(), newObject(id))
	must(t, "create object "+id, err)
}
