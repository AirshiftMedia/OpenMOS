// implements the main MOS API listener on port 443
// implementation as raw TCP API server according to MOS Protocol version 4.0
// https://mosprotocol.com/wp-content/MOS-Protocol-Documents/MOS-Protocol-Version-4.0.pdf
// server structure based on Vladimir Vivien's examples https://github.com/vladimirvivien
// TODO: implement observing by sentry.io if it somehow can utilize seelog
// TODO: Using hosted MongoDB as a backend
// TODO: Using Kafka as an event bus

package backend

import (
	"net"
	"strings"
	"time"
	// logger "github.com/cihub/seelog"
)

// process connections

func handleConnection(conn net.Conn) {

	defer conn.Close()

	// setting timestamp in MOS format (RFC3339 without timezone and Z separator)

	mosTimestamp := time.Now().Format("2006-01-02") + "T" + time.Now().Format("15:04:05")

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

func createMessage(mosType string) (msgString string) {
	msg := &mosMsg{mosID: "enter.your.mosid", ncsID: "enter.your.ncsid", messageID: "1"}

	return
}
