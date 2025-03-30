package model

// StatusType represents the status of an element in the running order
type StatusType string

const (
	// StatusPending indicates the element is waiting to be prepared
	StatusPending StatusType = "PENDING"

	// StatusReady indicates the element is ready to be used
	StatusReady StatusType = "READY"

	// StatusActive indicates the element is currently active
	StatusActive StatusType = "ACTIVE"

	// StatusCompleted indicates the element has been successfully completed
	StatusCompleted StatusType = "COMPLETED"

	// StatusSkipped indicates the element was skipped
	StatusSkipped StatusType = "SKIPPED"

	// StatusError indicates an error occurred with the element
	StatusError StatusType = "ERROR"
)
