package xml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Common errors
var (
	ErrInvalidXML      = errors.New("invalid XML format")
	ErrUnknownMessage  = errors.New("unknown message type")
	ErrIncompleteXML   = errors.New("incomplete XML data")
)

// MessageParser parses XML messages into their corresponding types
type MessageParser struct {
	buffer []byte
}

// NewMessageParser creates a new message parser
func NewMessageParser() *MessageParser {
	return &MessageParser{
		buffer: make([]byte, 0, 4096),
	}
}

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
	
	tagName := string(p.buffer[start+1 : start+nameEnd])
	
	// Check for self-closing tag like <heartbeat/>
	selfClosingEnd := bytes.Index(p.buffer[start:], []byte("/>"))
	if selfClosingEnd != -1 {
		return true
	}
	
	// Look for closing tag
	closingTag := fmt.Sprintf("</%s>", tagName)
	if bytes.Contains(p.buffer, []byte(closingTag)) {
		return true
	}
	
	return false
}

// Parse attempts to parse the buffer into a MOS message
func (p *MessageParser) Parse() (MOSMessage, []byte, error) {
	if !p.HasCompleteMessage() {
		return nil, p.buffer, ErrIncompleteXML
	}
	
	// Detect the message type based on the root element
	messageType, err := p.detectMessageType()
	if err != nil {
		return nil, p.buffer, err
	}
	
	var message MOSMessage
	
	// Parse based on message type
	switch messageType {
	case "heartbeat":
		var heartbeat Heartbeat
		remaining, err := p.parseMessage(&heartbeat)
		if err != nil {
			return nil, p.buffer, err
		}
		message = heartbeat
		p.buffer = remaining
		
	case "roReq":
		var roReq ReqRunningOrderList
		remaining, err := p.parseMessage(&roReq)
		if err != nil {
			return nil, p.buffer, err
		}
		message = roReq
		p.buffer = remaining
		
	case "roReqAll":
		var roReqAll ReqRunningOrder
		remaining, err := p.parseMessage(&roReqAll)
		if err != nil {
			return nil, p.buffer, err
		}
		message = roReqAll
		p.buffer = remaining
		
	case "roList":
		var roList RunningOrderList
		remaining, err := p.parseMessage(&roList)
		if err != nil {
			return nil, p.buffer, err
		}
		message = roList
		p.buffer = remaining
		
	case "roCreate":
		var roCreate RunningOrderInfo
		remaining, err := p.parseMessage(&roCreate)
		if err != nil {
			return nil, p.buffer, err
		}
		message = roCreate
		p.buffer = remaining
		
	case "mosAck":
		var mosAck MOSAck
		remaining, err := p.parseMessage(&mosAck)
		if err != nil {
			return nil, p.buffer, err
		}
		message = mosAck
		p.buffer = remaining
		
	default:
		return nil, p.buffer, fmt.Errorf("%w: %s", ErrUnknownMessage, messageType)
	}
	
	return message, p.buffer, nil
}

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