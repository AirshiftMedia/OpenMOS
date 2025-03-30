package repository

import (
	"context"

	"airshift/openmos/internal/model"
)

// RunningOrderRepository defines operations for running orders
type RunningOrderRepository interface {
	// Create creates a new running order
	Create(ctx context.Context, ro *model.RunningOrder) (*model.RunningOrder, error)

	// Get retrieves a running order by ID
	Get(ctx context.Context, id string) (*model.RunningOrder, error)

	// Update updates a running order
	Update(ctx context.Context, ro *model.RunningOrder) error

	// Delete deletes a running order
	Delete(ctx context.Context, id string) error

	// List returns all running orders
	List(ctx context.Context) ([]*model.RunningOrder, error)
}

// StoryRepository defines operations for stories
type StoryRepository interface {
	// Create creates a new story
	Create(ctx context.Context, story *model.Story) (*model.Story, error)

	// Get retrieves a story by ID
	Get(ctx context.Context, id string) (*model.Story, error)

	// Update updates a story
	Update(ctx context.Context, story *model.Story) error

	// Delete deletes a story
	Delete(ctx context.Context, id string) error

	// ListByRunningOrder returns all stories for a running order
	ListByRunningOrder(ctx context.Context, roID string) ([]*model.Story, error)
}

// ItemRepository defines operations for items
type ItemRepository interface {
	// Create creates a new item
	Create(ctx context.Context, item *model.Item) (*model.Item, error)

	// Get retrieves an item by ID
	Get(ctx context.Context, id string) (*model.Item, error)

	// Update updates an item
	Update(ctx context.Context, item *model.Item) error

	// Delete deletes an item
	Delete(ctx context.Context, id string) error

	// ListByStory returns all items for a story
	ListByStory(ctx context.Context, storyID string) ([]*model.Item, error)
}

// ObjectRepository defines operations for MOS objects
type ObjectRepository interface {
	// Create creates a new MOS object
	Create(ctx context.Context, obj *model.MOSObject) (*model.MOSObject, error)

	// Get retrieves a MOS object by ID
	Get(ctx context.Context, id string) (*model.MOSObject, error)

	// Update updates a MOS object
	Update(ctx context.Context, obj *model.MOSObject) error

	// Delete deletes a MOS object
	Delete(ctx context.Context, id string) error

	// List returns all MOS objects
	List(ctx context.Context) ([]*model.MOSObject, error)
}

// Repository combines all repositories
type Repository interface {
	RunningOrders() RunningOrderRepository
	Stories() StoryRepository
	Items() ItemRepository
	Objects() ObjectRepository
}
