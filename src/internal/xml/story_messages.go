package xml

import (
	"encoding/xml"
)

// NCSReqStoryAction represents a request from NCS to perform an action on a story
type NCSReqStoryAction struct {
	XMLName     xml.Name    `xml:"ncsReqStoryAction"`
	Operation   string      `xml:"operation,attr"`
	LeaseLock   string      `xml:"leaseLock,attr,omitempty"`
	Username    string      `xml:"username,attr,omitempty"`
	ROStorySend ROStorySend `xml:"roStorySend"`
}

// GetMessageType returns the type of the message
func (a NCSReqStoryAction) GetMessageType() string {
	return "ncsReqStoryAction"
}

// ROStorySend represents a story send operation
type ROStorySend struct {
	XMLName      xml.Name              `xml:"roStorySend"`
	ROID         string                `xml:"roID"`
	StoryID      string                `xml:"storyID"`
	StorySlug    string                `xml:"storySlug,omitempty"`
	StoryNum     string                `xml:"storyNum,omitempty"`
	StoryBody    StoryBody             `xml:"storyBody"`
	ExternalMeta []MosExternalMetadata `xml:"mosExternalMetadata,omitempty"`
}

// StoryBody represents the body content of a story
type StoryBody struct {
	XMLName    xml.Name         `xml:"storyBody"`
	ReadAsBody string           `xml:"Read1stMEMasBody,attr,omitempty"`
	Paragraphs []StoryParagraph `xml:"p"`
}

// StoryParagraph represents a paragraph in a story body
type StoryParagraph struct {
	XMLName      xml.Name         `xml:"p"`
	Content      string           `xml:",chardata"` // Plain text content
	Instructions []StoryPI        `xml:"pi,omitempty"`
	Presenters   []StoryPresenter `xml:"storyPresenter,omitempty"`
	PresenterRRs []string         `xml:"storyPresenterRR,omitempty"`
	Items        []StoryItem      `xml:"storyItem,omitempty"`
	// Handle formatting tags like b, i, u if needed
}

// StoryPI represents producer instructions in a story paragraph
type StoryPI struct {
	XMLName xml.Name `xml:"pi"`
	Content string   `xml:",chardata"`
}

// StoryPresenter represents a presenter assignment in a story
type StoryPresenter struct {
	XMLName xml.Name `xml:"storyPresenter"`
	Name    string   `xml:",chardata"`
}

// StoryItem represents a media item within a story paragraph
type StoryItem struct {
	XMLName           xml.Name              `xml:"storyItem"`
	ItemID            string                `xml:"itemID"`
	ItemSlug          string                `xml:"itemSlug,omitempty"`
	ObjID             string                `xml:"objID"`
	MosID             string                `xml:"mosID"`
	ItemEdStart       int                   `xml:"itemEdStart,omitempty"`
	ItemEdDur         int                   `xml:"itemEdDur,omitempty"`
	ItemUserTimingDur int                   `xml:"itemUserTimingDur,omitempty"`
	MacroIn           string                `xml:"macroIn,omitempty"`
	MacroOut          string                `xml:"macroOut,omitempty"`
	ExternalMeta      []MosExternalMetadata `xml:"mosExternalMetadata,omitempty"`
}
