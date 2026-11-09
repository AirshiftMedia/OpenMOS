package repotest

import (
	"sync"
	"testing"

	"airshift/openmos/internal/model"
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

	// Writers that read the same version race; exactly one may win, the
	// rest get ErrVersionConflict
	t.Run("ConcurrentUpdates", func(t *testing.T) {
		r, ctx := newRepo(t).RunningOrders(), t.Context()
		_, err := r.Create(ctx, newRunningOrder("a"))
		must(t, "Create", err)
		const n = 10
		copies := make([]*model.RunningOrder, n)
		for i := range copies {
			copies[i], err = r.Get(ctx, "a")
			must(t, "Get", err)
			copies[i].Slug = "writer " + itoa(i)
		}
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := range copies {
			wg.Go(func() { errs[i] = r.Update(ctx, copies[i]) })
		}
		wg.Wait()

		winner := -1
		for i, err := range errs {
			switch {
			case err == nil && winner < 0:
				winner = i
			case err == nil:
				t.Errorf("writers %d and %d both succeeded", winner, i)
			default:
				wantErr(t, "losing writer", err, repository.ErrVersionConflict)
			}
		}
		if winner < 0 {
			t.Fatal("no writer succeeded")
		}
		got, err := r.Get(ctx, "a")
		must(t, "Get", err)
		if got.Version != 2 || got.Slug != copies[winner].Slug {
			t.Errorf("after race: version %d, slug %q; want 2, %q", got.Version, got.Slug, copies[winner].Slug)
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
