package db

import (
	"context"
	"fmt"

	"airshift/openmos/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB implements the database interface for MongoDB
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.Config
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.Timeout)
	defer cancel()

	// Create client options
	clientOptions := options.Client().ApplyURI(cfg.Mongo.URI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database
	database := client.Database(cfg.Mongo.Database)

	return &MongoDB{
		client:   client,
		database: database,
		config:   cfg,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Database returns the mongo.Database instance
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Client returns the mongo.Client instance
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

// Collection returns a mongo.Collection for the given name
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// Ping checks if the database connection is alive
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// CreateIndexes creates the necessary indexes for the collections
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// This function would create all the necessary indexes for our collections
	// We'll implement it as needed when we define our schema requirements
	return nil
}
