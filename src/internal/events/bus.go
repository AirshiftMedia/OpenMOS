package events

import (
	"sync"
)

// EventType represents the type of event
type EventType string

const (
	RunningOrderUpdated EventType = "ro.updated"
	StoryModified       EventType = "story.modified"
	ItemChanged         EventType = "item.changed"
)

// Event represents an event in the system
type Event struct {
	Type    EventType
	Payload interface{}
	Source  string
}

// EventBus is a simple publish-subscribe event bus
type EventBus struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
}

// NewEventBus creates a new EventBus instance
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

// Subscribe registers a subscriber for a specific event type
// Returns a channel that will receive events of the specified type
func (eb *EventBus) Subscribe(eventType EventType, bufferSize int) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, bufferSize)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

// Publish sends an event to all subscribers of that event type
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if subscribers, ok := eb.subscribers[event.Type]; ok {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Subscriber is slow, skip to avoid blocking
			}
		}
	}
}
