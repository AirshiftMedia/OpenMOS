package repotest

import (
	"context"
	"testing"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
)

// ordered describes the children of one parent (stories of a running order
// or items of a story) to the shared ordering tests. A parent is named by a
// single string; the item variant places both test stories in one running
// order.
type ordered struct {
	parents    func(t *testing.T, r repository.Repository, names ...string)
	create     func(t *testing.T, r repository.Repository, parent, id string)
	list       func(ctx context.Context, r repository.Repository, parent string) (ids []string, orders []int, err error)
	del        func(ctx context.Context, r repository.Repository, parent, id string) error
	setOrder   func(ctx context.Context, r repository.Repository, parent, id string, order int) error // Update with Order changed
	move       func(ctx context.Context, r repository.Repository, parent string, ids []string, before string) error
	swap       func(ctx context.Context, r repository.Repository, parent, a, b string) error
	replaceAll func(ctx context.Context, r repository.Repository, parent string, ids []string) error
}

func testOrdering(t *testing.T, newRepo Factory, o ordered) {
	// setup creates parents p and q, with children a, b, c, d in p
	setup := func(t *testing.T) repository.Repository {
		r := newRepo(t)
		o.parents(t, r, "p", "q")
		for _, id := range []string{"a", "b", "c", "d"} {
			o.create(t, r, "p", id)
		}
		return r
	}
	// check fails unless parent lists want, with strictly increasing Order
	check := func(t *testing.T, r repository.Repository, op, parent string, want ...string) {
		t.Helper()
		ids, orders, err := o.list(t.Context(), r, parent)
		must(t, op+": list", err)
		equalIDs(t, op, ids, want)
		for i := 1; i < len(orders); i++ {
			if orders[i] <= orders[i-1] {
				t.Errorf("%s: Order values %v not strictly increasing", op, orders)
				break
			}
		}
	}

	t.Run("CreateAppends", func(t *testing.T) {
		r := newRepo(t)
		o.parents(t, r, "p", "q")
		for _, id := range []string{"c", "a", "b"} {
			o.create(t, r, "p", id)
		}
		check(t, r, "after creates", "p", "c", "a", "b")
		check(t, r, "other parent", "q")
		check(t, r, "unknown parent", "missing")
	})

	// A delete leaves a gap; the next create still goes after the last
	t.Run("CreateAfterDelete", func(t *testing.T) {
		r := setup(t)
		must(t, "delete b", o.del(t.Context(), r, "p", "b"))
		o.create(t, r, "p", "aa")
		check(t, r, "create after delete", "p", "a", "c", "d", "aa")
	})

	t.Run("UpdateKeepsOrder", func(t *testing.T) {
		r := setup(t)
		must(t, "Update with a new Order", o.setOrder(t.Context(), r, "p", "a", 99))
		check(t, r, "after Update", "p", "a", "b", "c", "d")
	})

	t.Run("Move", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		must(t, "move d before a", o.move(ctx, r, "p", []string{"d"}, "a"))
		check(t, r, "move d before a", "p", "d", "a", "b", "c")
		must(t, "move a, b to the end", o.move(ctx, r, "p", []string{"a", "b"}, ""))
		check(t, r, "move a, b to the end", "p", "d", "c", "a", "b")
		must(t, "move b, d before c", o.move(ctx, r, "p", []string{"b", "d"}, "c"))
		check(t, r, "move b, d before c", "p", "b", "d", "c", "a")
	})

	t.Run("MoveRejects", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		for _, tc := range []struct {
			name   string
			ids    []string
			before string
			want   error
		}{
			{"unknown child", []string{"x"}, "a", repository.ErrNotFound},
			{"unknown target", []string{"a"}, "x", repository.ErrNotFound},
			{"child of another parent", []string{"a"}, "", repository.ErrNotFound},
			{"target in the moved set", []string{"a", "b"}, "b", repository.ErrInvalid},
			{"child listed twice", []string{"a", "a"}, "c", repository.ErrInvalid},
			{"nothing to move", nil, "a", repository.ErrInvalid},
		} {
			parent := "p"
			if tc.name == "child of another parent" {
				parent = "q"
			}
			wantErr(t, "move "+tc.name, o.move(ctx, r, parent, tc.ids, tc.before), tc.want)
		}
		check(t, r, "after rejected moves", "p", "a", "b", "c", "d")
	})

	t.Run("Swap", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		must(t, "swap a and c", o.swap(ctx, r, "p", "a", "c"))
		check(t, r, "swap a and c", "p", "c", "b", "a", "d")
		wantErr(t, "swap with itself", o.swap(ctx, r, "p", "a", "a"), repository.ErrInvalid)
		wantErr(t, "swap with unknown", o.swap(ctx, r, "p", "a", "x"), repository.ErrNotFound)
		check(t, r, "after rejected swaps", "p", "c", "b", "a", "d")
	})

	t.Run("ReplaceAll", func(t *testing.T) {
		r, ctx := setup(t), t.Context()
		o.create(t, r, "q", "a") // same ID under the other parent
		must(t, "replace with x, c, a", o.replaceAll(ctx, r, "p", []string{"x", "c", "a"}))
		check(t, r, "after replace", "p", "x", "c", "a")
		check(t, r, "other parent untouched", "q", "a")

		wantErr(t, "replace with duplicates", o.replaceAll(ctx, r, "p", []string{"y", "y"}), repository.ErrInvalid)
		wantErr(t, "replace in unknown parent", o.replaceAll(ctx, r, "missing", []string{"y"}), repository.ErrNotFound)
		check(t, r, "after rejected replaces", "p", "x", "c", "a")

		must(t, "replace with nothing", o.replaceAll(ctx, r, "p", nil))
		check(t, r, "after replace with nothing", "p")
		// Appending after a replace continues after the new children
		must(t, "replace with a, b", o.replaceAll(ctx, r, "p", []string{"a", "b"}))
		o.create(t, r, "p", "c")
		check(t, r, "create after replace", "p", "a", "b", "c")
	})
}

