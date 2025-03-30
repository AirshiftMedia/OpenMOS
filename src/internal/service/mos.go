package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
}

// NewMOSService creates a new MOS service
func NewMOSService(
	runningOrderRepo repository.RunningOrderRepository,
	storyRepo repository.StoryRepository,
	itemRepo repository.ItemRepository,
	objectRepo repository.ObjectRepository,
) *MOSService {
	return &MOSService{
		runningOrderRepo: runningOrderRepo,
		storyRepo:        storyRepo,
		itemRepo:         itemRepo,
		objectRepo:       objectRepo,
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
			ID:       roInfo.ID,
			Slug:     roInfo.Slug,
			Status:   model.StatusPending,
			Duration: duration,
			Channel:  roInfo.Channel,
			Version:  1,
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

	// Process stories
	for i, storyInfo := range roInfo.Stories {
		// Check if story exists
		existingStory, err := s.storyRepo.Get(ctx, storyInfo.ID)

		// Parse duration if provided
		var storyDuration int
		if storyInfo.Duration != "" {
			storyDuration, err = strconv.Atoi(storyInfo.Duration)
			if err != nil {
				storyDuration = 0
			}
		}

		if err != nil { // Story doesn't exist
			// Create new story
			story := &model.Story{
				ID:             storyInfo.ID,
				RunningOrderID: roInfo.ID,
				Slug:           storyInfo.Slug,
				Number:         storyInfo.Number,
				Duration:       storyDuration,
				Status:         model.StatusPending,
				Order:          i + 1,
			}

			_, err = s.storyRepo.Create(ctx, story)
			if err != nil {
				return fmt.Errorf("failed to create story: %w", err)
			}

			// Process items
			for j, itemInfo := range storyInfo.Items {
				// Parse duration if provided
				var itemDuration int
				if itemInfo.Duration != "" {
					itemDuration, err = strconv.Atoi(itemInfo.Duration)
					if err != nil {
						itemDuration = 0
					}
				}

				// Create new item
				item := &model.Item{
					ID:       itemInfo.ID,
					StoryID:  storyInfo.ID,
					Slug:     itemInfo.Slug,
					Duration: itemDuration,
					ObjectID: itemInfo.ObjectID,
					Status:   model.StatusPending,
					Order:    j + 1,
				}

				_, err = s.itemRepo.Create(ctx, item)
				if err != nil {
					return fmt.Errorf("failed to create item: %w", err)
				}
			}
		} else {
			// Update existing story
			existingStory.Slug = storyInfo.Slug
			existingStory.Number = storyInfo.Number
			existingStory.Duration = storyDuration
			existingStory.Order = i + 1
			existingStory.UpdatedAt = time.Now()

			err = s.storyRepo.Update(ctx, existingStory)
			if err != nil {
				return fmt.Errorf("failed to update story: %w", err)
			}

			// Process items - more complex, would need to handle deletes, updates, etc.
			// For simplicity, we'll just ensure all items in the message exist
			for j, itemInfo := range storyInfo.Items {
				existingItem, err := s.itemRepo.Get(ctx, itemInfo.ID)

				// Parse duration if provided
				var itemDuration int
				if itemInfo.Duration != "" {
					itemDuration, err = strconv.Atoi(itemInfo.Duration)
					if err != nil {
						itemDuration = 0
					}
				}

				if err != nil { // Item doesn't exist
					// Create new item
					item := &model.Item{
						ID:       itemInfo.ID,
						StoryID:  storyInfo.ID,
						Slug:     itemInfo.Slug,
						Duration: itemDuration,
						ObjectID: itemInfo.ObjectID,
						Status:   model.StatusPending,
						Order:    j + 1,
					}

					_, err = s.itemRepo.Create(ctx, item)
					if err != nil {
						return fmt.Errorf("failed to create item: %w", err)
					}
				} else {
					// Update existing item
					existingItem.Slug = itemInfo.Slug
					existingItem.Duration = itemDuration
					existingItem.ObjectID = itemInfo.ObjectID
					existingItem.Order = j + 1
					existingItem.UpdatedAt = time.Now()

					err = s.itemRepo.Update(ctx, existingItem)
					if err != nil {
						return fmt.Errorf("failed to update item: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// CreateMOSObject creates a new MOS object
func (s *MOSService) CreateMOSObject(ctx context.Context, obj *model.MOSObject) (*model.MOSObject, error) {
	return s.objectRepo.Create(ctx, obj)
}

// GetMOSObject retrieves a MOS object by ID
func (s *MOSService) GetMOSObject(ctx context.Context, id string) (*model.MOSObject, error) {
	return s.objectRepo.Get(ctx, id)
}

// UpdateRunningOrderStatus updates the status of a running order
func (s *MOSService) UpdateRunningOrderStatus(ctx context.Context, id string, status model.StatusType) error {
	ro, err := s.runningOrderRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	ro.Status = status
	ro.UpdatedAt = time.Now()

	return s.runningOrderRepo.Update(ctx, ro)
}
