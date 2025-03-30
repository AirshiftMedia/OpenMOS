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

// MongoRunningOrderRepository implements RunningOrderRepository for MongoDB
type MongoRunningOrderRepository struct {
	db         db.Database
	collection *mongo.Collection
}

// NewMongoRunningOrderRepository creates a new MongoDB running order repository
func NewMongoRunningOrderRepository(database db.Database) *MongoRunningOrderRepository {
	return &MongoRunningOrderRepository{
		db:         database,
		collection: database.Collection("runningOrders"),
	}
}

// Create creates a new running order
func (r *MongoRunningOrderRepository) Create(ctx context.Context, ro *model.RunningOrder) (*model.RunningOrder, error) {
	now := time.Now()
	ro.CreatedAt = now
	ro.UpdatedAt = now
	ro.Version = 1

	// Ensure ID uniqueness
	if ro.ID == "" {
		return nil, errors.New("running order ID is required")
	}

	// Check if ID already exists
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": ro.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to check running order ID uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("running order with ID %s already exists", ro.ID)
	}

	_, err = r.collection.InsertOne(ctx, ro)
	if err != nil {
		return nil, fmt.Errorf("failed to create running order: %w", err)
	}

	return ro, nil
}

// Get retrieves a running order by ID
func (r *MongoRunningOrderRepository) Get(ctx context.Context, id string) (*model.RunningOrder, error) {
	var ro model.RunningOrder
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&ro)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("running order not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get running order: %w", err)
	}
	return &ro, nil
}

// Update updates a running order
func (r *MongoRunningOrderRepository) Update(ctx context.Context, ro *model.RunningOrder) error {
	now := time.Now()
	ro.UpdatedAt = now
	ro.Version++

	// Get current document to check version
	var current model.RunningOrder
	err := r.collection.FindOne(ctx, bson.M{"_id": ro.ID}).Decode(&current)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("running order not found: %s", ro.ID)
		}
		return fmt.Errorf("failed to get current running order: %w", err)
	}

	// Optimistic concurrency control
	if ro.Version <= current.Version {
		return fmt.Errorf("running order version conflict: expected > %d, got %d", current.Version, ro.Version)
	}

	filter := bson.M{"_id": ro.ID}
	_, err = r.collection.ReplaceOne(ctx, filter, ro)
	if err != nil {
		return fmt.Errorf("failed to update running order: %w", err)
	}

	return nil
}

// Delete deletes a running order
func (r *MongoRunningOrderRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete running order: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("running order not found: %s", id)
	}

	return nil
}

// List returns all running orders
func (r *MongoRunningOrderRepository) List(ctx context.Context) ([]*model.RunningOrder, error) {
	opts := options.Find().SetSort(bson.M{"createdAt": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list running orders: %w", err)
	}
	defer cursor.Close(ctx)

	var runningOrders []*model.RunningOrder
	err = cursor.All(ctx, &runningOrders)
	if err != nil {
		return nil, fmt.Errorf("failed to decode running orders: %w", err)
	}

	return runningOrders, nil
}
