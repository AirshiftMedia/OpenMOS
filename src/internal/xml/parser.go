package xml

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrInvalidXML     = errors.New("invalid XML format")
	ErrUnknownMessage = errors.New("unknown message type")
	ErrIncompleteXML  = errors.New("incomplete XML data")
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

	case "ncsReqStoryAction":
		var ncsReqStoryAction NCSReqStoryAction
		remaining, err := p.parseMessage(&ncsReqStoryAction)
		if err != nil {
			return nil, p.buffer, err
		}
		message = ncsReqStoryAction
		p.buffer = remaining

	default:
		return nil, p.buffer, fmt.Errorf("%w: %s", ErrUnknownMessage, messageType)
	}

	return message, p.buffer, nil
}
