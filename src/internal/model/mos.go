package model

import (
	"time"
)

// MOSObject represents the lowest level media object in the MOS hierarchy
type MOSObject struct {
	ID          string            `bson:"_id" json:"id"`                // Unique MOS Object ID
	ObjectType  string            `bson:"objectType" json:"objectType"` // Type of object (e.g., VIDEO, AUDIO, GRAPHIC)
	Slug        string            `bson:"slug" json:"slug"`             // Human-readable name
	Duration    int               `bson:"duration" json:"duration"`     // Duration in seconds
	TimeBase    int               `bson:"timeBase,omitempty" json:"timeBase,omitempty"`
	Status      StatusType        `bson:"status" json:"status"`
	ObjectID    string            `bson:"objectID,omitempty" json:"objectID,omitempty"`
	MediaID     string            `bson:"mediaID,omitempty" json:"mediaID,omitempty"`
	MosAbstract string            `bson:"mosAbstract,omitempty" json:"mosAbstract,omitempty"`
	Metadata    map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// Item represents a single item within a story
type Item struct {
	ID                string            `bson:"_id" json:"id"`                                // Unique Item ID
	ObjectID          string            `bson:"objectID,omitempty" json:"objectID,omitempty"` // Reference to MOS Object
	Slug              string            `bson:"slug" json:"slug"`
	Duration          int               `bson:"duration" json:"duration"` // Duration in seconds
	EditorialDuration int               `bson:"editorialDuration,omitempty" json:"editorialDuration,omitempty"`
	TimeBase          int               `bson:"timeBase,omitempty" json:"timeBase,omitempty"`
	Status            StatusType        `bson:"status" json:"status"`
	Order             int               `bson:"order" json:"order"`     // Order within the story
	StoryID           string            `bson:"storyID" json:"storyID"` // Parent story ID
	Metadata          map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt         time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// Story represents a story in the running order (collection of items)
type Story struct {
	ID             string            `bson:"_id" json:"id"`                        // Unique Story ID
	RunningOrderID string            `bson:"runningOrderID" json:"runningOrderID"` // Parent running order
	Slug           string            `bson:"slug" json:"slug"`
	Number         string            `bson:"number,omitempty" json:"number,omitempty"`
	Duration       int               `bson:"duration" json:"duration"` // Duration in seconds
	Status         StatusType        `bson:"status" json:"status"`
	Order          int               `bson:"order" json:"order"`                               // Order within the running order
	PreviousID     string            `bson:"previousID,omitempty" json:"previousID,omitempty"` // Previous story ID for linked list
	NextID         string            `bson:"nextID,omitempty" json:"nextID,omitempty"`         // Next story ID for linked list
	Presenter      string            `bson:"presenter,omitempty" json:"presenter,omitempty"`
	Metadata       map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt      time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time         `bson:"updatedAt" json:"updatedAt"`
}

// RunningOrder represents the top-level running order (collection of stories)
type RunningOrder struct {
	ID           string            `bson:"_id" json:"id"`      // Unique Running Order ID
	MosID        string            `bson:"mosID" json:"mosID"` // MOS ID for this running order
	Slug         string            `bson:"slug" json:"slug"`
	Status       StatusType        `bson:"status" json:"status"`
	Duration     int               `bson:"duration" json:"duration"`                             // Total duration in seconds
	FirstStoryID string            `bson:"firstStoryID,omitempty" json:"firstStoryID,omitempty"` // First story ID for linked list
	LastStoryID  string            `bson:"lastStoryID,omitempty" json:"lastStoryID,omitempty"`   // Last story ID for linked list
	AirTime      *time.Time        `bson:"airTime,omitempty" json:"airTime,omitempty"`
	Channel      string            `bson:"channel,omitempty" json:"channel,omitempty"`
	Metadata     map[string]string `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Version      int               `bson:"version" json:"version"`
	CreatedBy    string            `bson:"createdBy,omitempty" json:"createdBy,omitempty"`
	CreatedAt    time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time         `bson:"updatedAt" json:"updatedAt"`
}
