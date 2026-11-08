package xml

import (
	"bytes"
	"fmt"
)

// AppendData adds data to the parser's buffer
func (p *MessageParser) AppendData(data []byte) {
	p.buffer = append(p.buffer, data...)
}

// Clear clears the parser's buffer
func (p *MessageParser) Clear() {
	p.buffer = p.buffer[:0]
}

// HasCompleteMessage checks if the buffer contains a complete XML message
func (p *MessageParser) HasCompleteMessage() bool {
	// Check if we have an opening and closing tag
	if len(p.buffer) < 2 {
		return false
	}

	// Find the opening tag
	start := bytes.IndexByte(p.buffer, '<')
	if start == -1 {
		return false
	}

	// Extract the tag name
	nameEnd := bytes.IndexAny(p.buffer[start:], " \t\n\r/>")
	if nameEnd == -1 {
		return false
	}

	if start+nameEnd >= len(p.buffer) {
		return false
	}

	tagName := string(p.buffer[start+1 : start+nameEnd])

	// Check for self-closing tag like <heartbeat/>
	selfClosingEnd := bytes.Index(p.buffer[start:], []byte("/>"))
	if selfClosingEnd != -1 {
		return true
	}

	// Look for closing tag
	closingTag := fmt.Sprintf("</%s>", tagName)
	return bytes.Contains(p.buffer, []byte(closingTag))
}
