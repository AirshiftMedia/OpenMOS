package memory

import (
	"context"
	"maps"

	"airshift/openmos/internal/repository"
)

// lockWrite locks r for a write and returns the unlock function, for
// "defer r.lockWrite()()". Writes also take r.writer, which InTx holds for
// the whole transaction, so a direct write cannot be overwritten when the
// transaction's snapshot is swapped in.
func (r *Repository) lockWrite() func() {
	r.writer.Lock()
	r.mu.Lock()
	return func() {
		r.mu.Unlock()
		r.writer.Unlock()
	}
}

// InTx runs fn on a private copy of the data and swaps the copy in if fn
// returns nil. Readers see the committed data until then. Transactions and
// direct writes run one at a time; fn must write only through the
// Repository it is given, as a direct write to r from fn would wait for the
// transaction forever.
func (r *Repository) InTx(_ context.Context, fn func(repository.Repository) error) error {
	r.writer.Lock()
	defer r.writer.Unlock()

	r.mu.RLock()
	tx := r.snapshot()
	r.mu.RUnlock()

	if err := fn(tx); err != nil {
		return err
	}

	// A panic in fn skips this, discarding the copy
	tx.mu.Lock()
	defer tx.mu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runningOrders, r.stories, r.items, r.objects = tx.runningOrders, tx.stories, tx.items, tx.objects
	// Detach tx, so a leaked reference cannot change r's data
	tx.runningOrders, tx.stories, tx.items, tx.objects = nil, nil, nil, nil
	return nil
}

// snapshot returns a repository holding copies of r's records. The caller
// holds r.mu for reading.
func (r *Repository) snapshot() *Repository {
	return &Repository{
		runningOrders: cloneMap(r.runningOrders, cloneRunningOrder),
		stories:       cloneMap(r.stories, cloneStory),
		items:         cloneMap(r.items, cloneItem),
		objects:       cloneMap(r.objects, cloneObject),
	}
}

func cloneMap[K comparable, T any](m map[K]*T, clone func(*T) *T) map[K]*T {
	out := make(map[K]*T, len(m))
	for k, v := range maps.All(m) {
		out[k] = clone(v)
	}
	return out
}
