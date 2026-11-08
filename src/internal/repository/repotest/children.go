package repotest

import (
	"testing"

	"airshift/openmos/internal/repository"
)

func testStories(t *testing.T, newRepo Factory) {
	t.Run("CreateRequiresRunningOrder", func(t *testing.T) {
		r := newRepo(t)
		_, err := r.Stories().Create(t.Context(), newStory("s1", "missing", 1))
		wantErr(t, "Create story without running order", err, repository.ErrNotFound)
	})

	t.Run("ListByRunningOrder", func(t *testing.T) {
		r := newRepo(t)
		createRunningOrder(t, r, "ro-1")
		createRunningOrder(t, r, "ro-2")
		// Created out of order, so the list must follow Order, not insertion
		createStory(t, r, "s3", "ro-1", 3)
		createStory(t, r, "s1", "ro-1", 1)
		createStory(t, r, "s2", "ro-1", 2)
		createStory(t, r, "other", "ro-2", 1)

		list := func(roID string) []string {
			t.Helper()
			stories, err := r.Stories().ListByRunningOrder(t.Context(), roID)
			must(t, "ListByRunningOrder "+roID, err)
			return storyKind.ids(stories)
		}
		equalIDs(t, "ro-1", list("ro-1"), []string{"s1", "s2", "s3"})
		equalIDs(t, "ro-2", list("ro-2"), []string{"other"})
		equalIDs(t, "unknown running order", list("missing"), nil)

		// Moving a story first must reorder the list
		s3, err := r.Stories().Get(t.Context(), "s3")
		must(t, "Get s3", err)
		s3.Order = 0
		must(t, "Update s3", r.Stories().Update(t.Context(), s3))
		equalIDs(t, "ro-1 after move", list("ro-1"), []string{"s3", "s1", "s2"})
	})
}

func testItems(t *testing.T, newRepo Factory) {
	t.Run("CreateRequiresStory", func(t *testing.T) {
		r := newRepo(t)
		_, err := r.Items().Create(t.Context(), newItem("i1", "missing", 1))
		wantErr(t, "Create item without story", err, repository.ErrNotFound)
	})

	t.Run("ListByStory", func(t *testing.T) {
		r := newRepo(t)
		createRunningOrder(t, r, "ro-1")
		createStory(t, r, "s1", "ro-1", 1)
		createStory(t, r, "s2", "ro-1", 2)
		createItem(t, r, "i3", "s1", 3)
		createItem(t, r, "i1", "s1", 1)
		createItem(t, r, "i2", "s1", 2)
		createItem(t, r, "other", "s2", 1)

		list := func(storyID string) []string {
			t.Helper()
			items, err := r.Items().ListByStory(t.Context(), storyID)
			must(t, "ListByStory "+storyID, err)
			return itemKind.ids(items)
		}
		equalIDs(t, "s1", list("s1"), []string{"i1", "i2", "i3"})
		equalIDs(t, "s2", list("s2"), []string{"other"})
		equalIDs(t, "unknown story", list("missing"), nil)

		i3, err := r.Items().Get(t.Context(), "i3")
		must(t, "Get i3", err)
		i3.Order = 0
		must(t, "Update i3", r.Items().Update(t.Context(), i3))
		equalIDs(t, "s1 after move", list("s1"), []string{"i3", "i1", "i2"})
	})
}
