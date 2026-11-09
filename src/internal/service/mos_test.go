package service

import (
	"slices"
	"testing"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/repository/memory"
	"airshift/openmos/internal/xml"
)

func newTestService() (*MOSService, repository.Repository) {
	repo := memory.New()
	return NewMOSService(repo, nil), repo
}

func getRO(t *testing.T, repo repository.Repository, id string) *model.RunningOrder {
	t.Helper()
	ro, err := repo.RunningOrders().Get(t.Context(), id)
	if err != nil {
		t.Fatalf("Get running order %s: %v", id, err)
	}
	return ro
}

// Regression test for MOS-7: a new running order with a numeric duration
// used to take the update branch and dereference a nil running order.
func TestProcessRunningOrderInfoCreatesNewROWithNumericDuration(t *testing.T) {
	svc, repo := newTestService()

	err := svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{
		ID: "RO1", Slug: "Evening News", Duration: "1800",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}
	if ro := getRO(t, repo, "RO1"); ro.Duration != 1800 || ro.Slug != "Evening News" || ro.Version != 1 {
		t.Errorf("created RO = %+v, want duration 1800, slug %q, version 1", ro, "Evening News")
	}
}

func TestProcessRunningOrderInfoUpdatesExistingRO(t *testing.T) {
	svc, repo := newTestService()
	_, err := repo.RunningOrders().Create(t.Context(), &model.RunningOrder{ID: "RO1", Slug: "Old", Duration: 10})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{
		ID: "RO1", Slug: "New", Duration: "1800",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}
	if ro := getRO(t, repo, "RO1"); ro.Slug != "New" || ro.Duration != 1800 || ro.Version != 2 {
		t.Errorf("updated RO = %+v, want slug %q, duration 1800, version 2", ro, "New")
	}
}

func TestProcessRunningOrderInfoInvalidDurationFallsBackToZero(t *testing.T) {
	svc, repo := newTestService()

	err := svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{
		ID: "RO1", Duration: "not-a-number",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}
	if ro := getRO(t, repo, "RO1"); ro.Duration != 0 {
		t.Errorf("RO = %+v, want duration 0", ro)
	}
}

func TestProcessRunningOrderInfoReplacesStoriesAndItems(t *testing.T) {
	svc, repo := newTestService()
	ctx := t.Context()

	first := xml.RunningOrderInfo{ID: "RO1", Stories: []xml.StoryInfo{
		{ID: "S1", Items: []xml.ItemInfo{{ID: "I1"}, {ID: "I2"}}},
		{ID: "S2"},
	}}
	if err := svc.ProcessRunningOrderInfo(ctx, first); err != nil {
		t.Fatalf("first ProcessRunningOrderInfo: %v", err)
	}

	// The second message reorders the stories, drops S1's items except one
	// and adds an item to S2
	second := xml.RunningOrderInfo{ID: "RO1", Stories: []xml.StoryInfo{
		{ID: "S2", Items: []xml.ItemInfo{{ID: "I9", MosID: "a.mos", ObjectID: "O1", Duration: "250"}}},
		{ID: "S1", Items: []xml.ItemInfo{{ID: "I2"}}},
	}}
	if err := svc.ProcessRunningOrderInfo(ctx, second); err != nil {
		t.Fatalf("second ProcessRunningOrderInfo: %v", err)
	}

	_, stories, err := svc.GetRunningOrderWithStories(ctx, "RO1")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range stories {
		ids = append(ids, s.ID)
	}
	if want := []string{"S2", "S1"}; !slices.Equal(ids, want) {
		t.Errorf("stories = %v, want %v", ids, want)
	}

	for story, want := range map[string][]string{"S1": {"I2"}, "S2": {"I9"}} {
		items, err := svc.GetItemsForStory(ctx, "RO1", story)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, it := range items {
			got = append(got, it.ID)
		}
		if !slices.Equal(got, want) {
			t.Errorf("items of %s = %v, want %v", story, got, want)
		}
	}

	it, err := repo.Items().Get(ctx, "RO1", "S2", "I9")
	if err != nil {
		t.Fatal(err)
	}
	if it.MosID != "a.mos" || it.ObjectID != "O1" || it.Duration != 250 {
		t.Errorf("item I9 = %+v, want mosID a.mos, objID O1, duration 250", it)
	}
}

func TestProcessRunningOrderInfoRejectsDuplicateStories(t *testing.T) {
	svc, _ := newTestService()
	err := svc.ProcessRunningOrderInfo(t.Context(), xml.RunningOrderInfo{
		ID: "RO1", Stories: []xml.StoryInfo{{ID: "S1"}, {ID: "S1"}},
	})
	if err == nil {
		t.Fatal("ProcessRunningOrderInfo with a duplicate story succeeded, want an error")
	}
}
