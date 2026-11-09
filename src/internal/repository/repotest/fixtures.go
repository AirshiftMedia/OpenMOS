package repotest

import (
	"testing"
	"time"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

const (
	// Parents created for the story and item CRUD tests
	parentRO    = "ro-parent"
	parentStory = "story-parent"

	// MOS device of the object fixtures
	testMOS = "test.mos"
)

// The fixtures fill every field, so a store that drops one fails the
// round-trip comparison.

func newRunningOrder(id string) *model.RunningOrder {
	air := time.Date(2026, 11, 8, 18, 30, 0, 0, time.UTC)
	return &model.RunningOrder{
		ID: id, MosID: testMOS, Slug: "slug " + id, Status: model.StatusReady,
		Duration: 1800, AirTime: &air, Channel: "A",
		Metadata: map[string]string{"k": "v"}, Version: 1, CreatedBy: "repotest",
	}
}

func newStory(roID, id string) *model.Story {
	return &model.Story{
		RunningOrderID: roID, ID: id, Slug: "slug " + id, Number: "A1",
		Status: model.StatusPending, Presenter: "presenter",
		Metadata: map[string]string{"k": "v"},
	}
}

func newItem(roID, storyID, id string) *model.Item {
	return &model.Item{
		RunningOrderID: roID, StoryID: storyID, ID: id, ObjectID: "obj-1",
		MosID: testMOS, Slug: "slug " + id, Duration: 1500, TimeBase: 59.94,
		Status: model.StatusReady, Metadata: map[string]string{"k": "v"},
	}
}

func newObject(mosID, id string) *model.MOSObject {
	return &model.MOSObject{
		MosID: mosID, ID: id, ObjectType: "VIDEO", Slug: "slug " + id,
		Duration: 4294967295, TimeBase: 50, Status: model.StatusReady,
		Abstract: "abstract", Metadata: map[string]string{"k": "v"},
	}
}

func createRunningOrder(t *testing.T, r repository.Repository, id string) {
	t.Helper()
	_, err := r.RunningOrders().Create(t.Context(), newRunningOrder(id))
	must(t, "create running order "+id, err)
}

func createStory(t *testing.T, r repository.Repository, roID, id string) {
	t.Helper()
	_, err := r.Stories().Create(t.Context(), newStory(roID, id))
	must(t, "create story "+roID+"/"+id, err)
}

func createItem(t *testing.T, r repository.Repository, roID, storyID, id string) {
	t.Helper()
	_, err := r.Items().Create(t.Context(), newItem(roID, storyID, id))
	must(t, "create item "+roID+"/"+storyID+"/"+id, err)
}

func createObject(t *testing.T, r repository.Repository, mosID, id string) {
	t.Helper()
	_, err := r.Objects().Create(t.Context(), newObject(mosID, id))
	must(t, "create object "+mosID+"/"+id, err)
}
