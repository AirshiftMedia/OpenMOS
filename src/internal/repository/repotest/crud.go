package repotest

import (
	"strings"
	"testing"
	"time"

	"airshift/openmos/internal/repository"
)

// testCRUD runs the tests every record type must pass.
func testCRUD[T any](t *testing.T, newRepo Factory, k kind[T]) {
	// setup returns a fresh repository with the fixture's parents in place.
	setup := func(t *testing.T) repository.Repository {
		r := newRepo(t)
		k.parents(t, r)
		return r
	}
	key := k.fixture // a fixture carries the full key of the record

	t.Run("CreateAndGet", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		before := time.Now()
		created, err := k.create(ctx, r, k.fixture("a"))
		after := time.Now()
		must(t, "Create", err)
		c := k.fields(created)
		if !within(*c.CreatedAt, before, after) || !within(*c.UpdatedAt, before, after) {
			t.Errorf("Create set CreatedAt %v, UpdatedAt %v; want both between %v and %v",
				*c.CreatedAt, *c.UpdatedAt, before, after)
		}
		k.same(t, "Create result", created, k.fixture("a"))

		got, err := k.get(ctx, r, key("a"))
		must(t, "Get", err)
		k.same(t, "Get after Create", got, k.fixture("a"))
		if g := k.fields(got); !closeTo(*g.CreatedAt, *c.CreatedAt) || !closeTo(*g.UpdatedAt, *c.UpdatedAt) {
			t.Errorf("Get timestamps %v / %v, want %v / %v", *g.CreatedAt, *g.UpdatedAt, *c.CreatedAt, *c.UpdatedAt)
		}
	})

	// Unknown values (zero durations and time bases, empty types, no
	// metadata) must round-trip as such
	t.Run("CreateMinimal", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		created, err := k.create(ctx, r, k.minimal("a"))
		must(t, "Create", err)
		k.same(t, "Create result", created, k.minimal("a"))
		got, err := k.get(ctx, r, key("a"))
		must(t, "Get", err)
		k.same(t, "Get after Create", got, k.minimal("a"))
	})

	t.Run("CreateChecksID", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		for _, id := range []string{"", strings.Repeat("x", repository.MaxIDLength+1)} {
			_, err := k.create(ctx, r, k.fixture(id))
			wantErr(t, "Create with ID of length "+itoa(len(id)), err, repository.ErrInvalid)
		}
		// The longest allowed ID works, counted in characters, not bytes
		longest := strings.Repeat("ä", repository.MaxIDLength)
		_, err := k.create(ctx, r, k.fixture(longest))
		must(t, "Create with a 128-character ID", err)
		_, err = k.get(ctx, r, key(longest))
		must(t, "Get with a 128-character ID", err)
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		_, err := k.create(ctx, r, k.fixture("a"))
		must(t, "Create", err)
		dup := k.fixture("a")
		*k.fields(dup).Slug = "duplicate"
		_, err = k.create(ctx, r, dup)
		wantErr(t, "Create duplicate", err, repository.ErrAlreadyExists)

		got, err := k.get(ctx, r, key("a"))
		must(t, "Get", err)
		k.same(t, "Get after duplicate Create", got, k.fixture("a"))
	})

	t.Run("GetMissing", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		_, err := k.get(ctx, r, key("missing"))
		wantErr(t, "Get missing", err, repository.ErrNotFound)
	})

	t.Run("Update", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		created, err := k.create(ctx, r, k.fixture("a"))
		must(t, "Create", err)
		c := k.fields(created)

		got, err := k.get(ctx, r, key("a"))
		must(t, "Get", err)
		*k.fields(got).Slug = "updated"
		*k.fields(got).Metadata = map[string]string{"k": "v2", "k2": "v"}
		must(t, "Update", k.update(ctx, r, got))

		updated, err := k.get(ctx, r, key("a"))
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
		r, ctx := setup(t), t.Context()
		wantErr(t, "Update missing", k.update(ctx, r, k.fixture("missing")), repository.ErrNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		for _, id := range []string{"a", "b"} {
			_, err := k.create(ctx, r, k.fixture(id))
			must(t, "Create "+id, err)
		}
		must(t, "Delete", k.del(ctx, r, key("a")))

		_, err := k.get(ctx, r, key("a"))
		wantErr(t, "Get deleted", err, repository.ErrNotFound)
		_, err = k.get(ctx, r, key("b"))
		must(t, "Get other record after Delete", err)
		wantErr(t, "Delete again", k.del(ctx, r, key("a")), repository.ErrNotFound)
	})

	t.Run("ReturnsCopies", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		in := k.fixture("a")
		_, err := k.create(ctx, r, in)
		must(t, "Create", err)
		// Changing the value passed to Create must not change the stored one
		*k.fields(in).Slug = "changed"
		(*k.fields(in).Metadata)["k"] = "changed"

		got, err := k.get(ctx, r, key("a"))
		must(t, "Get", err)
		k.same(t, "Get after changing Create input", got, k.fixture("a"))

		// Neither may changing a value returned by Get
		(*k.fields(got).Metadata)["k"] = "changed"
		again, err := k.get(ctx, r, key("a"))
		must(t, "Get again", err)
		k.same(t, "Get after changing Get result", again, k.fixture("a"))
	})
}
