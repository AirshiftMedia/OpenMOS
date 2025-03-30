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

// MongoObjectRepository implements ObjectRepository for MongoDB
type MongoObjectRepository struct {
	db         db.Database
	collection *mongo.Collection
}

// NewMongoObjectRepository creates a new MongoDB object repository
func NewMongoObjectRepository(database db.Database) *MongoObjectRepository {
	return &MongoObjectRepository{
		db:         database,
		collection: database.Collection("mosObjects"),
	}
}

// Create creates a new MOS object
func (r *MongoObjectRepository) Create(ctx context.Context, obj *model.MOSObject) (*model.MOSObject, error) {
	now := time.Now()
	obj.CreatedAt = now
	obj.UpdatedAt = now

	// Ensure ID uniqueness
	if obj.ID == "" {
		return nil, errors.New("object ID is required")
	}

	// Check if ID already exists
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": obj.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to check object ID uniqueness: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("object with ID %s already exists", obj.ID)
	}

	_, err = r.collection.InsertOne(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	return obj, nil
}

// Get retrieves a MOS object by ID
func (r *MongoObjectRepository) Get(ctx context.Context, id string) (*model.MOSObject, error) {
	var obj model.MOSObject
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&obj)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("object not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return &obj, nil
}

// Update updates a MOS object
func (r *MongoObjectRepository) Update(ctx context.Context, obj *model.MOSObject) error {
	now := time.Now()
	obj.UpdatedAt = now

	filter := bson.M{"_id": obj.ID}
	_, err := r.collection.ReplaceOne(ctx, filter, obj)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	return nil
}

// Delete deletes a MOS object
func (r *MongoObjectRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("object not found: %s", id)
	}

	return nil
}

// List returns all MOS objects
func (r *MongoObjectRepository) List(ctx context.Context) ([]*model.MOSObject, error) {
	opts := options.Find().SetSort(bson.M{"createdAt": -1})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	defer cursor.Close(ctx)

	var objects []*model.MOSObject
	err = cursor.All(ctx, &objects)
	if err != nil {
		return nil, fmt.Errorf("failed to decode objects: %w", err)
	}

	return objects, nil
}
