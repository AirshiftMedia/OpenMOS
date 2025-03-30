package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Database defines the interface for database operations
type Database interface {
	// Close closes the database connection
	Close(ctx context.Context) error

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error

	// Collection returns a collection for the given name
	Collection(name string) *mongo.Collection

	// Database returns the underlying database instance
	Database() *mongo.Database

	// Client returns the underlying client instance
	Client() *mongo.Client

	// CreateIndexes creates the necessary indexes for the collections
	CreateIndexes(ctx context.Context) error
}
