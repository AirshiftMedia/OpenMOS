package server

import (
	"context"
	"fmt"

	"airshift/openmos/internal/events"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"
)

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
		items, err := c.server.service.GetItemsForStory(ctx, ro.ID, story.ID)
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
				MosID:    item.MosID,
			})
		}

		storyInfos = append(storyInfos, xml.StoryInfo{
			ID:     story.ID,
			Slug:   story.Slug,
			Number: story.Number,
			Items:  itemInfos,
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

// forwardEvents sends running order updates from the event bus to this
// client until the connection or context is closed.
func (c *ClientConnection) forwardEvents(ctx context.Context, roEvents <-chan events.Event) {
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
}
