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

// MongoStoryRepository implements StoryRepository for MongoDB
type MongoStoryRepository struct {
	db         db.Database
	collection *mongo.Collection
}

// NewMongoStoryRepository creates a new MongoDB story repository
func NewMongoStoryRepository(database db.Database) *MongoStoryRepository {
	return &MongoStoryRepository{
		db:         database,
		collection: database.Collection("stories"),
	}
}

// Create creates a new story
func (r *MongoStoryRepository) Create(ctx context.Context, story *model.Story) (*model.Story, error) {
	now := time.Now()
	story.CreatedAt = now
	story.UpdatedAt = now

	// Ensure ID uniqueness
	if story.ID == "" {
		return nil, errors.New("story ID is required")
	}

	// Check if ID already exists
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": story.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to check story ID uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("story with ID %s already exists", story.ID)
	}

	_, err = r.collection.InsertOne(ctx, story)
	if err != nil {
		return nil, fmt.Errorf("failed to create story: %w", err)
	}

	return story, nil
}

// Get retrieves a story by ID
func (r *MongoStoryRepository) Get(ctx context.Context, id string) (*model.Story, error) {
	var story model.Story
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&story)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("story not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get story: %w", err)
	}
	return &story, nil
}

// Update updates a story
func (r *MongoStoryRepository) Update(ctx context.Context, story *model.Story) error {
	now := time.Now()
	story.UpdatedAt = now

	filter := bson.M{"_id": story.ID}
	_, err := r.collection.ReplaceOne(ctx, filter, story)
	if err != nil {
		return fmt.Errorf("failed to update story: %w", err)
	}

	return nil
}

// Delete deletes a story
func (r *MongoStoryRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete story: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("story not found: %s", id)
	}

	return nil
}

// ListByRunningOrder returns all stories for a running order
func (r *MongoStoryRepository) ListByRunningOrder(ctx context.Context, roID string) ([]*model.Story, error) {
	opts := options.Find().SetSort(bson.M{"order": 1})
	cursor, err := r.collection.Find(ctx, bson.M{"runningOrderID": roID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list stories: %w", err)
	}
	defer cursor.Close(ctx)

	var stories []*model.Story
	err = cursor.All(ctx, &stories)
	if err != nil {
		return nil, fmt.Errorf("failed to decode stories: %w", err)
	}

	return stories, nil
}
