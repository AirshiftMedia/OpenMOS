package repotest

import (
	"testing"

	"airshift/openmos/internal/repository"
)

func testStories(t *testing.T, newRepo Factory) {
	t.Run("CreateRequiresRunningOrder", func(t *testing.T) {
		r := newRepo(t)
		_, err := r.Stories().Create(t.Context(), newStory("missing", "s1"))
		wantErr(t, "Create story without running order", err, repository.ErrNotFound)
	})

	// A storyID is unique only within its running order
	t.Run("KeyIncludesRunningOrder", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createRunningOrder(t, r, "ro-1")
		createRunningOrder(t, r, "ro-2")
		createStory(t, r, "ro-1", "s")
		createStory(t, r, "ro-2", "s")

		s, err := r.Stories().Get(ctx, "ro-1", "s")
		must(t, "Get ro-1/s", err)
		s.Slug = "changed in ro-1"
		must(t, "Update ro-1/s", r.Stories().Update(ctx, s))
		other, err := r.Stories().Get(ctx, "ro-2", "s")
		must(t, "Get ro-2/s", err)
		if other.Slug != "slug s" {
			t.Errorf("updating ro-1/s changed ro-2/s: slug %q", other.Slug)
		}

		must(t, "Delete ro-1/s", r.Stories().Delete(ctx, "ro-1", "s"))
		_, err = r.Stories().Get(ctx, "ro-2", "s")
		must(t, "Get ro-2/s after deleting ro-1/s", err)
		_, err = r.Stories().Get(ctx, "ro-3", "s")
		wantErr(t, "Get s in an unknown running order", err, repository.ErrNotFound)
	})
}

func testItems(t *testing.T, newRepo Factory) {
	t.Run("CreateRequiresStory", func(t *testing.T) {
		r := newRepo(t)
		createRunningOrder(t, r, "ro-1")
		_, err := r.Items().Create(t.Context(), newItem("ro-1", "missing", "i1"))
		wantErr(t, "Create item without story", err, repository.ErrNotFound)
	})

	// An itemID is unique only within its story, and a storyID only within
	// its running order
	t.Run("KeyIncludesStoryAndRunningOrder", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createRunningOrder(t, r, "ro-1")
		createRunningOrder(t, r, "ro-2")
		createStory(t, r, "ro-1", "s1")
		createStory(t, r, "ro-1", "s2")
		createStory(t, r, "ro-2", "s1")
		for _, k := range [][2]string{{"ro-1", "s1"}, {"ro-1", "s2"}, {"ro-2", "s1"}} {
			createItem(t, r, k[0], k[1], "i")
		}

		must(t, "Delete ro-1/s1/i", r.Items().Delete(ctx, "ro-1", "s1", "i"))
		for _, k := range [][2]string{{"ro-1", "s2"}, {"ro-2", "s1"}} {
			_, err := r.Items().Get(ctx, k[0], k[1], "i")
			must(t, "Get "+k[0]+"/"+k[1]+"/i after deleting ro-1/s1/i", err)
		}
	})
}
