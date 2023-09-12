package models

import "encoding/xml"

//	"encoding/xml"
//	"time"

// defining the standard objects as structs
// these will be turned into xml without the header according to MOS standard

func testStruct() {

}

type mosMsg struct {
	XMLName   xml.Name `xml:"mos"`
	mosID     string   `xml:"mosID"`
	ncsID     string   `xml:"ncsID"`
	messageID string   `xml:"messageID"`
}

/*

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

func mosTimestamp() string {
	return time.Now().Format("2006-01-02") + "T" + time.Now().Format("15:04:05")
}

*/
