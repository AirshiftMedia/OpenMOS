package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// detectMessageType determines the type of message in the buffer
func (p *MessageParser) detectMessageType() (string, error) {
	start := bytes.IndexByte(p.buffer, '<')
	if start == -1 {
		return "", ErrInvalidXML
	}

	nameEnd := bytes.IndexAny(p.buffer[start:], " \t\n\r/>")
	if nameEnd == -1 {
		return "", ErrInvalidXML
	}

	if start+nameEnd >= len(p.buffer) {
		return "", ErrInvalidXML
	}

	tagName := string(p.buffer[start+1 : start+nameEnd])
	return tagName, nil
}

// parseMessage parses the buffer into the given message type and returns the remaining data
func (p *MessageParser) parseMessage(message interface{}) ([]byte, error) {
	// Find the complete message
	messageType, err := p.detectMessageType()
	if err != nil {
		return p.buffer, err
	}

	// Find the end of the message
	var messageEnd int

	// Check for self-closing tag
	selfClosingEnd := bytes.Index(p.buffer, []byte("/>"))
	if selfClosingEnd != -1 && bytes.IndexAny(p.buffer[:selfClosingEnd], "<") == bytes.IndexByte(p.buffer, '<') {
		// This is a self-closing tag
		messageEnd = selfClosingEnd + 2
	} else {
		// Look for closing tag
		closingTag := fmt.Sprintf("</%s>", messageType)
		closingTagIndex := bytes.Index(p.buffer, []byte(closingTag))
		if closingTagIndex == -1 {
			return p.buffer, ErrIncompleteXML
		}
		messageEnd = closingTagIndex + len(closingTag)
	}

	// Parse the message
	if messageEnd > len(p.buffer) {
		return p.buffer, ErrIncompleteXML
	}

	err = xml.Unmarshal(p.buffer[:messageEnd], message)
	if err != nil {
		return p.buffer, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Return the remaining data
	if messageEnd >= len(p.buffer) {
		return []byte{}, nil
	}

	return p.buffer[messageEnd:], nil
}

// ParseMessage parses a complete XML string into a MOS message
func ParseMessage(xmlData string) (MOSMessage, error) {
	parser := NewMessageParser()
	parser.AppendData([]byte(xmlData))

	message, _, err := parser.Parse()
	return message, err
}

// ParseMessageFromReader parses an XML message from a reader
func ParseMessageFromReader(reader io.Reader) (MOSMessage, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return ParseMessage(string(data))
}
