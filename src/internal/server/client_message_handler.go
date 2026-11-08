package server

import (
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
)

// handleMessage processes a parsed MOS message
func (c *ClientConnection) handleMessage(ctx context.Context, message xml.MOSMessage) error {
	// Create a span for this message handling
	span := sentry.StartSpan(ctx, "handle_message")
	span.SetTag("message_type", message.GetMessageType())
	span.SetTag("client_id", c.id)
	defer span.Finish()

	var err error

	switch msg := message.(type) {
	case xml.Heartbeat:
		err = c.handleHeartbeat(ctx, msg)
	case xml.ReqRunningOrderList:
		err = c.handleReqRunningOrderList(ctx, msg)
	case xml.ReqRunningOrder:
		err = c.handleReqRunningOrder(ctx, msg)
	case xml.RunningOrderInfo:
		err = c.handleRunningOrderInfo(ctx, msg)
	case xml.MOSAck:
		err = c.handleMOSAck(ctx, msg)
	case xml.NCSReqStoryAction:
		err = c.handleNCSReqStoryAction(ctx, msg)
	default:
		err = fmt.Errorf("unknown message type: %T", message)
	}

	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		span.SetData("error", err.Error())
	}

	return err
}

// handleHeartbeat processes a heartbeat message
func (c *ClientConnection) handleHeartbeat(ctx context.Context, heartbeat xml.Heartbeat) error {
	logger.Infof("Received heartbeat from client %s, source: %s", c.id, heartbeat.Source)

	// Record the heartbeat
	c.heartbeat.RecordHeartbeat()

	// Send response
	response, err := c.heartbeat.CreateHeartbeatResponse(heartbeat.RequestID)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat response: %w", err)
	}

	return c.Write(response)
}

// handleMOSAck processes an acknowledgment message
func (c *ClientConnection) handleMOSAck(ctx context.Context, ack xml.MOSAck) error {
	logger.Infof("Received acknowledgment from client %s: %s - %s", c.id, ack.Status, ack.StatusDescription)
	// Just log for now
	return nil
}

// sendErrorAck sends an error acknowledgment
func (c *ClientConnection) sendErrorAck(requestID, status, description string) error {
	ack := xml.CreateMOSAck(c.config.MOS.ID, requestID, status, description)
	data, err := xml.GenerateMessage(ack)
	if err != nil {
		return fmt.Errorf("failed to generate error ack: %w", err)
	}

	return c.Write(data)
}

// sendSuccessAck sends a success acknowledgment
func (c *ClientConnection) sendSuccessAck(requestID, description string) error {
	return c.sendErrorAck(requestID, "ACK", description)
}
