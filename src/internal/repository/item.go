package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"airshift/openmos/internal/db"
	"airshift/openmos/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoItemRepository implements ItemRepository for MongoDB
type MongoItemRepository struct {
	db         db.Database
	collection *mongo.Collection
}

// NewMongoItemRepository creates a new MongoDB item repository
func NewMongoItemRepository(database db.Database) *MongoItemRepository {
	return &MongoItemRepository{
		db:         database,
		collection: database.Collection("items"),
	}
}

// Create creates a new item
func (r *MongoItemRepository) Create(ctx context.Context, item *model.Item) (*model.Item, error) {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	// Ensure ID uniqueness
	if item.ID == "" {
		return nil, errors.New("item ID is required")
	}

	// Check if ID already exists
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": item.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to check item ID uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("item with ID %s already exists", item.ID)
	}

	_, err = r.collection.InsertOne(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return item, nil
}

// Get retrieves an item by ID
func (r *MongoItemRepository) Get(ctx context.Context, id string) (*model.Item, error) {
	var item model.Item
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	return &item, nil
}

// Update updates an item
func (r *MongoItemRepository) Update(ctx context.Context, item *model.Item) error {
	now := time.Now()
	item.UpdatedAt = now

	filter := bson.M{"_id": item.ID}
	_, err := r.collection.ReplaceOne(ctx, filter, item)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// Delete deletes an item
func (r *MongoItemRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("item not found: %s", id)
	}

	return nil
}

// ListByStory returns all items for a story
func (r *MongoItemRepository) ListByStory(ctx context.Context, storyID string) ([]*model.Item, error) {
	opts := options.Find().SetSort(bson.M{"order": 1})
	cursor, err := r.collection.Find(ctx, bson.M{"storyID": storyID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer cursor.Close(ctx)

	var items []*model.Item
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to decode items: %w", err)
	}

	return items, nil
}
