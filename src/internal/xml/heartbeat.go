package xml

import (
	"context"
	"fmt"
	"log"
	"time"
)

// HeartbeatMonitor handles heartbeat messages for client connections
type HeartbeatMonitor struct {
	source        string
	clientID      string
	lastSeen      time.Time
	timeout       time.Duration
	checkInterval time.Duration
	onTimeout     func()
	stopChan      chan struct{}
	heartbeatChan chan struct{}
}

// NewHeartbeatMonitor creates a new heartbeat monitor
func NewHeartbeatMonitor(source, clientID string, timeout, checkInterval time.Duration, onTimeout func()) *HeartbeatMonitor {
	return &HeartbeatMonitor{
		source:        source,
		clientID:      clientID,
		lastSeen:      time.Now(),
		timeout:       timeout,
		checkInterval: checkInterval,
		onTimeout:     onTimeout,
		stopChan:      make(chan struct{}),
		heartbeatChan: make(chan struct{}, 5), // Buffer a few heartbeats
	}
}

// Start starts the heartbeat monitoring
func (h *HeartbeatMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-h.heartbeatChan:
			h.lastSeen = time.Now()
		case <-ticker.C:
			if time.Since(h.lastSeen) > h.timeout {
				log.Printf("Client %s timed out (no heartbeat for %v)", h.clientID, h.timeout)
				if h.onTimeout != nil {
					h.onTimeout()
				}
				return
			}
		}
	}
}

// Stop stops the heartbeat monitoring
func (h *HeartbeatMonitor) Stop() {
	close(h.stopChan)
}

// RecordHeartbeat records a heartbeat from the client
func (h *HeartbeatMonitor) RecordHeartbeat() {
	select {
	case h.heartbeatChan <- struct{}{}:
		// Successfully recorded
	default:
		// Channel full, this is fine as we've already recorded recent heartbeats
	}
}

// TimeSinceLastHeartbeat returns the duration since the last heartbeat
func (h *HeartbeatMonitor) TimeSinceLastHeartbeat() time.Duration {
	return time.Since(h.lastSeen)
}

// LastSeenTime returns the time of the last heartbeat
func (h *HeartbeatMonitor) LastSeenTime() time.Time {
	return h.lastSeen
}

// CreateHeartbeatResponse creates a response to a heartbeat request
func (h *HeartbeatMonitor) CreateHeartbeatResponse(requestID string) ([]byte, error) {
	response := CreateHeartbeatResponse(h.source, requestID)
	return GenerateMessage(response)
}

// ProcessHeartbeat processes an incoming heartbeat message
func ProcessHeartbeat(heartbeat Heartbeat, monitor *HeartbeatMonitor) ([]byte, error) {
	// Record the heartbeat
	monitor.RecordHeartbeat()

	// Create a response
	response, err := monitor.CreateHeartbeatResponse(heartbeat.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to create heartbeat response: %w", err)
	}

	return response, nil
}
