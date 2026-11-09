package repotest

import (
	"errors"
	"testing"
	"time"

	"airshift/openmos/internal/repository"
)

var errAbort = errors.New("abort transaction")

func testTx(t *testing.T, newRepo Factory) {
	t.Run("Commit", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		err := r.InTx(ctx, func(tx repository.Repository) error {
			createRunningOrder(t, tx, "ro")
			createStory(t, tx, "ro", "s")
			createItem(t, tx, "ro", "s", "i")
			createObject(t, tx, testMOS, "o")
			return nil
		})
		must(t, "InTx", err)
		exist(t, r, []string{"ro", "ro/s", "ro/s/i"}, "ro", "ro/s", "ro/s/i")
		_, err = r.Objects().Get(ctx, testMOS, "o")
		must(t, "Get object after commit", err)
	})

	t.Run("RollbackOnError", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createRunningOrder(t, r, "kept")
		createStory(t, r, "kept", "s")
		err := r.InTx(ctx, func(tx repository.Repository) error {
			createRunningOrder(t, tx, "ro")
			createStory(t, tx, "ro", "s")
			createItem(t, tx, "ro", "s", "i")
			createObject(t, tx, testMOS, "o")
			ro, err := tx.RunningOrders().Get(ctx, "kept")
			must(t, "Get in tx", err)
			ro.Slug = "changed"
			must(t, "Update in tx", tx.RunningOrders().Update(ctx, ro))
			must(t, "Delete in tx", tx.Stories().Delete(ctx, "kept", "s"))
			return errAbort
		})
		wantErr(t, "InTx", err, errAbort)
		assertRolledBack(t, r)
	})

	t.Run("RollbackRestoresOrder", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createRunningOrder(t, r, "ro")
		createStory(t, r, "ro", "s")
		for _, id := range []string{"a", "b", "c"} {
			createStory(t, r, "ro", id)
			createItem(t, r, "ro", "s", id)
		}
		err := r.InTx(ctx, func(tx repository.Repository) error {
			must(t, "Swap stories in tx", tx.Stories().Swap(ctx, "ro", "a", "c"))
			must(t, "Move items in tx", tx.Items().Move(ctx, "ro", "s", []string{"c"}, "a"))
			return errAbort
		})
		wantErr(t, "InTx", err, errAbort)
		stories, err := r.Stories().ListByRunningOrder(ctx, "ro")
		must(t, "list stories", err)
		equalIDs(t, "stories after rollback", storyKind.ids(stories), []string{"s", "a", "b", "c"})
		items, err := r.Items().ListByStory(ctx, "ro", "s")
		must(t, "list items", err)
		equalIDs(t, "items after rollback", itemKind.ids(items), []string{"a", "b", "c"})
	})

	t.Run("RollbackOnPanic", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		createRunningOrder(t, r, "kept")
		createStory(t, r, "kept", "s")
		func() {
			defer func() {
				if recover() == nil {
					t.Error("InTx swallowed the panic")
				}
			}()
			_ = r.InTx(ctx, func(tx repository.Repository) error {
				createRunningOrder(t, tx, "ro")
				must(t, "Delete in tx", tx.Stories().Delete(ctx, "kept", "s"))
				panic("boom")
			})
		}()
		assertRolledBack(t, r)
		// The repository is still usable
		createRunningOrder(t, r, "after")
	})

	t.Run("Isolation", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		err := r.InTx(ctx, func(tx repository.Repository) error {
			createRunningOrder(t, tx, "ro")
			_, err := tx.RunningOrders().Get(ctx, "ro")
			must(t, "Get own write in tx", err)
			_, err = r.RunningOrders().Get(ctx, "ro")
			wantErr(t, "Get uncommitted write from outside", err, repository.ErrNotFound)
			list, err := r.RunningOrders().List(ctx)
			must(t, "List from outside", err)
			equalIDs(t, "List from outside", runningOrderKind.ids(list), nil)
			return nil
		})
		must(t, "InTx", err)
	})

	// A write outside the transaction while it runs must survive its commit
	t.Run("DirectWriteDuringTx", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		direct := make(chan error, 1)
		err := r.InTx(ctx, func(tx repository.Repository) error {
			createRunningOrder(t, tx, "tx")
			go func() {
				_, err := r.RunningOrders().Create(ctx, newRunningOrder("direct"))
				direct <- err
			}()
			time.Sleep(20 * time.Millisecond) // let the direct write run or block
			return nil
		})
		must(t, "InTx", err)
		must(t, "direct Create", <-direct)
		exist(t, r, []string{"tx", "direct"}, "tx", "direct")
	})

	t.Run("NestedRollback", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		err := r.InTx(ctx, func(tx repository.Repository) error {
			createRunningOrder(t, tx, "outer")
			err := tx.InTx(ctx, func(inner repository.Repository) error {
				createRunningOrder(t, inner, "inner")
				createStory(t, inner, "outer", "s")
				return errAbort
			})
			wantErr(t, "inner InTx", err, errAbort)
			createStory(t, tx, "outer", "after")
			return nil
		})
		must(t, "outer InTx", err)
		exist(t, r, []string{"outer", "inner", "outer/s", "outer/after"}, "outer", "outer/after")
	})

	t.Run("NestedCommitRolledBackByOuter", func(t *testing.T) {
		r, ctx := newRepo(t), t.Context()
		err := r.InTx(ctx, func(tx repository.Repository) error {
			must(t, "inner InTx", tx.InTx(ctx, func(inner repository.Repository) error {
				createRunningOrder(t, inner, "inner")
				return nil
			}))
			return errAbort
		})
		wantErr(t, "outer InTx", err, errAbort)
		exist(t, r, []string{"inner"})
	})
}

// assertRolledBack checks the state the rollback tests started from: only
// running order kept with its story s and unchanged slug, and no object.
func assertRolledBack(t *testing.T, r repository.Repository) {
	t.Helper()
	ctx := t.Context()
	exist(t, r, []string{"ro", "ro/s", "ro/s/i", "kept", "kept/s"}, "kept", "kept/s")
	ro, err := r.RunningOrders().Get(ctx, "kept")
	must(t, "Get kept", err)
	if ro.Slug != "slug kept" || ro.Version != 1 {
		t.Errorf("kept after rollback: slug %q, version %d; want %q, 1", ro.Slug, ro.Version, "slug kept")
	}
	_, err = r.Objects().Get(ctx, testMOS, "o")
	wantErr(t, "Get object after rollback", err, repository.ErrNotFound)
}
