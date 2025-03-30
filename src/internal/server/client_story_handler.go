package server

import (
	"context"
	"encoding/xml"
	"fmt"

	mosxml "airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// handleNCSReqStoryAction processes a request from the NCS to perform an action on a story
func (c *ClientConnection) handleNCSReqStoryAction(ctx context.Context, ncsReq mosxml.NCSReqStoryAction) error {
	// Create a span for this operation
	span := sentry.StartSpan(ctx, "handle_story_action")
	defer span.Finish()

	// Add details to the span
	span.SetTag("operation", ncsReq.Operation)
	span.SetTag("username", ncsReq.Username)

	// Log the action
	logger.Infof("Received story action request: %s from %s", ncsReq.Operation, ncsReq.Username)

	// Process the action
	err := c.server.service.ProcessStoryAction(ctx, ncsReq)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError

		// Log and send error response
		logger.Errorf("Error processing story action: %v", err)
		return c.sendNCSErrorAck("ERROR", fmt.Sprintf("Failed to process story action: %v", err))
	}

	// Send success response
	return c.sendNCSSuccessAck("Story action processed successfully")
}

// sendNCSErrorAck sends an error acknowledgment for NCS requests
func (c *ClientConnection) sendNCSErrorAck(status, description string) error {
	// Create NCS acknowledgment
	ack := mosxml.NCSAck{
		Status:            status,
		StatusDescription: description,
	}

	// Wrap it in a MOS message
	mosMessage := struct {
		XMLName   xml.Name      `xml:"mos"`
		MosID     string        `xml:"mosID"`
		NcsID     string        `xml:"ncsID"`
		MessageID string        `xml:"messageID"`
		NCSAck    mosxml.NCSAck `xml:"ncsAck"`
	}{
		MosID:     c.config.MOS.ID,
		NcsID:     "ncs.client.com", // Should come from client context
		MessageID: "1",              // Should be generated
		NCSAck:    ack,
	}

	// Generate the XML
	data, err := xml.Marshal(mosMessage)
	if err != nil {
		return fmt.Errorf("failed to generate NCS ack: %w", err)
	}

	// Send the response
	return c.Write(data)
}

// sendNCSSuccessAck sends a success acknowledgment for NCS requests
func (c *ClientConnection) sendNCSSuccessAck(description string) error {
	return c.sendNCSErrorAck("ACK", description)
}
