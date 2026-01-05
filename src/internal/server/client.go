package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// ClientConnection represents a connected client
type ClientConnection struct {
	conn       net.Conn
	id         string
	server     *TCPServer // Forward declaration - TCPServer is defined in server.go
	heartbeat  *xml.HeartbeatMonitor
	parser     *xml.MessageParser
	closeChan  chan struct{}
	closeOnce  sync.Once
	writeMutex sync.Mutex
	config     *config.Config
}

// NewClientConnection creates a new client connection
func NewClientConnection(conn net.Conn, server *TCPServer, cfg *config.Config) *ClientConnection {
	clientID := fmt.Sprintf("%s", conn.RemoteAddr())

	client := &ClientConnection{
		conn:      conn,
		id:        clientID,
		server:    server,
		parser:    xml.NewMessageParser(),
		closeChan: make(chan struct{}),
		config:    cfg,
	}

	// Create heartbeat monitor
	client.heartbeat = xml.NewHeartbeatMonitor(
		cfg.MOS.ID,
		clientID,
		cfg.MOS.ClientTimeout,
		cfg.MOS.HeartbeatInterval/2,
		client.Close,
	)

	return client
}

// Start starts processing for this client connection
func (c *ClientConnection) Start(ctx context.Context) {
	defer c.Close()

	// Start heartbeat monitoring
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()
	go c.heartbeat.Start(monitorCtx)

	// Subscribe to relevant events if event bus is available
	if c.server.eventBus != nil {
		roEvents := c.server.eventBus.Subscribe(events.RunningOrderUpdated, 10)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-c.closeChan:
					return
				case event, ok := <-roEvents:
					if !ok {
						return
					}
					// Send notification to this client
					c.handleRunningOrderUpdate(ctx, event)
				}
			}
		}()
	}

	// Create a Sentry span for this client connection
	span := sentry.StartSpan(ctx, "client_connection")
	span.SetTag("client_id", c.id)
	span.SetTag("remote_addr", c.conn.RemoteAddr().String())
	defer span.Finish()

	// Read loop
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closeChan:
			return
		default:
			// Set read deadline
			err := c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if err != nil {
				c.trackError(err, "set_read_deadline", nil)
				return
			}

			n, err := c.conn.Read(buffer)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					// Just a timeout, continue
					continue
				}
				if err == io.EOF {
					// Client closed connection
					logger.Infof("Client %s closed connection", c.id)
					return
				}
				c.trackError(err, "read", nil)
				return
			}

			// Process the data
			if n > 0 {
				c.parser.AppendData(buffer[:n])

				// Try to parse and handle complete messages
				for c.parser.HasCompleteMessage() {
					message, remaining, err := c.parser.Parse()
					if err != nil {
						if err == xml.ErrIncompleteXML {
							// Wait for more data
							break
						}
						// Get the current buffer content for the error details
						bufferContent := string(buffer[:n])
						c.trackError(err, "parse", map[string]interface{}{
							"data": bufferContent,
						})
						// Continue parsing, discard this message
						c.parser.Clear()
						c.parser.AppendData(remaining)
						continue
					}

					// Handle the message
					err = c.handleMessage(ctx, message)
					if err != nil {
						c.trackError(err, "handle_message", map[string]interface{}{
							"message_type": message.GetMessageType(),
						})
					}
				}
			}
		}
	}
}

// trackError captures an error with Sentry and returns it
func (c *ClientConnection) trackError(err error, operationType string, details map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Create tags and context for error tracking
	tags := map[string]string{
		"client_id":      c.id,
		"operation_type": operationType,
	}

	// Add extra context if provided
	if details == nil {
		details = make(map[string]interface{})
	}

	// Add client connection info to context
	details["remote_addr"] = c.conn.RemoteAddr().String()

	// Use a scope to capture error with all context
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(tags)
		scope.SetContext("client", details)
		scope.SetLevel(sentry.LevelError)
		sentry.CaptureException(err)
	})

	// Log locally as well
	logger.Errorf("[Client %s] %s error: %v", c.id, operationType, err)

	return err
}

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

// handleReqRunningOrderList processes a request for running order list
func (c *ClientConnection) handleReqRunningOrderList(ctx context.Context, req xml.ReqRunningOrderList) error {
	logger.Infof("Received running order list request from client %s", c.id)

	// Get running orders from the server
	runningOrders, err := c.server.service.ListRunningOrders(ctx)
	if err != nil {
		return c.sendErrorAck(req.RequestID, "ERROR", fmt.Sprintf("Failed to list running orders: %v", err))
	}

	// Convert to ROListItem
	items := make([]xml.ROListItem, 0, len(runningOrders))
	for _, ro := range runningOrders {
		items = append(items, xml.ROListItem{
			ID:       ro.ID,
			Slug:     ro.Slug,
			Channel:  ro.Channel,
			Status:   string(ro.Status),
			Duration: fmt.Sprintf("%d", ro.Duration),
		})
	}

	// Create response
	response := xml.CreateRunningOrderList(c.config.MOS.ID, req.RequestID, items)
	data, err := xml.GenerateMessage(response)
	if err != nil {
		return fmt.Errorf("failed to generate running order list response: %w", err)
	}

	return c.Write(data)
}

