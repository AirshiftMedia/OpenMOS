package service

import (
	"context"
	"fmt"
	"testing"

	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/xml"
)

// fakeRORepo is an in-memory RunningOrderRepository. Methods not
// overridden panic via the nil embedded interface.
type fakeRORepo struct {
	repository.RunningOrderRepository
	ros map[string]*model.RunningOrder
}

func (f *fakeRORepo) Get(_ context.Context, id string) (*model.RunningOrder, error) {
	ro, ok := f.ros[id]
	if !ok {
		return nil, fmt.Errorf("running order not found: %s", id)
	}
	c := *ro
	return &c, nil
}

func (f *fakeRORepo) Create(_ context.Context, ro *model.RunningOrder) (*model.RunningOrder, error) {
	f.ros[ro.ID] = ro
	return ro, nil
}

func (f *fakeRORepo) Update(_ context.Context, ro *model.RunningOrder) error {
	f.ros[ro.ID] = ro
	return nil
}

func newTestService() (*MOSService, *fakeRORepo) {
	roRepo := &fakeRORepo{ros: map[string]*model.RunningOrder{}}
	return NewMOSService(roRepo, nil, nil, nil, nil), roRepo
}

// Regression test for MOS-7: a new running order with a numeric duration
// used to take the update branch and dereference a nil running order.
func TestProcessRunningOrderInfoCreatesNewROWithNumericDuration(t *testing.T) {
	svc, repo := newTestService()

	err := svc.ProcessRunningOrderInfo(context.Background(), xml.RunningOrderInfo{
		ID: "RO1", Slug: "Evening News", Duration: "1800",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}

	ro, ok := repo.ros["RO1"]
	if !ok {
		t.Fatal("running order RO1 was not created")
	}
	if ro.Duration != 1800 || ro.Slug != "Evening News" || ro.Version != 1 {
		t.Errorf("created RO = %+v, want duration 1800, slug %q, version 1", ro, "Evening News")
	}
}

func TestProcessRunningOrderInfoUpdatesExistingRO(t *testing.T) {
	svc, repo := newTestService()
	repo.ros["RO1"] = &model.RunningOrder{ID: "RO1", Slug: "Old", Duration: 10, Version: 1}

	err := svc.ProcessRunningOrderInfo(context.Background(), xml.RunningOrderInfo{
		ID: "RO1", Slug: "New", Duration: "1800",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}
	if ro := repo.ros["RO1"]; ro.Slug != "New" || ro.Duration != 1800 {
		t.Errorf("updated RO = %+v, want slug %q, duration 1800", ro, "New")
	}
}

func TestProcessRunningOrderInfoInvalidDurationFallsBackToZero(t *testing.T) {
	svc, repo := newTestService()

	err := svc.ProcessRunningOrderInfo(context.Background(), xml.RunningOrderInfo{
		ID: "RO1", Duration: "not-a-number",
	})
	if err != nil {
		t.Fatalf("ProcessRunningOrderInfo: %v", err)
	}
	if ro := repo.ros["RO1"]; ro == nil || ro.Duration != 0 {
		t.Errorf("RO = %+v, want created with duration 0", ro)
	}
}
