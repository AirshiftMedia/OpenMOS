package repotest

import (
	"testing"

	"airshift/openmos/internal/repository"
)

func testObjects(t *testing.T, newRepo Factory) {
	t.Run("ListNewestFirst", func(t *testing.T) {
		r := newRepo(t)
		list, err := r.Objects().List(t.Context())
		must(t, "List empty", err)
		equalIDs(t, "List empty", objectKind.ids(list), nil)

		for _, id := range []string{"a", "b", "c"} {
			createObject(t, r, testMOS, id)
			tick()
		}
		list, err = r.Objects().List(t.Context())
		must(t, "List", err)
		equalIDs(t, "List", objectKind.ids(list), []string{"c", "b", "a"})
	})

	// An objID is unique only within its MOS device
	t.Run("KeyIncludesMosID", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createObject(t, r, "a.mos", "obj")
		createObject(t, r, "b.mos", "obj")
		must(t, "Delete a.mos/obj", r.Objects().Delete(ctx, "a.mos", "obj"))
		_, err := r.Objects().Get(ctx, "b.mos", "obj")
		must(t, "Get b.mos/obj after deleting a.mos/obj", err)
		_, err = r.Objects().Get(ctx, "c.mos", "obj")
		wantErr(t, "Get obj of an unknown MOS device", err, repository.ErrNotFound)
	})
}
