package postgres

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/repository/repotest"
	"airshift/openmos/internal/service"
	"airshift/openmos/internal/xml"
)

// TestContract runs the repository contract suite against the test
// database. The schema is migrated once and the tables are emptied before
// each subtest.
func TestContract(t *testing.T) {
	migrated(t)
	pool, err := pgxpool.New(t.Context(), testDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	repo := New(pool)
	repotest.Run(t, func(t *testing.T) repository.Repository {
		_, err := pool.Exec(t.Context(), "TRUNCATE openmos.running_orders, openmos.mos_objects CASCADE")
		if err != nil {
			t.Fatalf("empty tables: %v", err)
		}
		return repo
	})
}

// Appends to one parent from several connections must not race for the
// next position: each is serialised by the lock on the parent row.
func TestConcurrentAppends(t *testing.T) {
	db := migrated(t)
	pool, err := pgxpool.New(t.Context(), testDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repo, ctx := New(pool), t.Context()
	if _, err := repo.RunningOrders().Create(ctx, &model.RunningOrder{ID: "ro"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Stories().Create(ctx, &model.Story{RunningOrderID: "ro", ID: "s"}); err != nil {
		t.Fatal(err)
	}

	const n = 20
	var wg sync.WaitGroup
	errs := make(chan error, 2*n)
	for i := range n {
		wg.Go(func() {
			id := fmt.Sprint("c", i)
			_, err := repo.Stories().Create(ctx, &model.Story{RunningOrderID: "ro", ID: id})
			errs <- err
			_, err = repo.Items().Create(ctx, &model.Item{RunningOrderID: "ro", StoryID: "s", ID: id})
			errs <- err
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent Create: %v", err)
		}
	}
	// The unique constraints already rule out duplicates; positions must
	// also have no gaps
	for _, table := range []string{"stories", "items"} {
		want := map[string]int{"stories": n, "items": n - 1}[table]
		if got := count(t, db, "SELECT max(position) FROM openmos."+table); got != want {
			t.Errorf("last %s position %d, want %d", table, got, want)
		}
	}
}

// MOS-15: an roCreate that fails halfway, after the running order, its
// stories and some items were written, must leave no rows behind.
func TestHalfFailedROCreateLeavesNoRows(t *testing.T) {
	db := migrated(t)
	pool, err := pgxpool.New(t.Context(), testDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	svc := service.NewMOSService(New(pool), nil)

	err = svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{ID: "RO1", Stories: []xml.StoryInfo{
		{ID: "S1", Items: []xml.ItemInfo{{ID: "I1"}}},
		{ID: "S2", Items: []xml.ItemInfo{{ID: "I2", Duration: "1"}, {ID: "I2"}}}, // duplicate item
	}})
	if !errors.Is(err, repository.ErrInvalid) {
		t.Fatalf("ProcessRunningOrderInfo err = %v, want %v", err, repository.ErrInvalid)
	}
	for _, table := range []string{"running_orders", "stories", "items"} {
		if n := count(t, db, "SELECT count(*) FROM openmos."+table); n != 0 {
			t.Errorf("%d rows left in %s, want 0", n, table)
		}
	}
}
