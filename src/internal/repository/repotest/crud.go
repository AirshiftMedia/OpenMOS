package repotest

import (
	"context"
	"testing"
	"time"

	"airshift/openmos/internal/repository"
)

// testCRUD runs the tests every record type must pass.
func testCRUD[T any](t *testing.T, newRepo Factory, k kind[T]) {
	// setup returns a fresh repository with the fixture's parents in place.
	setup := func(t *testing.T) (context.Context, crud[T]) {
		repo := newRepo(t)
		k.parents(t, repo)
		return t.Context(), k.repo(repo)
	}

	t.Run("CreateAndGet", func(t *testing.T) {
		ctx, r := setup(t)
		before := time.Now()
		created, err := r.Create(ctx, k.fixture("a"))
		after := time.Now()
		must(t, "Create", err)
		c := k.fields(created)
		if !within(*c.CreatedAt, before, after) || !within(*c.UpdatedAt, before, after) {
			t.Errorf("Create set CreatedAt %v, UpdatedAt %v; want both between %v and %v",
				*c.CreatedAt, *c.UpdatedAt, before, after)
		}
		k.same(t, "Create result", created, k.fixture("a"))

		got, err := r.Get(ctx, "a")
		must(t, "Get", err)
		k.same(t, "Get after Create", got, k.fixture("a"))
		if g := k.fields(got); !closeTo(*g.CreatedAt, *c.CreatedAt) || !closeTo(*g.UpdatedAt, *c.UpdatedAt) {
			t.Errorf("Get timestamps %v / %v, want %v / %v", *g.CreatedAt, *g.UpdatedAt, *c.CreatedAt, *c.UpdatedAt)
		}
	})

	t.Run("CreateRequiresID", func(t *testing.T) {
		ctx, r := setup(t)
		if _, err := r.Create(ctx, k.fixture("")); err == nil {
			t.Error("Create with empty ID succeeded, want error")
		}
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		ctx, r := setup(t)
		_, err := r.Create(ctx, k.fixture("a"))
		must(t, "Create", err)
		dup := k.fixture("a")
		*k.fields(dup).Slug = "duplicate"
		_, err = r.Create(ctx, dup)
		wantErr(t, "Create duplicate", err, repository.ErrAlreadyExists)

		got, err := r.Get(ctx, "a")
		must(t, "Get", err)
		k.same(t, "Get after duplicate Create", got, k.fixture("a"))
	})

	t.Run("GetMissing", func(t *testing.T) {
		ctx, r := setup(t)
		_, err := r.Get(ctx, "missing")
		wantErr(t, "Get missing", err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		ctx, r := setup(t)
		created, err := r.Create(ctx, k.fixture("a"))
		must(t, "Create", err)
		c := k.fields(created)

		got, err := r.Get(ctx, "a")
		must(t, "Get", err)
		*k.fields(got).Slug = "updated"
		*k.fields(got).Metadata = map[string]string{"k": "v2", "k2": "v"}
		must(t, "Update", r.Update(ctx, got))

		updated, err := r.Get(ctx, "a")
		must(t, "Get after Update", err)
		k.same(t, "Get after Update", updated, got)
		u := k.fields(updated)
		if !closeTo(*u.CreatedAt, *c.CreatedAt) {
			t.Errorf("Update changed CreatedAt from %v to %v", *c.CreatedAt, *u.CreatedAt)
		}
		if u.UpdatedAt.Before(c.UpdatedAt.Add(-time.Millisecond)) {
			t.Errorf("Update moved UpdatedAt back from %v to %v", *c.UpdatedAt, *u.UpdatedAt)
		}
	})

	t.Run("UpdateMissing", func(t *testing.T) {
		ctx, r := setup(t)
		err := r.Update(ctx, k.fixture("missing"))
		wantErr(t, "Update missing", err, repository.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		ctx, r := setup(t)
		for _, id := range []string{"a", "b"} {
			_, err := r.Create(ctx, k.fixture(id))
			must(t, "Create "+id, err)
		}
		must(t, "Delete", r.Delete(ctx, "a"))

		_, err := r.Get(ctx, "a")
		wantErr(t, "Get deleted", err, repository.ErrNotFound)
		_, err = r.Get(ctx, "b")
		must(t, "Get other record after Delete", err)
		wantErr(t, "Delete again", r.Delete(ctx, "a"), repository.ErrNotFound)
	})

	t.Run("ReturnsCopies", func(t *testing.T) {
		ctx, r := setup(t)
		in := k.fixture("a")
		_, err := r.Create(ctx, in)
		must(t, "Create", err)
		// Changing the value passed to Create must not change the stored one
		*k.fields(in).Slug = "changed"
		(*k.fields(in).Metadata)["k"] = "changed"

		got, err := r.Get(ctx, "a")
		must(t, "Get", err)
		k.same(t, "Get after changing Create input", got, k.fixture("a"))

		// Neither may changing a value returned by Get
		(*k.fields(got).Metadata)["k"] = "changed"
		again, err := r.Get(ctx, "a")
		must(t, "Get again", err)
		k.same(t, "Get after changing Get result", again, k.fixture("a"))
	})
}
