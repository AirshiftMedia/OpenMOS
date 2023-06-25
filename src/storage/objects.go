package storage

import "encoding/xml"

// defining the standard objects as structs
// these will be turned into xml without the header according to MOS standard

type mosMsg struct {
	XMLName   xml.Name `xml:"mos"`
	mosID     string   `xml:"mosID"`
	ncsID     string   `xml:"ncsID"`
	messageID string   `xml:"messageID"`
}

type mosAck struct {
	XMLName           xml.Name `xml:"mosAck"`
	objID             string   `xml:"objID"`
	objRev            string   `xml:"objRev"`
	status            string   `xml:"status"`
	statusDescription string   `xml:"statusDescription"`
}

type heartBeat struct {
	xmlName xml.Name `xml:"heartbeat"`
	time    string   `xml:"time"`
}

func getNextMessageID(idString string) {

	idString = "1" // TODO

	return
}
