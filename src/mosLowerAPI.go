// implements the main MOS API listener lower port 10540 messages
// implementation as raw TCP API server according to MOS Protocol version 4.0
// https://mosprotocol.com/wp-content/MOS-Protocol-Documents/MOS-Protocol-Version-4.0.pdf
// server structure based on Vladimir Vivien's examples https://github.com/vladimirvivien
// TODO: implement observing by sentry.io if it somehow can utilize seelog
// TODO: Using hosted MongoDB as a backend
// TODO: Using Kafka as an event bus

package main

import (
	"encoding/xml"
	"fmt"
	"net"
	"os"
	"strings"
	// logger "github.com/cihub/seelog"
)

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
	xml.Name	xml.Name	`xml:"heartbeat"`
	time string		`xml:"time"`
}

func main() {

	// set logging

	// defer logger.Flush()

	// starting listener

	ln, err := net.Listen("tcp", ":10540")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer ln.Close()

	fmt.Println("Starting OpenMOS service: Now listening on port 10540")

	// the following handles incoming requests

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}
		fmt.Println("New connection from ", conn.RemoteAddr())

		// let goroutine handle the connection

		go handleConnection(conn)
	}

}

// process connections

func handleConnection(conn net.Conn) {
	defer conn.Close()

}

func parseCommand(cmdLine string) (cmd, param string) {
	parts := strings.Split(cmdLine, " ")
	if len(parts) != 2 {
		return "", ""
	}
	cmd = strings.TrimSpace(parts[0])
	param = strings.TrimSpace(parts[1])
	return
}
