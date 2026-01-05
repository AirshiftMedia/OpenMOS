package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
func (s *MOSService) GetItemsForStory(ctx context.Context, storyID string) ([]*model.Item, error) {
	return s.itemRepo.ListByStory(ctx, storyID)
}

// ProcessRunningOrderInfo processes a running order creation/update message
func (s *MOSService) ProcessRunningOrderInfo(ctx context.Context, roInfo xml.RunningOrderInfo) error {
	// Check if running order exists
	existingRO, err := s.runningOrderRepo.Get(ctx, roInfo.ID)

	// Parse duration if provided
	var duration int
	if roInfo.Duration != "" {
		duration, err = strconv.Atoi(roInfo.Duration)
		if err != nil {
			duration = 0
		}
	}

	// Create or update running order
	if err != nil { // Running order doesn't exist
		// Create new running order
		ro := &model.RunningOrder{
			ID:        roInfo.ID,
			Slug:      roInfo.Slug,
			Status:    model.StatusPending,
			Duration:  duration,
			Channel:   roInfo.Channel,
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = s.runningOrderRepo.Create(ctx, ro)
		if err != nil {
			return fmt.Errorf("failed to create running order: %w", err)
		}
	} else {
		// Update existing running order
		existingRO.Slug = roInfo.Slug
		existingRO.Channel = roInfo.Channel
		existingRO.Duration = duration
		existingRO.UpdatedAt = time.Now()

		err = s.runningOrderRepo.Update(ctx, existingRO)
		if err != nil {
			return fmt.Errorf("failed to update running order: %w", err)
		}
	}

	// Process stories (simplified - full implementation would handle deletions, etc.)
	for i, storyInfo := range roInfo.Stories {
		// Create or update each story
		story := &model.Story{
			ID:             storyInfo.ID,
			RunningOrderID: roInfo.ID,
			Slug:           storyInfo.Slug,
			Number:         storyInfo.Number,
			Status:         model.StatusPending,
			Order:          i + 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Parse duration if provided
		if storyInfo.Duration != "" {
			if storyDuration, err := strconv.Atoi(storyInfo.Duration); err == nil {
				story.Duration = storyDuration
			}
		}

		// Create or update the story
		existingStory, err := s.storyRepo.Get(ctx, storyInfo.ID)
		if err != nil {
			// Story doesn't exist, create it
			_, err = s.storyRepo.Create(ctx, story)
			if err != nil {
				return fmt.Errorf("failed to create story: %w", err)
			}
		} else {
			// Story exists, update it
			existingStory.Slug = storyInfo.Slug
			existingStory.Number = storyInfo.Number
			existingStory.Order = i + 1
			existingStory.UpdatedAt = time.Now()

			if storyInfo.Duration != "" {
				if storyDuration, err := strconv.Atoi(storyInfo.Duration); err == nil {
					existingStory.Duration = storyDuration
				}
			}

			err = s.storyRepo.Update(ctx, existingStory)
			if err != nil {
				return fmt.Errorf("failed to update story: %w", err)
			}
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
