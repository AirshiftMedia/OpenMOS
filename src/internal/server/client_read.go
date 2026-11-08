package server

import (
	"context"

	"airshift/openmos/internal/xml"
)

// processData appends received bytes to the parser and handles every
// complete message in the buffer.
func (c *ClientConnection) processData(ctx context.Context, data []byte) {
	c.parser.AppendData(data)

	// Try to parse and handle complete messages
	for c.parser.HasCompleteMessage() {
		message, remaining, err := c.parser.Parse()
		if err != nil {
			if err == xml.ErrIncompleteXML {
				// Wait for more data
				break
			}
			// Get the current buffer content for the error details
			bufferContent := string(data)
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
