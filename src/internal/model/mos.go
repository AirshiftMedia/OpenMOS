package model

import (
	"time"
)

// MOSObject represents the lowest level media object in the MOS hierarchy
type MOSObject struct {
	ID          string            `json:"id"`         // Unique MOS Object ID
	ObjectType  string            `json:"objectType"` // Type of object (e.g., VIDEO, AUDIO, GRAPHIC)
	Slug        string            `json:"slug"`       // Human-readable name
	Duration    int               `json:"duration"`   // Duration in seconds
	TimeBase    int               `json:"timeBase,omitempty"`
	Status      StatusType        `json:"status"`
	ObjectID    string            `json:"objectID,omitempty"`
	MediaID     string            `json:"mediaID,omitempty"`
	MosAbstract string            `json:"mosAbstract,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// Item represents a single item within a story
type Item struct {
	ID                string            `json:"id"`                 // Unique Item ID
	ObjectID          string            `json:"objectID,omitempty"` // Reference to MOS Object
	Slug              string            `json:"slug"`
	Duration          int               `json:"duration"` // Duration in seconds
	EditorialDuration int               `json:"editorialDuration,omitempty"`
	TimeBase          int               `json:"timeBase,omitempty"`
	Status            StatusType        `json:"status"`
	Order             int               `json:"order"`   // Order within the story
	StoryID           string            `json:"storyID"` // Parent story ID
	Metadata          map[string]string `json:"metadata,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

// Story represents a story in the running order (collection of items)
type Story struct {
	ID             string            `json:"id"`             // Unique Story ID
	RunningOrderID string            `json:"runningOrderID"` // Parent running order
	Slug           string            `json:"slug"`
	Number         string            `json:"number,omitempty"`
	Duration       int               `json:"duration"` // Duration in seconds
	Status         StatusType        `json:"status"`
	Order          int               `json:"order"`                // Order within the running order
	PreviousID     string            `json:"previousID,omitempty"` // Previous story ID for linked list
	NextID         string            `json:"nextID,omitempty"`     // Next story ID for linked list
	Presenter      string            `json:"presenter,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// RunningOrder represents the top-level running order (collection of stories)
type RunningOrder struct {
	ID           string            `json:"id"`    // Unique Running Order ID
	MosID        string            `json:"mosID"` // MOS ID for this running order
	Slug         string            `json:"slug"`
	Status       StatusType        `json:"status"`
	Duration     int               `json:"duration"`               // Total duration in seconds
	FirstStoryID string            `json:"firstStoryID,omitempty"` // First story ID for linked list
	LastStoryID  string            `json:"lastStoryID,omitempty"`  // Last story ID for linked list
	AirTime      *time.Time        `json:"airTime,omitempty"`
	Channel      string            `json:"channel,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Version      int               `json:"version"`
	CreatedBy    string            `json:"createdBy,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}
