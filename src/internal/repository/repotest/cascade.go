package repotest

import (
	"testing"

	"airshift/openmos/internal/repository"
)

func testCascade(t *testing.T, newRepo Factory) {
	// setup creates two running orders with stories and items, and an object:
	//
	//	ro-1: s1 (i1, i2), s2 (i3)
	//	ro-2: s3 (i4)
	//	obj-1
	setup := func(t *testing.T) repository.Repository {
		r := newRepo(t)
		createRunningOrder(t, r, "ro-1")
		createRunningOrder(t, r, "ro-2")
		createStory(t, r, "s1", "ro-1", 1)
		createStory(t, r, "s2", "ro-1", 2)
		createStory(t, r, "s3", "ro-2", 1)
		createItem(t, r, "i1", "s1", 1)
		createItem(t, r, "i2", "s1", 2)
		createItem(t, r, "i3", "s2", 1)
		createItem(t, r, "i4", "s3", 1)
		createObject(t, r, "obj-1")
		return r
	}

	t.Run("DeleteRunningOrder", func(t *testing.T) {
		r := setup(t)
		must(t, "Delete ro-1", r.RunningOrders().Delete(t.Context(), "ro-1"))

		gone(t, r, []string{"s1", "s2"}, []string{"i1", "i2", "i3"})
		kept(t, r, []string{"ro-2"}, []string{"s3"}, []string{"i4"}, []string{"obj-1"})
	})

	t.Run("DeleteStory", func(t *testing.T) {
		r := setup(t)
		must(t, "Delete s1", r.Stories().Delete(t.Context(), "s1"))

		gone(t, r, []string{"s1"}, []string{"i1", "i2"})
		kept(t, r, []string{"ro-1", "ro-2"}, []string{"s2", "s3"}, []string{"i3", "i4"}, []string{"obj-1"})
	})
}

// gone fails the test unless the stories and items no longer exist.
func gone(t *testing.T, r repository.Repository, stories, items []string) {
	t.Helper()
	for _, id := range stories {
		_, err := r.Stories().Get(t.Context(), id)
		wantErr(t, "Get story "+id, err, repository.ErrNotFound)
	}
	for _, id := range items {
		_, err := r.Items().Get(t.Context(), id)
		wantErr(t, "Get item "+id, err, repository.ErrNotFound)
	}
}

// kept fails the test unless all the given records still exist.
func kept(t *testing.T, r repository.Repository, ros, stories, items, objects []string) {
	t.Helper()
	ctx := t.Context()
	check := func(what, id string, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("Get %s %s: %v, want it kept", what, id, err)
		}
	}
	for _, id := range ros {
		_, err := r.RunningOrders().Get(ctx, id)
		check("running order", id, err)
	}
	for _, id := range stories {
		_, err := r.Stories().Get(ctx, id)
		check("story", id, err)
	}
	for _, id := range items {
		_, err := r.Items().Get(ctx, id)
		check("item", id, err)
	}
	for _, id := range objects {
		_, err := r.Objects().Get(ctx, id)
		check("object", id, err)
	}
}
