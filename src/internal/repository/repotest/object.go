package repotest

import "testing"

func testObjects(t *testing.T, newRepo Factory) {
	t.Run("ListNewestFirst", func(t *testing.T) {
		r := newRepo(t)
		list, err := r.Objects().List(t.Context())
		must(t, "List empty", err)
		equalIDs(t, "List empty", objectKind.ids(list), nil)

		for _, id := range []string{"a", "b", "c"} {
			createObject(t, r, id)
			tick()
		}
		list, err = r.Objects().List(t.Context())
		must(t, "List", err)
		equalIDs(t, "List", objectKind.ids(list), []string{"c", "b", "a"})
	})
}
