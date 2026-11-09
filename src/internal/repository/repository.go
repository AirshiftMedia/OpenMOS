// Package repository defines the storage interfaces used by the MOS service.
//
// The behaviour every implementation must provide is specified by the
// contract suite in package repotest:
//
//   - Keys follow MOS: a running order by roID, a story by (roID, storyID),
//     an item by (roID, storyID, itemID) and an object by (mosID, objID). A
//     story or item ID is unique only within its parent.
//   - Create requires non-empty IDs of at most 128 characters (otherwise
//     ErrInvalid), sets CreatedAt and UpdatedAt, and returns ErrAlreadyExists
//     for a key in use. Stories and items require their parent to exist,
//     otherwise ErrNotFound.
//   - Get, Update and Delete return ErrNotFound for a missing key.
//   - Update sets UpdatedAt and keeps CreatedAt. Running orders are
//     versioned: Create sets Version to 1, and Update succeeds only if the
//     passed Version equals the stored one, then increments it (also on the
//     passed value). A stale Version returns ErrVersionConflict and changes
//     nothing.
//   - Order is managed by the repository: Create appends a story or item
//     after its siblings, Update ignores Order, and Move, Swap and
//     ReplaceAll change it. Listing returns children by Order, ascending.
//     Ordering requests that name a record twice return ErrInvalid; unknown
//     records return ErrNotFound. See MoveIDs and SwapIDs.
//   - Deleting a running order deletes its stories and their items; deleting
//     a story deletes its items. ReplaceAll deletes all existing children
//     with their items, then creates the given ones under the parent named
//     in the call (their own parent IDs are overwritten), as roReplace
//     replaces the whole running order. MOS objects are independent.
//   - List returns running orders and objects newest first. Listing the
//     children of an unknown parent returns an empty result, not an error.
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

	// Get retrieves a running order by roID
	Get(ctx context.Context, roID string) (*model.RunningOrder, error)

	// Update updates a running order, checking its version
	Update(ctx context.Context, ro *model.RunningOrder) error

	// Delete deletes a running order with its stories and items
	Delete(ctx context.Context, roID string) error

	// List returns all running orders, newest first
	List(ctx context.Context) ([]*model.RunningOrder, error)
}

// StoryRepository defines operations for the stories of a running order
type StoryRepository interface {
	// Create appends a new story to its running order
	Create(ctx context.Context, story *model.Story) (*model.Story, error)

	// Get retrieves a story by its running order and storyID
	Get(ctx context.Context, roID, storyID string) (*model.Story, error)

	// Update updates a story's fields; its Order is not changed
	Update(ctx context.Context, story *model.Story) error

	// Delete deletes a story with its items
	Delete(ctx context.Context, roID, storyID string) error

	// ListByRunningOrder returns the stories of a running order by Order
	ListByRunningOrder(ctx context.Context, roID string) ([]*model.Story, error)

	// Move moves the stories, in the given order, to just before beforeID;
	// an empty beforeID moves them to the end (roStoryMove)
	Move(ctx context.Context, roID string, storyIDs []string, beforeID string) error

	// Swap exchanges the positions of two stories (roStorySwap)
	Swap(ctx context.Context, roID, storyID1, storyID2 string) error

	// ReplaceAll replaces all stories of a running order with the given
	// ones, in that order (roCreate or roReplace on an existing order)
	ReplaceAll(ctx context.Context, roID string, stories []*model.Story) error
}

// ItemRepository defines operations for the items of a story
type ItemRepository interface {
	// Create appends a new item to its story
	Create(ctx context.Context, item *model.Item) (*model.Item, error)

	// Get retrieves an item by its running order, story and itemID
	Get(ctx context.Context, roID, storyID, itemID string) (*model.Item, error)

	// Update updates an item's fields; its Order is not changed
	Update(ctx context.Context, item *model.Item) error

	// Delete deletes an item
	Delete(ctx context.Context, roID, storyID, itemID string) error

	// ListByStory returns the items of a story by Order
	ListByStory(ctx context.Context, roID, storyID string) ([]*model.Item, error)

	// Move moves the items, in the given order, to just before beforeID;
	// an empty beforeID moves them to the end (roItemMove)
	Move(ctx context.Context, roID, storyID string, itemIDs []string, beforeID string) error

	// Swap exchanges the positions of two items (roItemSwap)
	Swap(ctx context.Context, roID, storyID, itemID1, itemID2 string) error

	// ReplaceAll replaces all items of a story with the given ones, in that
	// order
	ReplaceAll(ctx context.Context, roID, storyID string, items []*model.Item) error
}

// ObjectRepository defines operations for MOS objects
type ObjectRepository interface {
	// Create creates a new MOS object
	Create(ctx context.Context, obj *model.MOSObject) (*model.MOSObject, error)

	// Get retrieves a MOS object by its MOS device and objID
	Get(ctx context.Context, mosID, objID string) (*model.MOSObject, error)

	// Update updates a MOS object
	Update(ctx context.Context, obj *model.MOSObject) error

	// Delete deletes a MOS object
	Delete(ctx context.Context, mosID, objID string) error

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
