package xml

import (
	"encoding/xml"
	"fmt"
)

// GenerateMessage serializes a MOS message to XML
func GenerateMessage(message MOSMessage) ([]byte, error) {
	data, err := xml.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML declaration
	result := append([]byte(xml.Header), data...)
	return result, nil
}

// CreateHeartbeat creates a heartbeat message
func CreateHeartbeat(source string, requestID string) Heartbeat {
	return Heartbeat{
		RequestID: requestID,
		Timestamp: Now(),
		Source:    source,
	}
}

// CreateHeartbeatResponse creates a heartbeat response message
func CreateHeartbeatResponse(source string, requestID string) Heartbeat {
	return Heartbeat{
		RequestID: requestID,
		Timestamp: Now(),
		Source:    source,
	}
}

// CreateMOSAck creates an acknowledgment message
func CreateMOSAck(source string, requestID string, status string, description string) MOSAck {
	return MOSAck{
		RequestID:         requestID,
		Timestamp:         Now(),
		Source:            source,
		Status:            status,
		StatusDescription: description,
	}
}

// CreateRunningOrderList creates a running order list message
func CreateRunningOrderList(source string, requestID string, items []ROListItem) RunningOrderList {
	return RunningOrderList{
		RequestID:    requestID,
		Timestamp:    Now(),
		Source:       source,
		RunningOrder: items,
	}
}

// CreateRunningOrderInfo creates a full running order message
func CreateRunningOrderInfo(source string, requestID string, id string, slug string,
	channel string, editTime string, startTime string, duration string,
	stories []StoryInfo) RunningOrderInfo {

	return RunningOrderInfo{
		RequestID: requestID,
		Timestamp: Now(),
		Source:    source,
		ID:        id,
		Slug:      slug,
		Channel:   channel,
		EditTime:  editTime,
		StartTime: startTime,
		Duration:  duration,
		Stories:   stories,
	}
}

// CreateStoryResponse creates a response to a story creation request
func CreateStoryResponse(requestID, source, status, description string) ([]byte, error) {
	ack := MOSAck{
		RequestID:         requestID,
		Timestamp:         Now(),
		Source:            source,
		Status:            status,
		StatusDescription: description,
	}

	return GenerateMessage(ack)
}
