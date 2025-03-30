package xml

import (
	"encoding/xml"
	"time"
)

// MOSMessage is the base interface for all MOS messages
type MOSMessage interface {
	GetMessageType() string
}

// MosExternalMetadata represents external metadata in MOS messages
type MosExternalMetadata struct {
	XMLName    xml.Name `xml:"mosExternalMetadata"`
	MosScope   string   `xml:"mosScope,omitempty"`
	MosSchema  string   `xml:"mosSchema"`
	MosPayload string   `xml:"mosPayload"`
}

// Heartbeat represents a MOS heartbeat message
// Format: <heartbeat/>
// or <heartbeat timestamp="timestamp" source="source"/>
type Heartbeat struct {
	XMLName   xml.Name `xml:"heartbeat"`
	RequestID string   `xml:"requestID,attr,omitempty"`
	Timestamp string   `xml:"timestamp,attr,omitempty"`
	Source    string   `xml:"source,attr,omitempty"`
}

// GetMessageType returns the type of the message
func (h Heartbeat) GetMessageType() string {
	return "heartbeat"
}

// ReqRunningOrderList represents a request for running order list
// Format: <reqMachInfo/>
type ReqRunningOrderList struct {
	XMLName   xml.Name `xml:"roReq"`
	RequestID string   `xml:"requestID,attr,omitempty"`
	Timestamp string   `xml:"timestamp,attr,omitempty"`
	Source    string   `xml:"source,attr,omitempty"`
}

// GetMessageType returns the type of the message
func (r ReqRunningOrderList) GetMessageType() string {
	return "roReq"
}

// RunningOrderList represents a response with the list of running orders
type RunningOrderList struct {
	XMLName      xml.Name     `xml:"roList"`
	RequestID    string       `xml:"requestID,attr,omitempty"`
	Timestamp    string       `xml:"timestamp,attr,omitempty"`
	Source       string       `xml:"source,attr,omitempty"`
	RunningOrder []ROListItem `xml:"ro"`
}

// ROListItem represents a single running order in a list
type ROListItem struct {
	ID        string `xml:"roID"`
	Slug      string `xml:"roSlug"`
	Channel   string `xml:"roChannel,omitempty"`
	EditTime  string `xml:"roEdStart,omitempty"`
	StartTime string `xml:"roTrigger,omitempty"`
	Duration  string `xml:"roDur,omitempty"`
	Status    string `xml:"roStatus,omitempty"`
}

// GetMessageType returns the type of the message
func (r RunningOrderList) GetMessageType() string {
	return "roList"
}

// ReqRunningOrder represents a request for a specific running order
type ReqRunningOrder struct {
	XMLName   xml.Name `xml:"roReqAll"`
	RequestID string   `xml:"requestID,attr,omitempty"`
	Timestamp string   `xml:"timestamp,attr,omitempty"`
	Source    string   `xml:"source,attr,omitempty"`
	ROID      string   `xml:"roID"`
}

// GetMessageType returns the type of the message
func (r ReqRunningOrder) GetMessageType() string {
	return "roReqAll"
}

// RunningOrderInfo represents a full running order with stories and items
type RunningOrderInfo struct {
	XMLName   xml.Name    `xml:"roCreate"`
	RequestID string      `xml:"requestID,attr,omitempty"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Source    string      `xml:"source,attr,omitempty"`
	ID        string      `xml:"roID"`
	Slug      string      `xml:"roSlug"`
	Channel   string      `xml:"roChannel,omitempty"`
	EditTime  string      `xml:"roEdStart,omitempty"`
	StartTime string      `xml:"roTrigger,omitempty"`
	Duration  string      `xml:"roDur,omitempty"`
	Stories   []StoryInfo `xml:"story"`
}

// StoryInfo represents a story within a running order
type StoryInfo struct {
	ID       string     `xml:"storyID"`
	Slug     string     `xml:"storySlug"`
	Number   string     `xml:"storyNum,omitempty"`
	Duration string     `xml:"storyDur,omitempty"`
	Items    []ItemInfo `xml:"item,omitempty"`
}

// ItemInfo represents an item within a story
type ItemInfo struct {
	ID       string `xml:"itemID"`
	Slug     string `xml:"itemSlug"`
	Duration string `xml:"itemDur,omitempty"`
	ObjectID string `xml:"objID,omitempty"`
	MosID    string `xml:"mosID,omitempty"`
	ObjPath  string `xml:"objPath,omitempty"`
	Channel  string `xml:"itemChannel,omitempty"`
}

// GetMessageType returns the type of the message
func (r RunningOrderInfo) GetMessageType() string {
	return "roCreate"
}

// MOSAck represents a general acknowledgment message
type MOSAck struct {
	XMLName           xml.Name `xml:"mosAck"`
	RequestID         string   `xml:"requestID,attr,omitempty"`
	Timestamp         string   `xml:"timestamp,attr,omitempty"`
	Source            string   `xml:"source,attr,omitempty"`
	Status            string   `xml:"status"`
	StatusDescription string   `xml:"statusDescription,omitempty"`
}

// GetMessageType returns the type of the message
func (m MOSAck) GetMessageType() string {
	return "mosAck"
}

// NCSAck represents an acknowledgment from the MOS to the NCS
type NCSAck struct {
	XMLName           xml.Name `xml:"ncsAck"`
	Status            string   `xml:"status"`
	StatusDescription string   `xml:"statusDescription,omitempty"`
}

// GetMessageType returns the type of the message
func (m NCSAck) GetMessageType() string {
	return "ncsAck"
}

// Now returns the current timestamp in MOS format
func Now() string {
	return time.Now().Format(time.RFC3339)
}
