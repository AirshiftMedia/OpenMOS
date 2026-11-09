package model

import (
	"time"
)

// The model follows the MOS keys: an ID is unique only within its parent, so
// stories and items carry their parents' IDs, and objects their MOS device.
// Order is the position among siblings and is managed by the repositories.
// Durations of items and objects are counts of samples at TimeBase samples
// per second (MOS objTB, e.g. 50 for PAL or 59.94 for NTSC).

// MOSObject represents the lowest level media object in the MOS hierarchy
type MOSObject struct {
	MosID      string            `json:"mosID"`      // MOS device that owns the object
	ID         string            `json:"id"`         // objID, unique within MosID
	ObjectType string            `json:"objectType"` // STILL, AUDIO or VIDEO; empty if unknown
	Slug       string            `json:"slug"`       // Human-readable name
	Duration   int64             `json:"duration"`   // objDur, in samples
	TimeBase   float64           `json:"timeBase,omitempty"`
	Status     StatusType        `json:"status"`
	Abstract   string            `json:"abstract,omitempty"` // mosAbstract
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// Item represents a single item within a story
type Item struct {
	RunningOrderID string            `json:"runningOrderID"`     // Parent running order
	StoryID        string            `json:"storyID"`            // Parent story
	ID             string            `json:"id"`                 // itemID, unique within its story
	ObjectID       string            `json:"objectID,omitempty"` // Referenced MOS object (objID)
	MosID          string            `json:"mosID,omitempty"`    // MOS device of the referenced object
	Slug           string            `json:"slug"`
	Duration       int64             `json:"duration"` // itemEdDur, in samples
	TimeBase       float64           `json:"timeBase,omitempty"`
	Status         StatusType        `json:"status"`
	Order          int               `json:"order"` // Position within the story
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// Story represents a story in the running order (collection of items)
type Story struct {
	RunningOrderID string            `json:"runningOrderID"` // Parent running order
	ID             string            `json:"id"`             // storyID, unique within its running order
	Slug           string            `json:"slug"`
	Number         string            `json:"number,omitempty"`
	Status         StatusType        `json:"status"`
	Order          int               `json:"order"` // Position within the running order
	Presenter      string            `json:"presenter,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// RunningOrder represents the top-level running order (collection of stories)
type RunningOrder struct {
	ID        string            `json:"id"`    // roID
	MosID     string            `json:"mosID"` // MOS ID for this running order
	Slug      string            `json:"slug"`
	Status    StatusType        `json:"status"`
	Duration  int               `json:"duration"`          // roEdDur, in seconds
	AirTime   *time.Time        `json:"airTime,omitempty"` // roEdStart
	Channel   string            `json:"channel,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Version   int               `json:"version"`
	CreatedBy string            `json:"createdBy,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}