var storyOrdering = ordered{
	parents: func(t *testing.T, r repository.Repository, names ...string) {
		for _, n := range names {
			createRunningOrder(t, r, n)
		}
	},
	create: func(t *testing.T, r repository.Repository, parent, id string) { createStory(t, r, parent, id) },
	list: func(ctx context.Context, r repository.Repository, parent string) ([]string, []int, error) {
		list, err := r.Stories().ListByRunningOrder(ctx, parent)
		orders := make([]int, len(list))
		for i, s := range list {
			orders[i] = s.Order
		}
		return storyKind.ids(list), orders, err
	},
	del: func(ctx context.Context, r repository.Repository, parent, id string) error {
		return r.Stories().Delete(ctx, parent, id)
	},
	setOrder: func(ctx context.Context, r repository.Repository, parent, id string, order int) error {
		s, err := r.Stories().Get(ctx, parent, id)
		if err != nil {
			return err
		}
		s.Order = order
		return r.Stories().Update(ctx, s)
	},
	move: func(ctx context.Context, r repository.Repository, parent string, ids []string, before string) error {
		return r.Stories().Move(ctx, parent, ids, before)
	},
	swap: func(ctx context.Context, r repository.Repository, parent, a, b string) error {
		return r.Stories().Swap(ctx, parent, a, b)
	},
	replaceAll: func(ctx context.Context, r repository.Repository, parent string, ids []string) error {
		list := make([]*model.Story, len(ids))
		for i, id := range ids {
			// The parent named in the call wins over the stories' own
			list[i] = newStory("elsewhere", id)
		}
		return r.Stories().ReplaceAll(ctx, parent, list)
	},
}

// Items are tested in stories p and q of running order itemsRO.
const itemsRO = "ro-items"

var itemOrdering = ordered{
	parents: func(t *testing.T, r repository.Repository, names ...string) {
		createRunningOrder(t, r, itemsRO)
		for _, n := range names {
			createStory(t, r, itemsRO, n)
		}
	},
	create: func(t *testing.T, r repository.Repository, parent, id string) { createItem(t, r, itemsRO, parent, id) },
	list: func(ctx context.Context, r repository.Repository, parent string) ([]string, []int, error) {
		list, err := r.Items().ListByStory(ctx, itemsRO, parent)
		orders := make([]int, len(list))
		for i, it := range list {
			orders[i] = it.Order
		}
		return itemKind.ids(list), orders, err
	},
	del: func(ctx context.Context, r repository.Repository, parent, id string) error {
		return r.Items().Delete(ctx, itemsRO, parent, id)
	},
	setOrder: func(ctx context.Context, r repository.Repository, parent, id string, order int) error {
		it, err := r.Items().Get(ctx, itemsRO, parent, id)
		if err != nil {
			return err
		}
		it.Order = order
		return r.Items().Update(ctx, it)
	},
	move: func(ctx context.Context, r repository.Repository, parent string, ids []string, before string) error {
		return r.Items().Move(ctx, itemsRO, parent, ids, before)
	},
	swap: func(ctx context.Context, r repository.Repository, parent, a, b string) error {
		return r.Items().Swap(ctx, itemsRO, parent, a, b)
	},
	replaceAll: func(ctx context.Context, r repository.Repository, parent string, ids []string) error {
		list := make([]*model.Item, len(ids))
		for i, id := range ids {
			// The parent named in the call wins over the items' own
			list[i] = newItem("elsewhere", "elsewhere", id)
		}
		return r.Items().ReplaceAll(ctx, itemsRO, parent, list)
	},
}
