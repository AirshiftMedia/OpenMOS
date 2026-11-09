package service

import (
	"errors"
	"strings"
	"testing"

	"airshift/openmos/internal/events"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/repository/memory"
	"airshift/openmos/internal/xml"
)

// halfFailingROCreate is an roCreate whose second story has an item without
// an ID, so it fails after the running order, the stories and the first
// story's items were written.
var halfFailingROCreate = xml.RunningOrderInfo{ID: "RO1", Slug: "new", Stories: []xml.StoryInfo{
	{ID: "S1", Items: []xml.ItemInfo{{ID: "I1"}}},
	{ID: "S2", Items: []xml.ItemInfo{{ID: ""}}},
}}

func TestHalfFailedROCreateLeavesNoRows(t *testing.T) {
	repo, ctx := memory.New(), t.Context()
	eb := events.NewEventBus()
	sub := eb.Subscribe(events.RunningOrderUpdated, 1)
	svc := NewMOSService(repo, eb)

	err := svc.ProcessRunningOrderInfo(ctx, halfFailingROCreate)
	if !errors.Is(err, repository.ErrInvalid) {
		t.Fatalf("ProcessRunningOrderInfo err = %v, want %v", err, repository.ErrInvalid)
	}
	if _, err := repo.RunningOrders().Get(ctx, "RO1"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Get RO1 after failed roCreate: err = %v, want %v", err, repository.ErrNotFound)
	}
	for _, s := range []string{"S1", "S2"} {
		if _, err := repo.Stories().Get(ctx, "RO1", s); !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("Get story %s after failed roCreate: err = %v, want %v", s, err, repository.ErrNotFound)
		}
	}
	if _, err := repo.Items().Get(ctx, "RO1", "S1", "I1"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Get item I1 after failed roCreate: err = %v, want %v", err, repository.ErrNotFound)
	}
	select {
	case ev := <-sub:
		t.Errorf("failed roCreate published %+v", ev)
	default:
	}
}

// A failed roReplace keeps the running order as it was
func TestHalfFailedROReplaceKeepsOldState(t *testing.T) {
	svc, repo := newTestService()
	ctx := t.Context()
	old := xml.RunningOrderInfo{ID: "RO1", Slug: "old", Stories: []xml.StoryInfo{
		{ID: "S0", Items: []xml.ItemInfo{{ID: "I0"}}},
	}}
	if err := svc.ProcessRunningOrderInfo(ctx, old); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProcessRunningOrderInfo(ctx, halfFailingROCreate); err == nil {
		t.Fatal("ProcessRunningOrderInfo succeeded, want an error")
	}

	if ro := getRO(t, repo, "RO1"); ro.Slug != "old" || ro.Version != 1 {
		t.Errorf("RO1 after failed replace: slug %q, version %d; want %q, 1", ro.Slug, ro.Version, "old")
	}
	if _, err := repo.Items().Get(ctx, "RO1", "S0", "I0"); err != nil {
		t.Errorf("Get item S0/I0 after failed replace: %v", err)
	}
	if _, err := repo.Stories().Get(ctx, "RO1", "S1"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Get story S1 after failed replace: err = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestSuccessfulROCreatePublishesAfterCommit(t *testing.T) {
	repo := memory.New()
	eb := events.NewEventBus()
	sub := eb.Subscribe(events.RunningOrderUpdated, 1)
	svc := NewMOSService(repo, eb)

	if err := svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{ID: "RO1"}); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-sub:
		// A subscriber reacting to the event must see the committed data
		if _, err := repo.RunningOrders().Get(t.Context(), ev.Payload.(string)); err != nil {
			t.Errorf("Get %s on event: %v", ev.Payload, err)
		}
	default:
		t.Error("no event published")
	}
}

// A NEW story action that fails must not leave its auto-created running order
func TestFailedStoryActionRollsBackAutoCreatedRO(t *testing.T) {
	svc, repo := newTestService()
	ctx := t.Context()
	// A story ID over the MOS limit makes the story insert fail after the
	// running order was auto-created
	long := strings.Repeat("x", repository.MaxIDLength+1)
	err := svc.ProcessStoryAction(ctx, xml.NCSReqStoryAction{
		Operation:   "NEW",
		ROStorySend: xml.ROStorySend{ROID: "RO9", StoryID: long},
	})
	if !errors.Is(err, repository.ErrInvalid) {
		t.Fatalf("ProcessStoryAction err = %v, want %v", err, repository.ErrInvalid)
	}
	if _, err := repo.RunningOrders().Get(ctx, "RO9"); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("auto-created RO9 survived the failed action: err = %v", err)
	}
}

func TestStoryActionNewCreatesROAndStory(t *testing.T) {
	svc, repo := newTestService()
	ctx := t.Context()
	err := svc.ProcessStoryAction(ctx, xml.NCSReqStoryAction{
		Operation:   "new",
		ROStorySend: xml.ROStorySend{ROID: "RO9", StoryID: "S1", StorySlug: "first"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s, err := repo.Stories().Get(ctx, "RO9", "S1")
	if err != nil || s.Slug != "first" {
		t.Errorf("story RO9/S1 = %+v, %v; want slug %q", s, err, "first")
	}
}
