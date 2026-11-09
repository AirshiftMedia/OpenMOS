package repotest

import (
	"errors"
	"strings"
	"testing"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

func testCascade(t *testing.T, newRepo Factory) {
	// setup creates two running orders with stories and items, and an
	// object. Story s1 exists in both running orders.
	//
	//	ro-1: s1 (i1, i2), s2 (i3)
	//	ro-2: s1 (i1)
	//	test.mos/obj-1
	setup := func(t *testing.T) repository.Repository {
		r := newRepo(t)
		createRunningOrder(t, r, "ro-1")
		createRunningOrder(t, r, "ro-2")
		createStory(t, r, "ro-1", "s1")
		createStory(t, r, "ro-1", "s2")
		createStory(t, r, "ro-2", "s1")
		createItem(t, r, "ro-1", "s1", "i1")
		createItem(t, r, "ro-1", "s1", "i2")
		createItem(t, r, "ro-1", "s2", "i3")
		createItem(t, r, "ro-2", "s1", "i1")
		createObject(t, r, testMOS, "obj-1")
		return r
	}
	all := []string{"ro-1", "ro-2", "ro-1/s1", "ro-1/s2", "ro-2/s1",
		"ro-1/s1/i1", "ro-1/s1/i2", "ro-1/s2/i3", "ro-2/s1/i1"}

	t.Run("DeleteRunningOrder", func(t *testing.T) {
		r := setup(t)
		must(t, "Delete ro-1", r.RunningOrders().Delete(t.Context(), "ro-1"))
		exist(t, r, all, "ro-2", "ro-2/s1", "ro-2/s1/i1")
	})

	t.Run("DeleteStory", func(t *testing.T) {
		r := setup(t)
		must(t, "Delete ro-1/s1", r.Stories().Delete(t.Context(), "ro-1", "s1"))
		exist(t, r, all, "ro-1", "ro-2", "ro-1/s2", "ro-2/s1", "ro-1/s2/i3", "ro-2/s1/i1")
	})

	t.Run("ReplaceAllStories", func(t *testing.T) {
		r := setup(t)
		// s1 comes back, but without its items
		must(t, "ReplaceAll stories of ro-1", r.Stories().ReplaceAll(t.Context(), "ro-1",
			[]*model.Story{newStory("ro-1", "s1"), newStory("ro-1", "s9")}))
		exist(t, r, append(all, "ro-1/s9"), "ro-1", "ro-2", "ro-1/s1", "ro-1/s9", "ro-2/s1", "ro-2/s1/i1")
	})

	t.Run("ObjectsAreIndependent", func(t *testing.T) {
		r := setup(t)
		must(t, "Delete ro-1", r.RunningOrders().Delete(t.Context(), "ro-1"))
		must(t, "Delete ro-2", r.RunningOrders().Delete(t.Context(), "ro-2"))
		_, err := r.Objects().Get(t.Context(), testMOS, "obj-1")
		must(t, "Get object after deleting all running orders", err)
	})
}

// exist fails the test unless, of the records named by paths ("ro",
// "ro/story" or "ro/story/item"), exactly those in want exist.
func exist(t *testing.T, r repository.Repository, paths []string, want ...string) {
	t.Helper()
	ctx := t.Context()
	for _, p := range paths {
		var err error
		switch k := strings.Split(p, "/"); len(k) {
		case 1:
			_, err = r.RunningOrders().Get(ctx, k[0])
		case 2:
			_, err = r.Stories().Get(ctx, k[0], k[1])
		default:
			_, err = r.Items().Get(ctx, k[0], k[1], k[2])
		}
		kept := false
		for _, w := range want {
			kept = kept || w == p
		}
		switch {
		case kept && err != nil:
			t.Errorf("Get %s: %v, want it kept", p, err)
		case !kept && !errors.Is(err, repository.ErrNotFound):
			t.Errorf("Get %s: err = %v, want %v", p, err, repository.ErrNotFound)
		}
	}
}
