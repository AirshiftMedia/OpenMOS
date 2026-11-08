package server

import (
	"context"
	"fmt"

	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
)

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
