package repotest

import (
	"testing"

	"airshift/openmos/internal/repository"
)

func testRunningOrders(t *testing.T, newRepo Factory) {
	t.Run("CreateSetsVersion", func(t *testing.T) {
		r := newRepo(t).RunningOrders()
		in := newRunningOrder("a")
		in.Version = 7
		created, err := r.Create(t.Context(), in)
		must(t, "Create", err)
		got, err := r.Get(t.Context(), "a")
		must(t, "Get", err)
		if created.Version != 1 || got.Version != 1 {
			t.Errorf("version after Create = %d (stored %d), want 1", created.Version, got.Version)
		}
	})

	t.Run("UpdateIncrementsVersion", func(t *testing.T) {
		r := newRepo(t).RunningOrders()
		_, err := r.Create(t.Context(), newRunningOrder("a"))
		must(t, "Create", err)
		ro, err := r.Get(t.Context(), "a")
		must(t, "Get", err)
		for want := 2; want <= 3; want++ {
			must(t, "Update", r.Update(t.Context(), ro))
			if ro.Version != want {
				t.Errorf("Update set passed Version to %d, want %d", ro.Version, want)
			}
		}
		got, err := r.Get(t.Context(), "a")
		must(t, "Get", err)
		if got.Version != 3 {
			t.Errorf("stored Version = %d, want 3", got.Version)
		}
	})

	t.Run("UpdateStaleVersion", func(t *testing.T) {
		r := newRepo(t).RunningOrders()
		_, err := r.Create(t.Context(), newRunningOrder("a"))
		must(t, "Create", err)
		first, err := r.Get(t.Context(), "a")
		must(t, "Get", err)
		second, err := r.Get(t.Context(), "a")
		must(t, "Get", err)

		first.Slug = "first writer"
		must(t, "first Update", r.Update(t.Context(), first))
		second.Slug = "second writer"
		wantErr(t, "Update with stale version", r.Update(t.Context(), second), repository.ErrVersionConflict)

		got, err := r.Get(t.Context(), "a")
		must(t, "Get", err)
		if got.Slug != "first writer" || got.Version != 2 {
			t.Errorf("after conflict: Slug %q, Version %d; want %q, 2", got.Slug, got.Version, "first writer")
		}
	})

	t.Run("ListNewestFirst", func(t *testing.T) {
		r := newRepo(t).RunningOrders()
		list, err := r.List(t.Context())
		must(t, "List empty", err)
		equalIDs(t, "List empty", runningOrderKind.ids(list), nil)

		for _, id := range []string{"a", "b", "c"} {
			_, err := r.Create(t.Context(), newRunningOrder(id))
			must(t, "Create "+id, err)
			tick()
		}
		list, err = r.List(t.Context())
		must(t, "List", err)
		equalIDs(t, "List", runningOrderKind.ids(list), []string{"c", "b", "a"})
	})
}