// handleReqRunningOrder processes a request for a specific running order
func (c *ClientConnection) handleReqRunningOrder(ctx context.Context, req xml.ReqRunningOrder) error {
	logger.Infof("Received running order request from client %s for RO %s", c.id, req.ROID)

	// Get the running order from the server
	ro, stories, err := c.server.service.GetRunningOrderWithStories(ctx, req.ROID)
	if err != nil {
		return c.sendErrorAck(req.RequestID, "ERROR", fmt.Sprintf("Failed to get running order: %v", err))
	}

	// Convert to StoryInfo
	storyInfos := make([]xml.StoryInfo, 0, len(stories))
	for _, story := range stories {
		// Get items for this story
		items, err := c.server.service.GetItemsForStory(ctx, story.ID)
		if err != nil {
			logger.Warningf("Failed to get items for story %s: %v", story.ID, err)
			continue
		}

		// Convert items
		itemInfos := make([]xml.ItemInfo, 0, len(items))
		for _, item := range items {
			itemInfos = append(itemInfos, xml.ItemInfo{
				ID:       item.ID,
				Slug:     item.Slug,
				Duration: fmt.Sprintf("%d", item.Duration),
				ObjectID: item.ObjectID,
			})
		}

		// Add story info
		storyInfos = append(storyInfos, xml.StoryInfo{
			ID:       story.ID,
			Slug:     story.Slug,
			Number:   story.Number,
			Duration: fmt.Sprintf("%d", story.Duration),
			Items:    itemInfos,
		})
	}

	// Create response
	response := xml.CreateRunningOrderInfo(
		c.config.MOS.ID,
		req.RequestID,
		ro.ID,
		ro.Slug,
		ro.Channel,
		"", // EditTime
		"", // StartTime
		fmt.Sprintf("%d", ro.Duration),
		storyInfos,
	)

	data, err := xml.GenerateMessage(response)
	if err != nil {
		return fmt.Errorf("failed to generate running order response: %w", err)
	}

	return c.Write(data)
}

// handleRunningOrderInfo processes a running order create/update message
func (c *ClientConnection) handleRunningOrderInfo(ctx context.Context, roInfo xml.RunningOrderInfo) error {
	logger.Infof("Received running order info from client %s for RO %s", c.id, roInfo.ID)

	// Process the running order creation/update
	err := c.server.service.ProcessRunningOrderInfo(ctx, roInfo)
	if err != nil {
		return c.sendErrorAck(roInfo.RequestID, "ERROR", fmt.Sprintf("Failed to process running order: %v", err))
	}

	// Send acknowledgment
	return c.sendSuccessAck(roInfo.RequestID, "Running order processed successfully")
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

// Write sends data to the client
func (c *ClientConnection) Write(data []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// Set write deadline
	err := c.conn.SetWriteDeadline(time.Now().Add(c.config.Server.WriteTimeout))
	if err != nil {
		return c.trackError(err, "set_write_deadline", nil)
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return c.trackError(err, "write", map[string]interface{}{
			"data_length": len(data),
		})
	}

	return nil
}

// Close closes the client connection
func (c *ClientConnection) Close() {
	c.closeOnce.Do(func() {
		logger.Infof("Closing connection for client %s", c.id)
		close(c.closeChan)

		// Track connection closure in Sentry
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("client_id", c.id)
			scope.SetLevel(sentry.LevelInfo)
			sentry.CaptureMessage(fmt.Sprintf("Client connection %s closed", c.id))
		})

		if err := c.conn.Close(); err != nil {
			c.trackError(err, "close", nil)
		}

		c.server.unregisterClient(c.id)
	})
}

// ID returns the client ID
func (c *ClientConnection) ID() string {
	return c.id
}

// handleRunningOrderUpdate sends a running order update notification to the client
func (c *ClientConnection) handleRunningOrderUpdate(ctx context.Context, event events.Event) {
	roID, ok := event.Payload.(string)
	if !ok {
		logger.Warningf("Invalid running order ID in event payload for client %s", c.id)
		return
	}

	logger.Infof("Sending running order update notification to client %s for RO %s", c.id, roID)

	// Get the updated running order from the service
	ro, stories, err := c.server.service.GetRunningOrderWithStories(ctx, roID)
	if err != nil {
		logger.Errorf("Failed to get running order %s for notification: %v", roID, err)
		return
	}

	// Convert stories to StoryInfo
	storyInfos := make([]xml.StoryInfo, 0, len(stories))
	for _, story := range stories {
		// Get items for this story
		items, err := c.server.service.GetItemsForStory(ctx, story.ID)
		if err != nil {
			logger.Warningf("Failed to get items for story %s: %v", story.ID, err)
			continue
		}

		// Convert items
		itemInfos := make([]xml.ItemInfo, 0, len(items))
		for _, item := range items {
			itemInfos = append(itemInfos, xml.ItemInfo{
				ID:       item.ID,
				Slug:     item.Slug,
				Duration: fmt.Sprintf("%d", item.Duration),
				ObjectID: item.ObjectID,
			})
		}

		storyInfos = append(storyInfos, xml.StoryInfo{
			ID:       story.ID,
			Slug:     story.Slug,
			Number:   story.Number,
			Duration: fmt.Sprintf("%d", story.Duration),
			Items:    itemInfos,
		})
	}

	// Create and send the running order update message
	response := xml.CreateRunningOrderInfo(
		c.config.MOS.ID,
		"", // No request ID for push notifications
		ro.ID,
		ro.Slug,
		ro.Channel,
		"",
		"",
		fmt.Sprintf("%d", ro.Duration),
		storyInfos,
	)

	data, err := xml.GenerateMessage(response)
	if err != nil {
		logger.Errorf("Failed to generate running order notification for client %s: %v", c.id, err)
		return
	}

	if err := c.Write(data); err != nil {
		logger.Errorf("Failed to send running order notification to client %s: %v", c.id, err)
	}
}
