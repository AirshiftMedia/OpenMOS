package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"airshift/openmos/internal/events"
	"airshift/openmos/internal/model"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/xml"
)

// MOSService provides business logic for MOS operations
type MOSService struct {
	runningOrderRepo repository.RunningOrderRepository
	storyRepo        repository.StoryRepository
	itemRepo         repository.ItemRepository
	objectRepo       repository.ObjectRepository
	eventBus         *events.EventBus
}

// NewMOSService creates a new MOS service
func NewMOSService(
	runningOrderRepo repository.RunningOrderRepository,
	storyRepo repository.StoryRepository,
	itemRepo repository.ItemRepository,
	objectRepo repository.ObjectRepository,
	eventBus *events.EventBus,
) *MOSService {
	return &MOSService{
		runningOrderRepo: runningOrderRepo,
		storyRepo:        storyRepo,
		itemRepo:         itemRepo,
		objectRepo:       objectRepo,
		eventBus:         eventBus,
	}
}

// ListRunningOrders returns all running orders
func (s *MOSService) ListRunningOrders(ctx context.Context) ([]*model.RunningOrder, error) {
	return s.runningOrderRepo.List(ctx)
}

// GetRunningOrderWithStories retrieves a running order with all its stories
func (s *MOSService) GetRunningOrderWithStories(ctx context.Context, id string) (*model.RunningOrder, []*model.Story, error) {
	// Get the running order
	ro, err := s.runningOrderRepo.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Get all stories for this running order
	stories, err := s.storyRepo.ListByRunningOrder(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stories: %w", err)
	}

	return ro, stories, nil
}

// GetItemsForStory retrieves all items for a story
func (s *MOSService) GetItemsForStory(ctx context.Context, roID, storyID string) ([]*model.Item, error) {
	return s.itemRepo.ListByStory(ctx, roID, storyID)
}

// ProcessRunningOrderInfo processes a running order creation/update message.
// The stories and items in the message replace those stored before, in the
// order the message lists them.
func (s *MOSService) ProcessRunningOrderInfo(ctx context.Context, roInfo xml.RunningOrderInfo) error {
	// Parse duration if provided; an unparsable value falls back to 0
	var duration int
	if roInfo.Duration != "" {
		if d, parseErr := strconv.Atoi(roInfo.Duration); parseErr == nil {
			duration = d
		}
	}

	existingRO, err := s.runningOrderRepo.Get(ctx, roInfo.ID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		ro := &model.RunningOrder{
			ID:       roInfo.ID,
			Slug:     roInfo.Slug,
			Status:   model.StatusPending,
			Duration: duration,
			Channel:  roInfo.Channel,
		}
		if _, err := s.runningOrderRepo.Create(ctx, ro); err != nil {
			return fmt.Errorf("failed to create running order: %w", err)
		}
	case err != nil:
		return fmt.Errorf("failed to get running order: %w", err)
	default:
		existingRO.Slug = roInfo.Slug
		existingRO.Channel = roInfo.Channel
		existingRO.Duration = duration
		if err := s.runningOrderRepo.Update(ctx, existingRO); err != nil {
			return fmt.Errorf("failed to update running order: %w", err)
		}
	}

	stories := make([]*model.Story, len(roInfo.Stories))
	for i, storyInfo := range roInfo.Stories {
		stories[i] = &model.Story{
			ID:             storyInfo.ID,
			RunningOrderID: roInfo.ID,
			Slug:           storyInfo.Slug,
			Number:         storyInfo.Number,
			Status:         model.StatusPending,
		}
	}
	if err := s.storyRepo.ReplaceAll(ctx, roInfo.ID, stories); err != nil {
		return fmt.Errorf("failed to replace stories: %w", err)
	}

	for _, storyInfo := range roInfo.Stories {
		items := make([]*model.Item, len(storyInfo.Items))
		for j, itemInfo := range storyInfo.Items {
			items[j] = &model.Item{
				ID:             itemInfo.ID,
				RunningOrderID: roInfo.ID,
				StoryID:        storyInfo.ID,
				ObjectID:       itemInfo.ObjectID,
				MosID:          itemInfo.MosID,
				Slug:           itemInfo.Slug,
				Status:         model.StatusPending,
			}
			// An unparsable or out-of-range duration falls back to 0
			if d, err := strconv.ParseUint(itemInfo.Duration, 10, 32); err == nil {
				items[j].Duration = int64(d)
			}
		}
		if err := s.itemRepo.ReplaceAll(ctx, roInfo.ID, storyInfo.ID, items); err != nil {
			return fmt.Errorf("failed to replace items of story %s: %w", storyInfo.ID, err)
		}
	}

	// Publish event after successful update
	if s.eventBus != nil {
		s.eventBus.Publish(events.Event{
			Type:    events.RunningOrderUpdated,
			Payload: roInfo.ID,
			Source:  "mos_service",
		})
	}

	return nil
}
