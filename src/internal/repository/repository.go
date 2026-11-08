// Package repository defines the storage interfaces used by the MOS service.
//
// The behaviour every implementation must provide is specified by the
// contract suite in package repotest:
//
//   - Create requires an ID, sets CreatedAt and UpdatedAt, and returns
//     ErrAlreadyExists for an ID in use. Stories and items require their
//     parent to exist, otherwise ErrNotFound.
//   - Get, Update and Delete return ErrNotFound for a missing ID.
//   - Update sets UpdatedAt. Running orders are versioned: Create sets
//     Version to 1, and Update succeeds only if the passed Version equals the
//     stored one, then increments it (also on the passed value). A stale
//     Version returns ErrVersionConflict and changes nothing.
//   - Deleting a running order deletes its stories and their items; deleting
//     a story deletes its items. MOS objects are independent.
//   - List returns running orders and objects newest first. Stories and items
//     are listed by Order, ascending. Listing an unknown parent returns an
//     empty result, not an error.
//   - Returned values are copies: changing them does not change stored data.
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

	// Update updates a running order, checking its version
	Update(ctx context.Context, ro *model.RunningOrder) error

	// Delete deletes a running order with its stories and items
	Delete(ctx context.Context, id string) error

	// List returns all running orders, newest first
	List(ctx context.Context) ([]*model.RunningOrder, error)
}

// StoryRepository defines operations for stories
type StoryRepository interface {
	// Create creates a new story in an existing running order
	Create(ctx context.Context, story *model.Story) (*model.Story, error)

	// Get retrieves a story by ID
	Get(ctx context.Context, id string) (*model.Story, error)

	// Update updates a story
	Update(ctx context.Context, story *model.Story) error

	// Delete deletes a story with its items
	Delete(ctx context.Context, id string) error

	// ListByRunningOrder returns the stories of a running order by Order
	ListByRunningOrder(ctx context.Context, roID string) ([]*model.Story, error)
}

// ItemRepository defines operations for items
type ItemRepository interface {
	// Create creates a new item in an existing story
	Create(ctx context.Context, item *model.Item) (*model.Item, error)

	// Get retrieves an item by ID
	Get(ctx context.Context, id string) (*model.Item, error)

	// Update updates an item
	Update(ctx context.Context, item *model.Item) error

	// Delete deletes an item
	Delete(ctx context.Context, id string) error

	// ListByStory returns the items of a story by Order
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

	// List returns all MOS objects, newest first
	List(ctx context.Context) ([]*model.MOSObject, error)
}

// Repository combines all repositories
type Repository interface {
	RunningOrders() RunningOrderRepository
	Stories() StoryRepository
	Items() ItemRepository
	Objects() ObjectRepository
}
