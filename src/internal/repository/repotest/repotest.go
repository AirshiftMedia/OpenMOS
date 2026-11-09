// Package repotest is the contract test suite for repository.Repository.
// Every storage implementation runs it from its own tests:
//
//	func TestContract(t *testing.T) {
//		repotest.Run(t, func(t *testing.T) repository.Repository {
//			return memory.New()
//		})
//	}
//
// The behaviour it checks is summarised in the repository package comment.
package repotest

import (
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"airshift/openmos/internal/repository"
)

// Factory returns an empty repository. It is called once per subtest, and
// subtests run one at a time, so a database-backed factory may reset shared
// tables. It may register cleanup with t.Cleanup.
type Factory func(t *testing.T) repository.Repository

// Run runs the whole contract suite against repositories from newRepo.
func Run(t *testing.T, newRepo Factory) {
	t.Run("RunningOrders", func(t *testing.T) {
		testCRUD(t, newRepo, runningOrderKind)
		testRunningOrders(t, newRepo)
	})
	t.Run("Stories", func(t *testing.T) {
		testCRUD(t, newRepo, storyKind)
		testOrdering(t, newRepo, storyOrdering)
		testStories(t, newRepo)
	})
	t.Run("Items", func(t *testing.T) {
		testCRUD(t, newRepo, itemKind)
		testOrdering(t, newRepo, itemOrdering)
		testItems(t, newRepo)
	})
	t.Run("Objects", func(t *testing.T) {
		testCRUD(t, newRepo, objectKind)
		testObjects(t, newRepo)
	})
	t.Run("Cascade", func(t *testing.T) { testCascade(t, newRepo) })
}

// must stops the test if a setup step fails.
func must(t *testing.T, op string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", op, err)
	}
}

// wantErr fails the test unless err wraps target.
func wantErr(t *testing.T, op string, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: err = %v, want %v", op, err, target)
	}
}

// tick waits long enough that records created before and after it get
// distinct CreatedAt values, even in a store with millisecond precision.
func tick() { time.Sleep(2 * time.Millisecond) }

// clockSkew allows for a store that sets timestamps with its own clock.
const clockSkew = 2 * time.Second

// within reports whether ts lies between from and to, allowing clockSkew.
func within(ts, from, to time.Time) bool {
	return !ts.Before(from.Add(-clockSkew)) && !ts.After(to.Add(clockSkew))
}

// closeTo reports whether a and b are equal at the precision a store may
// round timestamps to.
func closeTo(a, b time.Time) bool {
	d := a.Sub(b)
	return d > -time.Millisecond && d < time.Millisecond
}

// equalIDs fails the test unless got lists exactly the IDs in want, in order.
func equalIDs(t *testing.T, op string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s: IDs = %v, want %v", op, got, want)
	}
}

func itoa(n int) string { return strconv.Itoa(n) }
